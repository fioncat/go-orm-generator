package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/compile/parse"
	"github.com/fioncat/go-gendb/compile/scan/scango"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
)

type parseArg struct {
	Conn   string `flag:"conn"`
	DbType string `flag:"db-type" default:"mysql"`

	Cache bool `flag:"cache-table"`
	Log   bool `flag:"log"`

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

		if arg.Log {
			log.Init(true, "")
		}

		if arg.Conn != "" {
			err := rdb.Init(arg.Conn, arg.DbType)
			if err != nil {
				return errors.Trace("init db connection", err)
			}
		}

		if arg.Cache {
			rdb.ENABLE_TABLE_CACHE = true
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
			fmt.Printf("// >>>>>>>>> type: %s\n", r.Type())
			term.Show(r)
			fmt.Println()
		}

		return nil
	},
}
