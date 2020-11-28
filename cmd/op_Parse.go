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

	Usage: parseUsage,
	Help:  parseHelp,

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

const parseUsage = `parse <file-path>`

const parseHelp = `
Parse performs lexical analysis and syntax analysis on
the file. The input of grammatical analysis is the result
of lexical analysis. Syntax analysis processes the output
structure of lexical analysis to produce intermediate
structures suitable for code generation. This command
outputs this intermediate structure in json format.

The syntax analysis will detect whether the usage of tags
in the go code and the writing of the code (including SQL)
meet expectations, and all possible syntax errors will be
checked out in this step.

Different marked go code will use different parsers. For
example, Go code marked with "+gendb sql" will be parsed
using the sql parser.

Use the following command to perform syntax analysis on
a go file:

go-gendb parse /path/to/parse.go

See also: scan, gen`
