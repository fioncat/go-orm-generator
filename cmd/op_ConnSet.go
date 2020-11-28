package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/store"
)

type connSetArg struct {
	LogPath   string `flag:"log-file"`
	EnableLog bool   `flag:"enable-log"`
}

var connSetOp = &Operation{
	Name:     "conn-set",
	ParamPtr: (*connSetArg)(nil),

	Usage: connSetUsage,
	Help:  connSetHelp,

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*connSetArg)
		if arg.EnableLog {
			log.Init(true, arg.LogPath)
		}

		conn, err := store.InputConn()
		if err != nil {
			fmt.Println(err)
			return false
		}
		err = store.SaveConn(conn)
		if err != nil {
			fmt.Println(err)
			return false
		}
		return true
	},
}

const connSetUsage = `go-gendb conn-set`

const connSetHelp = `
Set the database connection.

In some cases, the generated code needs to read
database data, and then the database connection
is used, which is configured through this command.

Use the "--conn-key" option in "go-gendb gen" to
introduce the configured connection.

A connection requires the following configuration:

key(required):  The index key of the connection.
                "gen", "conn-get", "conn-del" use this
                key to find the connection.
addr(required): The database address.
user:           The database username.
password:       The database password.
database:       The database name.

See alse: gen, conn-get, conn-del`
