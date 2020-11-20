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
)

type Arg struct {
	LogPath   string `flag:"log-file"`
	EnableLog bool   `flag:"enable-log"`

	FilePath string `arg:"file-path"`
}

var op = &cmd.Operation{
	ParamPtr: (*Arg)(nil),

	Action: func(ctx *cmd.Context) bool {
		arg := ctx.Param().(*Arg)
		if arg.EnableLog {
			log.Init(true, arg.LogPath)
		}

		log.Info("Begin to Scan...")
		start := time.Now()
		task, err := scanner.Golang(arg.FilePath)
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

func main() {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		term.EnableColor()
	}
	cmd.Register(op)
	ctx := cmd.New(os.Args)
	ctx.Run()
}
