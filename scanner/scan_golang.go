package scanner

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/parser"
)

var (
	goMissTag   = errors.New(`The file is missing the gendb tag. Please add the comment tag "// +gendb {dbType}" before "package".`)
	goMissType  = errors.New(`The gendb tag missing the type definition, it must be "// +gendb {dbType}"`)
	goMissPkg   = errors.New(`The file is missing "package" definition.`)
	goPkgBadFmt = errors.New(`package definition bad format`)
	goImpBadFmt = errors.New(`import definition bad format`)
)

func Golang(path string) (*generator.Task, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	// From the beginning of the file to
	// the definition of "package",
	// there must be a "+gendb" tag.
	var dbType string
	var pkg string
	pkgIdx := 0
	for i, line := range lines {
		line = parser.TS(line)
		if opt := parser.GoTag(line); opt != nil {
			if len(opt.Args) == 0 {
				return nil, errors.Line(path, i+1, goMissType)
			}
			dbType = opt.Args[0]
			continue
		}
		if parser.HasL(line, "package") {
			if dbType == "" {
				return nil, errors.Trace(path, goMissTag)
			}
			tmp := parser.FD(line)
			if len(tmp) != 2 {
				return nil, errors.Line(path, i+1, goPkgBadFmt)
			}
			pkg = tmp[1]
			pkgIdx = i
			break
		}
	}
	if pkg == "" {
		return nil, errors.Trace(path, goMissPkg)
	}
	log.Info("file:", path, "lines:", len(lines))
	log.Info("package", pkg)
	log.Info("dbType", dbType)

	t := new(generator.Task)
	if pkgIdx+1 >= len(lines) {
		// Nothing else except the package
		return t, nil
	}
	t.Type = dbType
	t.Path = path
	t.Pkg = pkg

	err = parseGolang(path, t, lines, pkgIdx+1)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func parseGolang(path string, t *generator.Task, lines []string, start int) error {
	idx := start
	for idx < len(lines) {
		line := parser.TS(lines[idx])
		if parser.HasL(line, "import (") {
			start := idx + 1
			stopped := false
			for j := start; j < len(lines); j++ {
				line := parser.TS(lines[j])
				if line == ")" {
					idx = j
					stopped = true
					break
				}
				imp := parser.GoImport(line)
				if imp == nil {
					return errors.Line(path, j+1, goImpBadFmt)
				}
				log.Infof(" > import %s", imp.String())
			}
			if !stopped {
				return errors.New("missing import terminator")
			}
			continue
		}
		if parser.HasL(line, "import") {
			imp := parser.GoImport(line)
			if imp == nil {
				return errors.Line(path, idx+1, goImpBadFmt)
			}
			t.Imports = append(t.Imports, *imp)
			log.Infof(" > import %s", imp.String())
		}
		if tag := parser.GoTag(line); tag != nil {
			var endI int
			var err error
			switch tag.Tag {
			case "struct":
				endI, err = parseGolangStruct(path, t, tag, lines, idx+1)
			case "inter":
				endI, err = parseGolangInter(path, t, tag, lines, idx+1)
			default:
				err = errors.Fmt("unknown tag %s", tag.Tag)
				err = errors.Line(path, idx+1, err)
			}
			if err != nil {
				return err
			}
			idx = endI
			continue
		}
		idx += 1
	}
	return nil
}

func parseGolangStruct(path string, t *generator.Task, opt *generator.TaskOption, lines []string, start int) (int, error) {
	var s generator.TaskStruct
	s.Args = opt.Args
	end := 0
	for i := start; i < len(lines); i++ {
		line := parser.TS(lines[i])
		if line == "" {
			continue
		}
		if parser.GoComment(line) {
			if opt := parser.GoTag(line); opt != nil {
				opt.Line = i + 1
				s.Options = append(s.Options, *opt)
			}
			continue
		}
		s.Name = parser.GoStruct(line)
		if !isNameValid(s.Name) {
			err := errors.Fmt(`bad struct name "%s"`, s.Name)
			return 0, errors.Line(path, i+1, err)
		}
		end = i
		break
	}
	if s.Name == "" {
		err := errors.New("missing struct definition")
		return 0, errors.Line(path, start+1, err)
	}
	t.Structs = append(t.Structs, s)
	log.Infof("add struct: %s, args=[%s], options=%s", s.Name,
		strings.Join(s.Args, ","),
		strings.Join(s.OptionNames(), ";"))
	return end + 1, nil
}

func parseGolangInter(path string, t *generator.Task, opt *generator.TaskOption, lines []string, start int) (int, error) {
	var inter generator.TaskInterface
	inter.Args = opt.Args
	end := 0
	for i := start; i < len(lines); i++ {
		line := parser.TS(lines[i])
		if line == "" {
			continue
		}
		if parser.GoComment(line) {
			if opt := parser.GoTag(line); opt != nil {
				opt.Line = i + 1
				inter.Options = append(inter.Options, *opt)
			}
			continue
		}
		inter.Name = parser.GoInter(line)
		if inter.Name == "" {
			err := errors.New("interface bad format")
			return 0, errors.Line(path, i+1, err)
		}
		start = i + 1
		break
	}
	log.Infof("add interface: %s, args=[%s] options=%s",
		inter.Name, strings.Join(inter.Args, ","),
		inter.OptionNames())

	var opts []generator.TaskOption
	for i := start; i < len(lines); i++ {
		line := parser.TS(lines[i])
		if line == "" {
			continue
		}
		if line == "}" {
			end = i
			break
		}
		if parser.GoComment(line) {
			if opt := parser.GoTag(line); opt != nil {
				opt.Line = i + 1
				opts = append(opts, *opt)
			}
			continue
		}
		method, err := parser.MustGoMethod(line)
		if err != nil {
			return 0, errors.Line(path, i+1, err)
		}
		method.Options = opts
		opts = nil
		log.Infof("  # %s: %s, options=%s, pkgs=[%s]",
			inter.Name, method.String(), method.OptionNames(),
			strings.Join(method.Pkgs.Slice(), ","))
		inter.Methods = append(inter.Methods, *method)
	}

	if end == 0 {
		err := errors.Fmt("missing terminator "+
			"for interface %s", inter.Name)
		return 0, errors.Line(path, start, err)
	}
	t.Interfaces = append(t.Interfaces, inter)
	return end + 1, nil
}

var nameRe = regexp.MustCompile(`\w+`)

func isNameValid(name string) bool {
	return nameRe.MatchString(name)
}
