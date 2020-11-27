package genstruct

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/parser/common"
)

type Generator struct {
}

func (*Generator) Name() string {
	return "go-struct"
}

func (*Generator) Generate(c *coder.Coder, r common.Result, _ string) error {
	structResult, ok := r.(*common.StructResult)
	if !ok {
		return errors.ErrInvalidType
	}

	for _, stc := range structResult.Structs {
		c.AddStruct(stc)
	}

	return nil
}
