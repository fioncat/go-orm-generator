package cmd

import (
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

type ParseArg struct {
	Path string `arg:"path"`
}

var parseCmd = &cmdt.Command{
	Name: "parse",
	Pv:   (*ParseArg)(nil),

	Usage: help.ParseUsage,
	Help:  help.Parse,

	Action: func(p interface{}) error {
		return nil
	},
}
