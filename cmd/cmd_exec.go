package cmd

import (
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/database/tools/exec"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var execCmd = &cmdt.Command{
	Name: "exec",
	Pv:   (*exec.Arg)(nil),

	Usage: help.ExecUsage,
	Help:  help.Exec,

	Action: func(p interface{}) error {
		return exec.Do(p.(*exec.Arg))
	},
}
