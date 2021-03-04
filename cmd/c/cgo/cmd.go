package cgo

import (
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type Arg struct {
	Path string `arg:"path"`
}

var Cmder = &cmdt.Command{
	Name: "cgo",
	Pv:   (*Arg)(nil),

	Action: func(p interface{}) error {
		arg := p.(*Arg)
		file, err := golang.ReadFile(arg.Path)
		if err != nil {
			return err
		}
		term.Show(file)
		return nil
	},
}
