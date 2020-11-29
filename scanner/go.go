package scanner

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/scanner/token"
)

type GoResult struct {
	Path string `json:"path"`
	Type string `json:"type"`

	Package string     `json:"package"`
	Imports []GoImport `json:"imports"`

	Indents    []GoIndent    `json:"indents"`
	Interfaces []GoInterface `json:"interfaces"`
}

type GoImport struct {
	Line int    `json:"line"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type GoTag struct {
	Line int      `json:"line"`
	Args []string `json:"args"`
}

type GoIndent struct {
	Line  int     `json:"line"`
	Tags  []GoTag `json:"tags"`
	Token string  `json:"token"`
}

type GoInterface struct {
	Line    int        `json:"line"`
	Name    string     `json:"name"`
	Tags    []GoTag    `json:"tags"`
	Methods []GoMethod `json:"methods"`
}

type GoMethod struct {
	Line     int           `json:"line"`
	Tags     []GoTag       `json:"tags"`
	Def      string        `json:"def"`
	TokenStr string        `json:"tokens"`
	Tokens   []token.Token `json:"-"`
}

var (
	ErrNoGendb = errors.New("file is not gendb(missing header definition)")

	errMissType = errors.New("missing type deinition")
	errPkgEmpty = errors.New("empty package name")
)

func GoFile(path string, debug bool) (*GoResult, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	return Go(path, content, debug)
}

func Go(path, content string, debug bool) (*GoResult, error) {
	hasImport := strings.Contains(content, "import")
	scanner := newLines(content)
	res, err := scanHeader(scanner, hasImport)
	if err != nil {
		return nil, errors.TraceFile(err, path)
	}
	res.Path = path

	err = scanBody(scanner, res, debug)
	if err != nil {
		return nil, errors.TraceFile(err, path)
	}

	return res, nil
}

func scanHeader(s *lineScanner, includeImport bool) (*GoResult, error) {
	var r GoResult
	for {
		line, num := s.next()
		if num == -1 {
			break
		}
		if line == "" {
			continue
		}
		if isGoTag(line) {
			// This is a master tag
			tag := scanTag(line, num)
			if len(tag.Args) == 0 {
				return nil, errors.Line(errMissType, num)
			}
			r.Type = tag.Args[0]
			continue
		}

		if strings.HasPrefix(line, "package") {
			if r.Type == "" {
				return nil, ErrNoGendb
			}
			pkg := strings.TrimSpace(
				strings.TrimLeft(line, "package"))
			if pkg == "" {
				return nil, errors.Line(errPkgEmpty, num)
			}
			r.Package = pkg
			if !includeImport {
				// If the Go code is without "import" deinition,
				// After package deinition, header is over
				break
			}
			continue
		}

		if strings.HasPrefix(line, "import (") {
			// multi-imports situation
			for {
				line, num := s.next()
				if num == -1 {
					break
				}
				if line == ")" {
					break
				}
				imp := scanImport(line)
				if imp.Name == "" || imp.Path == "" {
					continue
				}
				imp.Line = num
				r.Imports = append(r.Imports, imp)
			}
			// In standard Go code(processed by go-fmt tool),
			// Only one imports statement will appear, after
			// imports, go header definition is over
			// So we break here
			break
		}

		if strings.HasPrefix(line, "import") {
			// single-import
			imp := scanImport(line)
			if imp.Name != "" && imp.Path != "" {
				imp.Line = num
				r.Imports = append(r.Imports, imp)
			}
			// The same as imports ...
			break
		}
	}
	return &r, nil
}

func isGoTag(line string) bool {
	return strings.HasPrefix(line, "// +gendb")
}

func scanImport(line string) GoImport {
	s := newFields(line)
	var imp GoImport
	for {
		field, ok := s.next()
		if !ok {
			return imp
		}
		if field == "" || field == "import" {
			continue
		}
		if strings.HasPrefix(field, `"`) {
			imp.Path = strings.Trim(field, `"`)
			if imp.Name == "" {
				imp.Name = filepath.Base(imp.Path)
			}
			return imp
		}
		imp.Name = field
	}
}

func scanTag(line string, num int) GoTag {
	s := newFields(line)
	var tag GoTag
	tag.Line = num
	for {
		field, ok := s.next()
		if !ok {
			return tag
		}
		if field == "+gendb" || field == "//" || field == "" {
			continue
		}
		tag.Args = append(tag.Args, field)
	}
}

var (
	errEmptyInterface    = errors.New("empty interface name")
	errInterfaceNoMethod = errors.New("interface has no method")
)

func scanBody(s *lineScanner, r *GoResult, debug bool) error {
	var tags []GoTag
	for {
		line, num := s.next()
		if num == -1 {
			return nil
		}
		if line == "" {
			continue
		}

		if isGoTag(line) {
			tag := scanTag(line, num)
			tags = append(tags, tag)
			continue
		}

		// Go interface definition
		if strings.Contains(line, "interface") {
			if len(tags) == 0 {
				// This interface is not marked by
				// the "gendb" tag, so ignore it.
				continue
			}

			// Try to parse interface name.
			name := parseInterfaceName(line)
			if name == "" {
				return errors.Line(errEmptyInterface, num)
			}
			var inter GoInterface
			inter.Line = num
			inter.Name = name
			inter.Tags = tags
			tags = nil

			// Find all methods for this interface
			for {
				line, num := s.next()
				if num == -1 || line == "}" {
					break
				}
				if line == "" {
					continue
				}
				if isGoTag(line) {
					tags = append(tags, scanTag(line, num))
					continue
				}
				if strings.HasPrefix(line, "//") {
					continue
				}

				var method GoMethod
				var err error
				method.Tokens, err = scanGoMethod(line)
				if err != nil {
					return errors.Line(err, num)
				}
				method.Def = line
				method.Tags = tags
				method.Line = num
				if debug {
					method.TokenStr = token.Join(method.Tokens)
				}

				inter.Methods = append(inter.Methods, method)
				tags = nil
			}

			tags = nil
			if len(inter.Methods) == 0 {
				// Now we do not allow empty interface
				return errors.Line(
					errInterfaceNoMethod, inter.Line)
			}

			r.Interfaces = append(r.Interfaces, inter)
			continue
		}

		if len(tags) > 0 && !strings.HasPrefix(line, "//") {
			var indent GoIndent
			indent.Line = num
			indent.Token = line
			indent.Tags = tags
			tags = nil

			r.Indents = append(r.Indents, indent)
			continue
		}
	}
}

func parseInterfaceName(line string) string {
	s := newFields(line)
	for {
		field, ok := s.next()
		if !ok {
			return ""
		}
		switch field {
		case "type", "interface", "{":
			continue
		default:
			return field
		}
	}
}

var (
	errSliceFmt = errors.New("slice bad format, need '[]'")
)

func scanGoMethod(def string) ([]token.Token, error) {
	scanner := newChars(def)

	var tokens []token.Token
	var bucket []rune
	addBucket := func() {
		if len(bucket) == 0 {
			return
		}
		s := string(bucket)
		tokens = append(tokens, token.NewIndent(s))
		bucket = nil
	}
	addFlag := func(s string) {
		addBucket()
		tokens = append(tokens, token.NewFlag(s))
	}
	for {
		ch, ok := scanner.next()
		if !ok {
			addBucket()
			break
		}

		if ch == ' ' {
			addBucket()
			continue
		}

		switch ch {
		case '(':
			addFlag("(")

		case ')':
			addFlag(")")

		case '[':
			next, ok := scanner.next()
			if !ok || next != ']' {
				return nil, errSliceFmt
			}
			addFlag("[]")

		case '*':
			addFlag("*")

		case ',':
			addFlag(",")

		case '.':
			addFlag(".")

		default:
			bucket = append(bucket, ch)

		}
	}

	return tokens, nil

}
