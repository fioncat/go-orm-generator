package cmd

import (
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/generate"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var batchCmd = &cmdt.Command{
	Name: "batch",
	Pv:   (*generate.Arg)(nil),

	Usage: help.BatchUsage,
	Help:  help.Batch,

	Action: func(p interface{}) error {
		return generate.Batch(p.(*generate.Arg))
	},
}
