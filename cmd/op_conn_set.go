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
