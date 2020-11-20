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
	Name:     "get-conn",
	ParamPtr: (*connArg)(nil),
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
