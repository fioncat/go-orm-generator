package convert

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/misc/errors"
)

type Result struct {
	Name   string
	Method string

	Target string

	Imports []string

	Fields []*Field
}

type Field struct {
	Left  string
	Right string

	ignore bool
	name   string
}

var structDecode = base.DecodeOptionMap{
	"name": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		r.Name = val
		return nil
	},

	"to": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		r.Target = val
		return nil
	},

	"method": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		r.Method = val
		return nil
	},

	"import": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		r.Imports = append(r.Imports, val)
		return nil
	},
}

var fieldDecode = base.DecodeOptionMap{
	"ignore": func(line int, val string, vs []interface{}) error {
		f := vs[0].(*Field)
		f.ignore = true
		return nil
	},

	"map": func(line int, val string, vs []interface{}) error {
		f := vs[0].(*Field)
		f.Left = "a." + val
		return nil
	},

	"set": func(line int, val string, vs []interface{}) error {
		f := vs[0].(*Field)
		f.Right = base.ReplacePlaceholder(val, "b."+f.name, "b", f.name)
		return nil
	},
}

func Parse(stc *golang.Struct, opts []base.Option) (*Result, error) {
	r := new(Result)

	err := base.DecodeOptions(opts, structDecode, r)
	if err != nil {
		return nil, err
	}

	if r.Name == "" {
		r.Name = stc.Name
	}
	if r.Method == "" {
		r.Method = "Convert"
	}
	if r.Target == "" {
		return nil, errors.TraceFmt(stc.Line,
			`convert: missing "to" option.`)
	}

	for _, gf := range stc.Fields {
		if !coder.IsExport(gf.Name) {
			continue
		}
		flag, err := golang.ParseTypeFlag(gf.Type)
		if err != nil {
			return nil, errors.TraceFmt(gf.Line,
				`go type "%s" is bad format`, gf.Type)
		}
		if flag.Map || flag.Slice {
			continue
		}
		f := new(Field)
		f.name = gf.Name
		f.Left = gf.Name
		if flag.Simple {
			f.Right = gf.Name
		} else {
			f.Right = gf.Name + ".Convert()"
		}

		f.Left = "a." + f.Left
		f.Right = "b." + f.Right

		err = base.DecodeTags(gf.Tags,
			"convert", fieldDecode, f)
		if err != nil {
			return nil, err
		}
		if f.ignore {
			continue
		}
		r.Fields = append(r.Fields, f)
	}
	return r, nil
}
