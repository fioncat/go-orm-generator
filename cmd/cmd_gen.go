package cmd

import (
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/generate"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var genCmd = &cmdt.Command{
	Name: "gen",
	Pv:   (*generate.Arg)(nil),

	Usage: help.GenUsage,
	Help:  help.Gen,

	Action: func(p interface{}) error {
		return generate.One(p.(*generate.Arg))
	},
}
