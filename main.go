package main

import (
	"os"
	"runtime"

	"github.com/fioncat/go-gendb/cmd"
	"github.com/fioncat/go-gendb/misc/term"
)

func main() {
	// enable color only in unix-like term.
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		term.EnableColor()
	}
	cmd.Execute(os.Args)
}
