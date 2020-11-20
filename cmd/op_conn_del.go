package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/store"
)

var connDelOp = &Operation{
	Name:     "del-conn",
	ParamPtr: (*connArg)(nil),
	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*connArg)
		if arg.EnableLog {
			log.Init(true, arg.LogPath)
		}
		err := store.DelConn(arg.Key)
		if err != nil {
			fmt.Println(err)
			return false
		}
		return true
	},
}
