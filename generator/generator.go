package generator

import "github.com/fioncat/go-gendb/generator/coder"

type Generator interface {
	Name() string
	Parse(task *Task) ([]File, error)
	Generate(c *coder.Coder, file File)
}

type File struct {
	Path   string
	Result interface{}
}
