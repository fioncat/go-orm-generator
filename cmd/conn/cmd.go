package conn

import (
	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type Arg struct {
	Add bool `flag:"u"`
	Del bool `flag:"d"`

	Key string `arg:"key"`
}

var Cmder = &cmdt.Command{
	Name: "conn",
	Pv:   (*Arg)(nil),

	Usage: "conn [-u|-d|-a] <key>",
	Help:  desc,

	Action: func(p interface{}) error {
		arg := p.(*Arg)
		switch {
		case arg.Add:
			cfg := inputCfg()
			return conn.Set(arg.Key, cfg)

		case arg.Del:
			return conn.Remove(arg.Key)

		default:
			cfg, err := conn.Get(arg.Key)
			if err != nil {
				return err
			}
			term.Show(cfg)
		}
		return nil
	},
}

func inputCfg() *conn.Config {
	cfg := new(conn.Config)
	cfg.Addr = term.Input("Please input address")
	cfg.User = term.Input("Please input user")
	cfg.Password = term.Input("Please input password")
	cfg.Database = term.Input("Please input database")

	return cfg
}

const desc = `
Conn is used to configure or view the database connections
which may be used in code generation.

Database connections will be stored on the local disk, and
they will be stored in "~/.go_gendb_store/conn".

Using different flags will perform different operations.
If there is no flag, the database connection will be displayed
directly in the form of JSON.

Command Flags:
    <key>
         Connection key. Used to uniquely identify a connection.
    -u
         Upsert mode.
    -a
         Add mode.
    -d
         Delete mode.

See alse: gen`
