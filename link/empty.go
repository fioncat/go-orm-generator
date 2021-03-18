package link

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
)

type EmptyConf struct{}

func (*EmptyConf) DefaultConf() map[string]string {
	return nil
}

type EmptyDo struct{}

func (*EmptyDo) Do(file *golang.File, conf map[string]string) ([]coder.Target, error) {
	return nil, nil
}

type emptyLinker struct {
	EmptyConf
	EmptyDo
}
