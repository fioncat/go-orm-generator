package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/store"
)

var cleanCacheCmd = &cmdt.Command{
	Name: "clean-cache",
	Pv:   (*cacheArg)(nil),

	Usage: help.CleanCacheUsage,
	Help:  help.CleanCache,

	Action: func(p interface{}) error {
		prefix := p.(*cacheArg).Prefix

		var total int
		err := store.WalkCache(prefix, func(path string, info os.FileInfo) error {
			err := os.Remove(path)
			if err != nil {
				return err
			}
			name := filepath.Base(path)
			fmt.Printf("Remove: %s\n", name)
			total += 1
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Printf("done, totally removed %d file(s).\n", total)
		return nil
	},
}
