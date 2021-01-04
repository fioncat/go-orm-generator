package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/fioncat/go-gendb/build"
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/compile/scan/scango"
	"github.com/fioncat/go-gendb/compile/scan/scansql"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type scanArg struct {
	Mode string `arg:"mode"`
	Path string `arg:"path"`
}

var scanCmd = &cmdt.Command{
	Name: "scan",
	Pv:   (*scanArg)(nil),

	Usage: help.ScanUsage,
	Help:  help.Scan,

	Action: func(p interface{}) error {
		arg := p.(*scanArg)

		data, err := ioutil.ReadFile(arg.Path)
		if err != nil {
			return err
		}
		content := string(data)

		build.DEBUG = true
		switch arg.Mode {
		case "go":
			var r *scango.Result
			r, err = scango.Do(arg.Path, content)
			if err == nil {
				term.Show(r)
			}
		case "sql":
			var r *scansql.Result
			r, err = scansql.Do(arg.Path, content)
			if err == nil {
				term.Show(r)
			}
		default:
			return fmt.Errorf("unknown mode %s", arg.Mode)
		}

		return err
	},
}
