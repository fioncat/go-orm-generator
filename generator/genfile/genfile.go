package genfile

import (
	"time"

	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/generator/coder"
	"github.com/fioncat/go-gendb/generator/gensql"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/scanner"
)

var gens = make(map[string]generator.Generator)

func init() {
	sqlGenerator := &gensql.Generator{}

	gens["mysql"] = sqlGenerator
}

func Do(path string) error {
	log.Info("Begin to Scan...")
	start := time.Now()
	task, err := scanner.Golang(path)
	if err != nil {
		return err
	}
	log.Infof("Scan done, took: %s",
		time.Since(start).String())

	generator, ok := gens[task.Type]
	if !ok {
		return errors.Fmt("can not find "+
			"generator for %s", task.Type)
	}

	start = time.Now()
	log.Infof("Begin to parsing...")
	files, err := generator.Parse(task)
	if err != nil {
		return err
	}
	log.Infof("Parse done, took: %s",
		time.Since(start).String())

	start = time.Now()
	log.Infof("Begin to generating...")
	for _, file := range files {
		c := &coder.Coder{}
		c.Pkg = task.Pkg
		c.Source = task.Path

		generator.Generate(c, file)
		err = c.Write(file.Path)
		if err != nil {
			return err
		}
		log.Infof("write file: %s", file.Path)
	}
	log.Infof("Generate done, took: %s",
		time.Since(start).String())

	return nil
}
