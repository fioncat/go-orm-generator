package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var delConnCmd = &cmdt.Command{
	Name: "conn-del",

	Pv: (*getConnArg)(nil),

	Usage: help.ConnDelUsage,
	Help:  help.ConnDel,

	Action: func(p interface{}) error {
		key := p.(*getConnArg).Key
		if key == "" {
			return fmt.Errorf("key is empty")
		}

		return conn.Remove(key)
	},
}
