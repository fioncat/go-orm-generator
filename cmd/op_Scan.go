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

	Usage: scanUsage,
	Help:  scanHelp,

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

const scanUsage = `scan <mode> <file-path>`

const scanHelp = `
Scan performs lexical analysis on the target file
and outputs the result in the form of json.

Lexical analysis is the first step in generating
code. Taking Go code as an example, Scan will
analyze various "tags" in it and generate a general
analysis structure. Scan only scans, and does not
check its grammar too much.

This command is mainly used for debugging. If the
code generation fails, and the reason is not clear,
you can use this command to view the results of
lexical analysis to troubleshoot the error.

Perform lexical analysis on a Go code file that
has been tagged with +gendb, and execute:

go-gendb scan go /path/to/scan.go

This will output json-format data to the terminal.

Now supported modes:

  go    golang code tagged with "+gendb {args}"
  sql   sql file tagged with "-- !{name}"

See also: parse, gen`
