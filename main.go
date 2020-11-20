package main

import (
	"os"
	"runtime"

	"github.com/fioncat/go-gendb/cmd"
	"github.com/fioncat/go-gendb/misc/term"
)

func main() {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		term.EnableColor()
	}
	cmd.Init()
	ctx := cmd.New(os.Args)
	ctx.Run()
}
