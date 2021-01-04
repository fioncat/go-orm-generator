package scango

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/build"
	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/iter"
)

// implements of scan go file.

// Flag represents a marked go element. It can be
// a function, interface, etc. A go element
// can be marked by multiple tags.
type Flag struct {
	// Line number of the go element.
	Line int `json:"line"`
	// arked tags
	Tags []Tag `json:"tags"`
}

// Tag represents a basic go-gendb tag. It is a
// special comment in go and will be recognized
// by go-gendb. Each time a tag is identified,
// a corresponding Tag structure is created.
//
// The format of Tag is: "// +gendb [args...]"
// There can be multiple args, separated by spaces.
//
// For example:
//    // +gendb mysql
//    // +gendb insert auto-ret
//
type Tag struct {
	// Line number of the tag.
	Line int `json:"line"`

	Args []string `json:"args"`
}

// Import represents a line of goimport statement.
// If no import name is specified,
// "filepath.Base(importPath)" will be used as the name.
type Import struct {
	// Line number of the import.
	Line int `json:"line"`

	Name string `json:"name"`
	Path string `json:"path"`
}

// Indent represents the go field marked by Tag.
// It is generally field in struct, variable or
// constant.
type Indent struct {
	Flag `json:"flag"`

	Name string `json:"name"`

	// TODO: At present, we will not parse the
	// type for indent. If the type is useful in
	// the future, additional fields are required.
}

// Method represents a function defined in go.
// The function can be defined externally or in an
// interface. The parsed function tokens will be
// stored in this structure to wait for further
// syntax analysis.
type Method struct {
	Flag `json:"flag"`

	// The original go statement defined the
	// method, is generally only used to display
	// and generate code.
	Origin string `json:"origin"`

	// All scanned tokens for the method.
	Tokens []token.Token `json:"-"`

	TokenStr string `json:"tokens"`
}

// Interface represents the interface defined in go,
// including names and multiple methods.
type Interface struct {
	Flag `json:"flag"`

	// The name for the interface
	// syntax: "type {name} interface {"
	Name string `json:"name"`

	// All methods for the interface
	Methods []Method `json:"methods"`
}

// Result is the result of scanning a go file, mainly
// concerned with the parts useful for code generation.
// For example, package definitions, package imports,
// and code modules marked with tags. For other parts
// that are not marked by tags and parts that are not
// useful for generating code, Result will not store
// these contents. So Result is a subset of all the
// information in the entire go file.
//
// During the code generation process, the Result will
// be passed to the parser to continue parsing to create
// an intermediate structure for code generation. So this
// structure is the first step in the compilation process.
type Result struct {
	// Path is the input go file's disk path
	Path string `json:"path"`

	// The scanned go file must contain a global definition
	// tag at the beginning. This field stores the globally
	// defined type. Later, different parsers will be selected
	// according to the type to parse the Result.
	// global tag syntax: "// +gendb {type}"
	Type string `json:"type"`

	// the package definition for go file. According to the
	// go syntax, all go files must contain this definition.
	// So in practice, this field cannot be empty.
	Package string `json:"package"`

	// All imported packages in the go file. When generating
	// code, it may be necessary to import the package of the
	// source file, so this field is needed.
	Imports []Import `json:"imports"`

	// Tagged indents and interfaces in go file. May be empty.
	Indents    []Indent    `json:"indents"`
	Interfaces []Interface `json:"interfaces"`
}

// Do scans the content of the go code file and converts it
// into a Result structure. It is actually a simple lexical
// analysis process.
//
// Some compilation errors can be found in the scanning phase
// (not all errors can be found), so all errors returned by
// this function are of type *errors.CompileError.
//
// Do is different from the Go compiler in that it only scans
// the lexical elements marked by tags and other elements that
// are needed for code generation (such as package, import, etc.)
func Do(sourcePath, content string) (*Result, error) {
	iter := iter.New(strings.Split(content, "\n"))
	hasImport := token.GO_IMPORT.Contains(content)

	result, err := header(sourcePath, iter, hasImport)
	if err != nil {
		return nil, err
	}
	result.Path = sourcePath

	err = body(sourcePath, iter, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// scan header
// The header contains: "master-tag", "package", "import"
// The first two is required.
func header(path string, iter *iter.Iter, hasImport bool) (*Result, error) {
	var r Result
	var line string
	for {
		idx := iter.NextP(&line)
		if idx < 0 {
			break
		}
		lineNum := idx + 1
		line = strings.TrimSpace(line)

		if isTag(line) {
			// master tag
			tag := _tag(line, lineNum)
			if len(tag.Args) == 0 {
				return nil, errors.NewComp(path,
					lineNum, "master tag miss type")
			}
			r.Type = tag.Args[0]
			continue
		}

		if token.GO_PACKAGE.Prefix(line) {
			if r.Type == "" {
				return nil, errors.NewComp(path,
					lineNum, "no master tag before package")
			}
			pkg := strings.TrimLeft(line, token.GO_PACKAGE.Get())
			pkg = strings.TrimSpace(pkg)
			r.Package = pkg
			if !hasImport {
				// if there is no import, the code body is
				// under the package definition, so there
				// is no need to continue the scan header.
				// we can EXIT EARLY here.
				break
			}
		}

		if token.GO_IMPORTS.Prefix(line) {
			// import (
			//   [name] "path"
			//   ...
			// )
			for {
				idx = iter.NextP(&line)
				if idx < 0 {
					break
				}
				if token.RPAREN.EqString(line) {
					break
				}
				imp := _import(line)
				if imp.Name == "" || imp.Path == "" {
					continue
				}
				imp.Line = idx + 1
				r.Imports = append(r.Imports, imp)
			}

			// In standard Go code(processed by go-fmt tool),
			// Only one imports statement will appear, after
			// imports, go header definition is over
			// So we break here
			// TODO: this is no work for no-standard go code
			//       so in no-standard situation, imports may
			//       be incomplete.
			//       DO WE NEED TO DEAL WITH IT???
			break
		}

		if token.GO_IMPORT.Prefix(line) {
			// import [name] "path"
			imp := _import(line)
			if imp.Name != "" && imp.Path != "" {
				imp.Line = lineNum
				r.Imports = append(r.Imports, imp)
			}

			// In standard Go code(processed by go-fmt tool),
			// single import situation will only appear once.
			// If there are multiple imports, go-fmt will
			// automatically aggregate them into import (...).
			// So as long as there is an import, the code body
			// is directly behind, and there is no need to
			// continue scanning.
			// The same as imports, For non-standard code,
			// for example, there are multiple import statements:
			//        import a
			//        import b
			//        import c
			// The "b" and "c" will be ignored.
			break
		}
	}

	return &r, nil
}

// scan code body
func body(path string, iter *iter.Iter, result *Result) error {
	var tags []Tag
	var line string
	for {
		idx := iter.NextP(&line)
		if idx < 0 {
			return nil
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lineNum := idx + 1

		if isTag(line) {
			tag := _tag(line, lineNum)
			tags = append(tags, tag)
			continue
		}

		if token.GO_INTERFACE.Contains(line) {
			if len(tags) == 0 {
				// This interface is not marked by tag(s)
				// ignore it
				continue
			}

			interName := _interfaceName(line)
			if interName == "" {
				return errors.NewComp(path, lineNum,
					"interface name is empty")
			}

			var inter Interface
			inter.Line = lineNum
			inter.Name = interName
			inter.Tags = tags
			tags = nil

			var method Method
			var err error
			// Scan all methods for interface
			for {
				idx = iter.NextP(&line)
				line = strings.TrimSpace(line)
				if idx < 0 || token.RBRACE.EqString(line) {
					break
				}
				lineNum = idx + 1
				if token.EMPTY.EqString(line) {
					continue
				}
				if isTag(line) {
					tags = append(tags, _tag(line, lineNum))
					continue
				}
				if token.GO_COMMENT.Prefix(line) {
					// ignore comment
					continue
				}

				method.Tokens, err = _method(path, lineNum, line)
				if err != nil {
					return err
				}

				method.Origin = line
				method.Tags = tags
				method.Line = lineNum

				if build.DEBUG {
					ts := fmt.Sprint(method.Tokens)
					if len(ts) > 2 {
						ts = ts[1 : len(ts)-1]
						method.TokenStr = ts
					}
				}

				inter.Methods = append(inter.Methods, method)
				tags = nil
			}

			tags = nil
			if len(inter.Methods) == 0 {
				// NOTE: now we do not allow empty tagged interface
				return errors.NewComp(path, lineNum,
					"empty interface")
			}

			result.Interfaces = append(result.Interfaces, inter)
			continue
		}

		if len(tags) > 0 && !token.GO_COMMENT.Prefix(line) {
			var indent Indent
			indent.Line = lineNum
			indent.Name = line
			indent.Tags = tags
			tags = nil

			result.Indents = append(result.Indents, indent)
			continue
		}
	}
}

func isTag(line string) bool {
	return token.TAG_PREFIX.Prefix(line)
}

// convert line into tag struct.
func _tag(line string, lineNum int) Tag {
	var tag Tag
	tag.Line = lineNum
	fields := strings.Fields(line)
	for _, field := range fields {
		if token.TAG_NAME.EqString(field) ||
			token.GO_COMMENT.EqString(field) ||
			token.EMPTY.EqString(field) {
			continue
		}
		tag.Args = append(tag.Args, field)
	}
	return tag
}

// convert line into import struct.
func _import(line string) Import {
	fields := strings.Fields(line)
	var imp Import
	for _, field := range fields {
		if token.EMPTY.EqString(field) ||
			token.GO_IMPORT.EqString(field) {
			continue
		}
		if len(field) <= 2 {
			continue
		}
		lastidx := len(field) - 1
		quo := token.QUO
		if quo.EqRune(rune(field[0])) && quo.EqRune(rune(field[lastidx])) {
			imp.Path = field[1:lastidx]
			if imp.Name == "" {
				imp.Name = filepath.Base(imp.Path)
			}
			return imp
		}
		imp.Name = field
	}
	return imp
}

// return the interface's name
func _interfaceName(line string) string {
	fields := strings.Fields(line)
	for _, field := range fields {
		switch field {
		case token.GO_TYPE.Get(),
			token.GO_INTERFACE.Get(),
			token.LBRACE.Get():
			continue
		default:
			return field
		}
	}
	return ""
}

// scan interface's method definition, convert it into tokens.
func _method(path string, lineNum int, line string) ([]token.Token, error) {
	iter := iter.New([]rune(line))
	bucket := token.NewBucket()

	var ch rune
	for iter.Next(&ch) {
		if token.SPACE.EqRune(ch) {
			bucket.Indent()
			continue
		}
		switch ch {
		case token.LPAREN.Rune():
			bucket.Key(token.LPAREN)

		case token.RPAREN.Rune():
			bucket.Key(token.RPAREN)

		case token.LBRACK.Rune():
			idx := iter.NextP(&ch)
			if idx < 0 || !token.RBRACK.EqRune(ch) {
				return nil, errors.NewComp(path, lineNum,
					"LBRACK's next char is not RBRACK").
					WithCharNum(idx + 1)
			}
			bucket.Key(token.BRACKS)

		case token.MUL.Rune():
			bucket.Key(token.MUL)

		case token.COMMA.Rune():
			bucket.Key(token.COMMA)

		case token.PERIOD.Rune():
			bucket.Key(token.PERIOD)

		default:
			bucket.Append(ch)
		}
	}

	bucket.Indent()
	return bucket.Get(), nil
}
