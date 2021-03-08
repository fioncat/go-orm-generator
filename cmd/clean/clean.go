package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/misc/cmdt"
)

type Arg struct {
	Cache bool `flag:"cache"`
}

var Cmder = &cmdt.Command{
	Name: "clean",
	Pv:   (*Arg)(nil),

	Action: func(p interface{}) error {
		arg := p.(*Arg)
		if arg.Cache {
			// TODO: clean cache
			return nil
		}
		return cleanGen()
	},
}

const genPrefix = "zz_generated"

func cleanGen() error {
	var paths []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if !strings.HasSuffix(filename, ".go") {
			return nil
		}
		if strings.HasPrefix(filename, genPrefix) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, path := range paths {
		err = os.Remove(path)
		if err != nil {
			return err
		}
		fmt.Printf("rm: %s\n", path)
	}
	return nil
}
