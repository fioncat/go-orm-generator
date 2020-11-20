package parser

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/misc/col"
	"github.com/fioncat/go-gendb/misc/errors"
)

func GoImport(line string) *generator.Import {
	line = TL(line, "import")
	fs := FD(line)
	var (
		path string
		name string
	)
	switch len(fs) {
	case 1:
		path = T(fs[0], `"`)
		name = filepath.Base(path)
	case 2:
		path = T(fs[1], `"`)
		name = fs[0]
	default:
		return nil
	}
	return &generator.Import{
		Path: path,
		Name: name,
	}
}

func GoTag(line string) *generator.TaskOption {
	if !HasL(line, "//") {
		return nil
	}
	line = TS(TL(line, "//"))
	if !HasL(line, "+gendb") {
		return nil
	}

	fs := FD(line)
	if len(fs) == 0 {
		return nil
	}
	opt := new(generator.TaskOption)
	if len(fs) > 1 {
		opt.Args = fs[1:]
	}
	namef := fs[0]
	namef = TL(namef, "+gendb")
	namef = TL(namef, "-")
	opt.Tag = namef

	return opt
}

func GoComment(line string) bool {
	return HasL(line, "//")
}

func GoInter(line string) string {
	fs := FD(line)
	if len(fs) != 4 {
		return ""
	}
	if fs[0] != "type" {
		return ""
	}
	if fs[2] != "interface" {
		return ""
	}
	if fs[3] != "{" {
		return ""
	}
	return fs[1]
}

func GoStruct(line string) string {
	line = TL(line, "*")
	return line
}

var (
	goMethodRe = regexp.MustCompile(`(\w+)\((.*)\) \((.+), error\)`)

	goMethodMatchErr = errors.Fmt(`function bad format, it must be: "{name}({param}) ({ret}, error)"`)
)

func MustGoMethod(line string) (*generator.TaskMethod, error) {
	matches := goMethodRe.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil, goMethodMatchErr
	}
	name := matches[1]
	params := matches[2]
	ret := matches[3]

	m := new(generator.TaskMethod)
	m.Name = name
	m.ReturnType = ret
	m.Params = params

	err := parseGoMethodParam(m, params)
	if err != nil {
		return nil, err
	}

	if strings.Contains(m.ReturnType, ".") {
		pkg, err := goParsePkgRef(m.ReturnType)
		if err != nil {
			return nil, err
		}
		m.Pkgs.Add(pkg)
	}

	return m, nil
}

var goMethodParamIndentRe = regexp.MustCompile(`(\w+)\.\w+`)

func parseGoMethodParam(m *generator.TaskMethod, param string) error {
	fields := SP(param, ",")
	m.ParamNames = col.NewSet(len(fields))
	m.Pkgs = col.NewSet(0)

	for _, field := range fields {
		if field == "" {
			continue
		}
		fieldTmp := FD(field)
		if len(fieldTmp) != 2 {
			return errors.Fmt(`bad param field definition "%s", `+
				`it must be "{name} {type}"`, field)
		}
		name := fieldTmp[0]
		fieldType := fieldTmp[1]

		if m.ParamNames.Exists(name) {
			return errors.Fmt(`param "%s" is duplcate`, name)
		}
		m.ParamNames.Add(name)

		// If there has a definition of "pkg.indent" in the
		// type. It means that this functionc refers to
		// another package
		if strings.Contains(fieldType, ".") {
			// Resolve the referenced package name
			pkg, err := goParsePkgRef(fieldType)
			if err != nil {
				return err
			}
			m.Pkgs.Add(pkg)
		}
	}
	return nil
}

func goParsePkgRef(ref string) (string, error) {
	matches := goMethodParamIndentRe.FindStringSubmatch(ref)
	if len(matches) != 2 {
		return "", errors.Fmt(`Bad package reference: "%s", `+
			`is must be "{pkg}.{Type}"`, ref)
	}
	pkg := matches[1]
	return pkg, nil
}
