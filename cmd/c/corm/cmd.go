package corm

import (
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/orm"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type Arg struct {
	Conn string `arg:"conn"`
	Path string `arg:"path"`
}

var Cmder = &cmdt.Command{
	Name: "corm",
	Pv:   (*Arg)(nil),

	Action: func(p interface{}) error {
		arg := p.(*Arg)
		path := arg.Path

		if err := rdb.Init(arg.Conn, "mysql"); err != nil {
			return err
		}

		gfile, err := golang.ReadFile(path)
		if err != nil {
			return err
		}

		file, err := orm.Parse(gfile)
		if err != nil {
			return err
		}

		term.Show(file)
		return nil
	},
}
