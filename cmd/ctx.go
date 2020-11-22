package cmd

import (
	"fmt"
	"os"

	"github.com/fioncat/go-gendb/build"
)

func Init() {
	Register(genOp)
	Register(cleanOp)
	Register(connSetOp)
	Register(connGetOp)
	Register(connDelOp)
}

type Context struct {
	op    *Operation
	param interface{}

	help string
}

func (ctx *Context) Usage() {
	fmt.Printf("Usage: go-gendb %s\n", ctx.op.Usage)
	fmt.Printf("Use \"go-gendb %s -h\" for more information\n", ctx.op.Name)
}

func (ctx *Context) MainUsage() {
	fmt.Println(ctx.help)
}

func (ctx *Context) Param() interface{} {
	return ctx.param
}

func (ctx *Context) Help() {
	fmt.Printf("Usage: go-gendb %s\n", ctx.op.Usage)
	fmt.Println(ctx.op.Help)
}

func (ctx *Context) MainHelp() {
	fmt.Println(ctx.help)
}

func New(args []string) *Context {
	args = args[1:]
	ctx := new(Context)
	ctx.help = help
	if len(args) == 0 {
		ctx.MainUsage()
		os.Exit(1)
	}

	master := args[0]
	if master == "-h" {
		ctx.MainHelp()
		os.Exit(0)
	}
	switch master {
	case "-h":
		ctx.MainHelp()
		os.Exit(0)
	case "help":
		if len(args) == 1 {
			ctx.MainHelp()
			os.Exit(0)
		}
		cmd := args[1]
		op := ops[cmd]
		if op == nil {
			fmt.Printf("help: unknown command \"%s\"\n", cmd)
			os.Exit(1)
		}
		fmt.Printf("Usage: go-gendb %s\n", op.Usage)
		fmt.Println(op.Help)
		os.Exit(0)
	case "-v", "version":
		build.ShowVersion()
		os.Exit(0)
	}

	ctx.op = ops[master]
	if ctx.op == nil {
		fmt.Printf("go-gendb: unknown command \"%s\"\n", master)
		os.Exit(1)
	}
	args = args[1:]
	if len(args) > 0 {
		if args[0] == "-h" {
			ctx.Help()
			os.Exit(0)
		}
	}
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
