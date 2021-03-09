package check

import (
	"github.com/fioncat/go-gendb/database/tools/check"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var Cmder = &cmdt.Command{
	Name: "check",
	Pv:   (*check.Arg)(nil),

	Action: func(p interface{}) error {
		return check.Do(p.(*check.Arg))
	},
}
