package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/store"
)

var connDelOp = &Operation{
	Name:     "conn-del",
	ParamPtr: (*connArg)(nil),

	Usage: connDelUsage,
	Help:  connDelHelp,

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

const connDelUsage = `conn-del <key>`

const connDelHelp = `
Delete connection by the key.

See alse: conn-set, conn-get`
