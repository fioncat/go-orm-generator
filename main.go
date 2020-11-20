package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/fioncat/go-gendb/cmd"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/scanner"
	"github.com/fioncat/go-gendb/store"
)

func main() {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		term.EnableColor()
	}
	cmd.Register(op)
	ctx := cmd.New(os.Args)
	ctx.Run()
}

type Arg struct {
	LogPath     string `flag:"log-file"`
	EnableLog   bool   `flag:"enable-log"`
	EnableCache bool   `flag:"enable-cache"`
	CacheTTL    string `flag:"cache-ttl"`
	SetConn     bool   `flag:"set-conn"`
	DelConn     bool   `flag:"del-conn"`
	ShowConn    bool   `flag:"show-conn"`
	ConnKey     string `flag:"conn-key"`
	Path        string `flag:"path"`
}

var op = &cmd.Operation{
	ParamPtr: (*Arg)(nil),
	Action: func(ctx *cmd.Context) bool {
		arg := ctx.Param().(*Arg)
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

		if arg.SetConn {
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
		}

		if arg.ShowConn || arg.DelConn {
			if arg.ConnKey == "" {
				fmt.Println(`missing "--conn-key" flag`)
				return false
			}
			if arg.ShowConn {
				conn, err := store.GetConn(arg.ConnKey)
				if err != nil {
					fmt.Println(err)
					return false
				}
				store.ShowConn(conn)
				return true
			}
			err := store.DelConn(arg.ConnKey)
			if err != nil {
				fmt.Println(err)
				return false
			}
			return true
		}
		if arg.Path == "" {
			ctx.Usage()
			return false
		}

		log.Info("Begin to Scan...")
		start := time.Now()
		task, err := scanner.Golang(arg.Path)
		if err != nil {
			fmt.Printf("%s %v\n", term.Red("[error]"), err)
			return false
		}
		log.Infof("Scan done, took: %s",
			time.Since(start).String())

		fmt.Println(task)
		return true
	},
}
