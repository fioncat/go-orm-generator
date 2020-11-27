package cmd

import "github.com/fioncat/go-gendb/generator/dogen"

var genOp = &Operation{
	Name:     "gen",
	ParamPtr: (*dogen.Arg)(nil),

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*dogen.Arg)
		dogen.Prepare(arg)
		return dogen.One(arg.Path, arg.ConfPath)
	},
}
