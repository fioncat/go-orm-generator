package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/scanner"
)

type ScanArg struct {
	Mode string `arg:"mode"`
	Path string `arg:"path"`
}

var scanOp = &Operation{
	Name:     "scan",
	ParamPtr: (*ScanArg)(nil),

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*ScanArg)
		var err error
		switch arg.Mode {
		case "go":
			var res *scanner.GoResult
			res, err = scanner.Go(arg.Path, true)
			if err == nil {
				term.Show(res)
			}
		case "sql":
			var res *scanner.SQLResult
			res, err = scanner.SQLFile(arg.Path, true)
			if err == nil {
				term.Show(res)
			}
		default:
			fmt.Printf(`unknown mode "%s"`, arg.Mode)
			return false
		}

		if err != nil {
			fmt.Println(err)
			return false
		}
		return true
	},
}
