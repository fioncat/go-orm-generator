package genstruct

import (
	"github.com/fioncat/go-gendb/compile/mediate"
	"github.com/fioncat/go-gendb/generate/coder"
)

type Generator struct {
}

func (*Generator) Name() string {
	return "struct"
}

func (*Generator) ConfType() interface{}    { return nil }
func (*Generator) DefaultConf() interface{} { return nil }

func (*Generator) Do(c *coder.Coder, result mediate.Result, _ interface{}) error {
	ss := result.GetStructs()
	for _, s := range ss {
		c.AddStruct(*s)
	}
	return nil
}
