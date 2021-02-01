package cmd

import (
	"fmt"
	"os"

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
		var total int
		err := store.WalkCache(func(path string, info os.FileInfo) error {
			err := os.Remove(path)
			if err != nil {
				return err
			}
			total += 1
			return nil
		})
		if err != nil {
			return err
		}

		err = store.RemoveCache()
		if err != nil {
			return err
		}

		fmt.Printf("done, totally removed %d file(s).\n", total)
		return nil
	},
}
