package gstruct

import (
	"sort"

	"github.com/fioncat/go-gendb/compile/mediate"
	"github.com/fioncat/go-gendb/generate/coder"
)

// implementation of struct generator

type Generator struct {
}

func (*Generator) Name() string {
	return "struct"
}

func (*Generator) ConfType() interface{}    { return nil }
func (*Generator) DefaultConf() interface{} { return nil }

func (*Generator) Do(c *coder.Coder, result mediate.Result, _ interface{}) error {
	ss := result.GetStructs()
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name < ss[j].Name
	})
	for _, s := range ss {
		c.AddStruct(*s)
	}
	return nil
}
