package cmd

import (
	"fmt"
	"os"

	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/misc/humansize"
	"github.com/fioncat/go-gendb/store"
)

type cacheArg struct {
}

var cacheSizeCmd = &cmdt.Command{
	Name: "cache-size",
	Pv:   (*cacheArg)(nil),

	Usage: help.CacheSizeUsage,
	Help:  help.CacheSize,

	Action: func(p interface{}) error {
		var totalSize int64
		var total int
		err := store.WalkCache(func(path string, info os.FileInfo) error {
			totalSize += info.Size()
			total += 1
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Printf("Total cached items: %d\n", total)

		fmt.Printf("Size: %s (%d Bytes)\n",
			humansize.Bytes(uint64(totalSize)), totalSize)
		return nil
	},
}
