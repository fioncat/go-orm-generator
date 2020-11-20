package cmd

import (
	"fmt"
	"os"

	"github.com/fioncat/go-gendb/build"
)

type Context struct {
	op    *Operation
	param interface{}

	usage string
	help  string
}

func (ctx *Context) Usage() {
	fmt.Printf("Usage: %s\n", ctx.usage)
}

func (ctx *Context) Param() interface{} {
	return ctx.param
}

func (ctx *Context) Help() {
	fmt.Printf("Usage:  %s\n", ctx.usage)
	fmt.Println(ctx.help)
}

func New(args []string) *Context {
	args = args[1:]
	ctx := new(Context)
	if len(args) == 0 {
		ctx.Help()
		os.Exit(1)
	}

	master := args[0]
	if master == "-h" {
		ctx.Help()
		os.Exit(0)
	}
	switch master {
	case "-h", "help":
		ctx.Help()
		os.Exit(0)
	case "-v", "version":
		build.ShowVersion()
		os.Exit(0)
	}

	ctx.op = genop

	param, err := ctx.op.ParseParam(args)
	if err != nil {
		fmt.Println(err)
		ctx.Usage()
		os.Exit(1)
	}
	ctx.param = param

	return ctx
}

func (ctx *Context) Run() {
	suc := ctx.op.Action(ctx)
	if suc {
		os.Exit(0)
	}
	os.Exit(1)
}
