package golang

import (
	"fmt"

	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/misc/errors"
)

type Import struct {
	Name string
	Path string
}

type Interface struct {
	line int

	Name    string
	Tag     *base.Tag
	Methods []*Method
}

func (i *Interface) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(i.line, err)
}

type Method struct {
	line int

	Tags []*base.Tag

	Name    string
	Imports []string

	RetSlice   bool
	RetPointer bool
	RetSimple  bool

	RetType string

	Def string
}

func (m *Method) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(m.line, err)
}
