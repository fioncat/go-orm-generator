package exec

import (
	"github.com/fioncat/go-gendb/database/tools/exec"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var Cmder = &cmdt.Command{
	Name: "exec",
	Pv:   (*exec.Arg)(nil),

	Action: func(p interface{}) error {
		return exec.Do(p.(*exec.Arg))
	},
}
