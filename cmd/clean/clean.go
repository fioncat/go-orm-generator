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

	Usage: "clean [--cache]",
	Help:  desc,

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

const desc = `
Clean is used to delete the generated code or cached data.

Use this command directly can delete all go files starting
with "zz_generated" in the current directory.

Use the "--cache" flag, the clean command will clear all
current cached data. Cached data is generally obtained from
the database to speed up code generation. If the data is
inconsistent due to the cache or other reasons, you can use
"go-gendb clean --cache" to clear it directly.

See also: gen`
