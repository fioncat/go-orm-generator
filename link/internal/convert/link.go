package convert

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/convert"
	"github.com/fioncat/go-gendb/compile/golang"
)

type Linker struct{}

func (*Linker) Do(stc *golang.Struct, opts []base.Option) ([]coder.Target, error) {
	t := new(target)
	r, err := convert.Parse(stc, opts)
	if err != nil {
		return nil, err
	}
	t.r = r
	return []coder.Target{t}, nil
}
