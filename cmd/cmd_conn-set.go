package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

type setConnArg struct {
	User     string `flag:"user"`
	Password string `flag:"pass"`
	Database string `flag:"db"`

	Key  string `arg:"key"`
	Addr string `arg:"addr"`
}

var setConnCmd = &cmdt.Command{
	Name: "conn-set",

	Pv: (*setConnArg)(nil),

	Usage: help.ConnSetUsage,
	Help:  help.ConnSet,

	Action: func(p interface{}) error {
		arg := p.(*setConnArg)

		if arg.Key == "" || arg.Addr == "" {
			return fmt.Errorf("key or addr is empty")
		}

		cfg := &conn.Config{
			Addr:     arg.Addr,
			Key:      arg.Key,
			User:     arg.User,
			Password: arg.Password,
			Database: arg.Database,
		}

		return conn.Set(arg.Key, cfg)
	},
}
