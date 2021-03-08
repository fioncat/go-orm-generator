package csql

import (
	"github.com/fioncat/go-gendb/compile/sql"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type Arg struct {
	Path string `arg:"path"`
}

var Cmder = &cmdt.Command{
	Name: "csql",
	Pv:   (*Arg)(nil),

	Usage: "csql <path>",
	Help:  desc,

	Action: func(p interface{}) error {
		arg := p.(*Arg)
		file, err := sql.ReadFile(arg.Path)
		if err != nil {
			return err
		}
		term.Show(file)
		return nil
	},
}

const desc = `
Csql executes the compilation process on a sql file and
outputs the result in the form of JSON.

The input file must meet the go-gendb specification.

See also: cgo`
