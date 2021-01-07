package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type getConnArg struct {
	Key string `arg:"key"`
}

var getConnCmd = &cmdt.Command{
	Name: "conn-get",

	Pv: (*getConnArg)(nil),

	Usage: help.ConnGetUsage,
	Help:  help.ConnGet,

	Action: func(p interface{}) error {
		key := p.(*getConnArg).Key
		if key == "" {
			return fmt.Errorf("key is empty")
		}

		cfg, err := conn.Get(key)
		if err != nil {
			return err
		}

		term.Show(cfg)
		return nil
	},
}
