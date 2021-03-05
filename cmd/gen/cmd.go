package gen

import (
	"github.com/fioncat/go-gendb/generate"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var Cmder = &cmdt.Command{
	Name: "gen",
	Pv:   (*generate.Arg)(nil),

	Action: func(p interface{}) error {
		return generate.Do(p.(*generate.Arg))
	},
}
