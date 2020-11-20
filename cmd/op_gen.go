package cmd

import (
	"fmt"
	"time"

	"github.com/fioncat/go-gendb/dbaccess"
	"github.com/fioncat/go-gendb/generator/genfile"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/store"
)

type genArg struct {
	LogPath     string `flag:"log-file"`
	EnableLog   bool   `flag:"enable-log"`
	EnableCache bool   `flag:"enable-cache"`
	CacheTTL    string `flag:"cache-ttl"`
	ConnKey     string `flag:"conn-key"`

	Path string `arg:"path"`
}

var genOp = &Operation{
	Name:     "gen",
	ParamPtr: (*genArg)(nil),

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*genArg)
		if arg.EnableLog {
			log.Init(true, arg.LogPath)
		}
		if arg.EnableCache {
			store.EnableCache()
		}

		if arg.CacheTTL != "" {
			ttl, err := time.ParseDuration(arg.CacheTTL)
			if err != nil {
				fmt.Printf(`time "%s" bad format`, arg.CacheTTL)
				fmt.Println()
				return false
			}
			store.SetCacheTTL(ttl)
		}
		if arg.ConnKey != "" {
			err := dbaccess.SetConn(arg.ConnKey)
			if err != nil {
				fmt.Printf("get connection %s failed: %v",
					arg.ConnKey, err)
				return false
			}
		}
		err := genfile.Do(arg.Path)
		if err != nil {
			fmt.Printf("%s %v\n", term.Red("[error]"), err)
			return false
		}
		return true
	},
}
