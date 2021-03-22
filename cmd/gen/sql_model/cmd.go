package sql_model

import (
	"github.com/fioncat/go-gendb/generate"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var Cmder = &cmdt.Command{
	Name: "gen-sql-model",
	Pv:   (*generate.SqlModelArg)(nil),

	Action: func(p interface{}) error {
		return generate.SqlModel(p.(*generate.SqlModelArg))
	},
}
