package cmd

import (
	"fmt"

	"github.com/fioncat/go-gendb/parser/gosql"
	"github.com/fioncat/go-gendb/scanner"
)

type ParseArg struct {
	Path string `arg:"path"`
}

var parseOp = &Operation{
	Name:     "parse",
	ParamPtr: (*ParseArg)(nil),

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*ParseArg)
		sr, err := scanner.Go(arg.Path, false)
		if err != nil {
			fmt.Println(err)
			return false
		}

		switch sr.Type {
		case "sql":
			_, err = gosql.Parse(sr, true)
			if err != nil {
				fmt.Println(err)
				return false
			}

		default:
			fmt.Printf("Unknown type '%s'\n", sr.Type)
			return false
		}

		return true
	},
}
