package deepcopy

import (
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/golang"
)

type Result struct {
	Name   string
	Method string

	Target string

	Imports []string

	Fields []*Field
}

type Field struct {
	Left string

	Name string
	Type *golang.DeepType

	Set string

	ignore bool
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
		f.Left = val
		return nil
	},

	"set": func(line int, val string, vs []interface{}) error {
		f := vs[0].(*Field)
		f.Set = val
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

	for _, gf := range stc.Fields {
		f := new(Field)
		f.Name = gf.Name
		err := base.DecodeTags(gf.Tags, "deepcopy", fieldDecode, f)
		if err != nil {
			return nil, err
		}
		if f.ignore {
			continue
		}
		f.Type, err = golang.ParseDeepType(gf.Type)
		if err != nil {
			return nil, err
		}
		r.Fields = append(r.Fields, f)
	}
	return r, nil
}
