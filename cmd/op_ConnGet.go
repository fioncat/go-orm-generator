package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/store"
)

type connArg struct {
	LogPath   string `flag:"log-file"`
	EnableLog bool   `flag:"enable-log"`

	Key string `arg:"key"`
}

var connGetOp = &Operation{
	Name:     "conn-get",
	ParamPtr: (*connArg)(nil),

	Usage: connGetUsage,
	Help:  connGetHelp,

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*connArg)
		if arg.EnableLog {
			log.Init(true, arg.LogPath)
		}
		conn, err := store.GetConn(arg.Key)
		if err != nil {
			fmt.Println(err)
			return false
		}
		store.ShowConn(conn)
		return true
	},
}

const connGetUsage = `conn-get <key>`

const connGetHelp = `
Show the database connection in json-form.

Specify the connection to be showed through
the connection key.

For example, to show the "test" connection can use:

go-gendb conn-get test

See also: conn-set, conn-del`
