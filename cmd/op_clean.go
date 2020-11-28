package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type cleanArg struct {
}

var cleanOp = &Operation{
	Name:     "clean",
	ParamPtr: (*cleanArg)(nil),

	Usage: cleanUsage,
	Help:  cleanHelp,

	Action: func(ctx *Context) bool {
		fileNameRe := regexp.MustCompile(`zz_generated\\.[^\\.]+\\.go`)

		paths := make([]string, 0)
		err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			fileName := filepath.Base(path)
			if strings.HasPrefix(fileName, ".") {
				return nil
			}
			if fileNameRe.MatchString(fileName) {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			fmt.Println(err)
			return false
		}

		for _, path := range paths {
			err := os.Remove(path)
			if err != nil {
				fmt.Println(err)
				return false
			}
			fmt.Printf("rm: %s\n", path)
		}
		return true
	},
}

const cleanUsage = `clean`

const cleanHelp = `
Clean will scan all the codes generated by go-gendb
in the current directory and delete them.

All generated codes meet the form of "zz_generated.*.go".
This command will delete such files.`
