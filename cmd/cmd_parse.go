package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/compile/parse"
	"github.com/fioncat/go-gendb/compile/scan/scango"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/term"
)

type parseArg struct {
	Path string `arg:"path"`
}

var parseCmd = &cmdt.Command{
	Name: "parse",
	Pv:   (*parseArg)(nil),

	Usage: help.ParseUsage,
	Help:  help.Parse,

	Action: func(p interface{}) error {
		arg := p.(*parseArg)
		if arg.Path == "" {
			return fmt.Errorf("path is empty")
		}

		data, err := ioutil.ReadFile(arg.Path)
		if err != nil {
			return err
		}

		scanResult, err := scango.Do(arg.Path, string(data))
		if err != nil {
			return err
		}

		results, err := parse.Do(scanResult)
		if err != nil {
			return err
		}
		for _, r := range results {
			term.Show(r)
			fmt.Println()
		}

		return nil
	},
}
