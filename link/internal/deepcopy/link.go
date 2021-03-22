package deepcopy

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/deepcopy"
	"github.com/fioncat/go-gendb/compile/golang"
)

type Linker struct{}

func (*Linker) Do(file *golang.File, stc *golang.Struct, opts []base.Option) ([]coder.Target, error) {
	r, err := deepcopy.Parse(stc, opts)
	if err != nil {
		return nil, err
	}
	if r.Target == "" {
		r.Target = r.Name
	}
	if r.Method == "" {
		r.Method = "DeepCopy"
	}

	return []coder.Target{newTarget(file, r)}, nil
}
