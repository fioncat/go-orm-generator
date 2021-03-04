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
