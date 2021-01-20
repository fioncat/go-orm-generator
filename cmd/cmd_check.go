package cmd

import (
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/database/tools/check"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var checkCmd = &cmdt.Command{
	Name: "check",
	Pv:   (*check.Arg)(nil),

	Usage: help.CheckUsage,
	Help:  help.Check,

	Action: func(p interface{}) error {
		return check.Do(p.(*check.Arg))
	},
}
