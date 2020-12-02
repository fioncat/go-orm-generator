package cmd

import "github.com/fioncat/go-gendb/dbaccess/dbcheck"

var checkSQLOp = &Operation{
	Name:     "check-sql",
	ParamPtr: (*dbcheck.Arg)(nil),

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*dbcheck.Arg)
		return dbcheck.Run(arg)
	},
}
