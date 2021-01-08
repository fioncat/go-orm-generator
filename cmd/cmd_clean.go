package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/misc/cmdt"
)

type cleanArg struct {
	Path string `flag:"path"`
}

var cleanCmd = &cmdt.Command{
	Name: "clean",
	Pv:   (*cleanArg)(nil),

	Action: func(p interface{}) error {
		path := p.(*cleanArg).Path
		if path == "" {
			path = "."
		}
		var total int
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			filename := filepath.Base(path)
			if strings.HasPrefix(filename, "zz_generated") &&
				strings.HasSuffix(filename, ".go") {
				fmt.Printf("Remove: %s\n", path)
				total += 1
				return os.Remove(path)
			}
			return nil
		})
		if err != nil {
			return err
		}
		fmt.Printf("done, removed %d file(s)\n", total)
		return nil
	},
}
