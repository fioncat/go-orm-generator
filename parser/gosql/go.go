package gosql

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/parser/common"
	"github.com/fioncat/go-gendb/scanner"
)

const (
	sqlExecAffect = iota
	sqlExecLastid
	sqlExecResult
	sqlQueryMany
	sqlQueryOne
)

type OperResult struct {
	Name string
}

type Method struct {
	Type int

	Def string

	RetType string
	Imports []coder.Import

	QueryFields []string
}

func Parse(sr *scanner.GoResult) ([]common.Result, error) {
}
