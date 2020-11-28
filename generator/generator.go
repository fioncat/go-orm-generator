package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/generator/internal/gensql"
	"github.com/fioncat/go-gendb/generator/internal/genstruct"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/parser/common"
)

type Generator interface {
	Name() string
	Generate(c *coder.Coder, r common.Result, confPath string) error
}

var generators = make(map[string]Generator)

func Reg(generator Generator) {
	generators[generator.Name()] = generator
}

func init() {
	Reg(&genstruct.Generator{})
	Reg(&gensql.Generator{})
}

func Do(path, dir, pkg, confPath string, results []common.Result) error {
	generatedPaths := make([]string, 0, len(results))
	for _, r := range results {
		genor := generators[r.Generator()]
		if genor == nil {
			onRemove(generatedPaths)
			return errors.Fmt("Can not find generator '%s'",
				r.Generator())
		}

		c := new(coder.Coder)
		c.Source = path
		c.Pkg = pkg
		err := genor.Generate(c, r, confPath)
		if err != nil {
			onRemove(generatedPaths)
			return errors.Trace(genor.Name(), err)
		}

		targetName := fmt.Sprintf("zz_generated.%s.go", r.Key())
		targetPath := filepath.Join(dir, targetName)
		err = c.Write(targetPath)
		if err != nil {
			onRemove(generatedPaths)
			return errors.Trace(path, err)
		}
		generatedPaths = append(generatedPaths, targetPath)
	}
	return nil
}

func onRemove(paths []string) {
	for _, path := range paths {
		os.Remove(path)
	}
}
