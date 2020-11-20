package gensql

import (
	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/generator/coder"
)

type Generator struct {
}

func (*Generator) Name() string {
	return "sql"
}

func (*Generator) Parse(task *generator.Task) ([]generator.File, error) {
	files := make([]generator.File, 0,
		len(task.Structs)+len(task.Interfaces))
	for _, ts := range task.Structs {
		file, err := parseStruct(task, ts)
		if err != nil {
			return nil, err
		}
		files = append(files, *file)
	}
	return files, nil
}

func (*Generator) Generate(c *coder.Coder, file generator.File) {
	s := file.Result.(*coder.Struct)
	c.AddStruct(*s)
}
