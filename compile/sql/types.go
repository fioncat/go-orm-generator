package sql

import (
	"fmt"

	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/misc/errors"
)

type Method struct {
	line int

	Inter string
	Name  string

	State *Statement

	Exec bool

	Dyn bool
	Dps []*DynamicPart

	Fields []*QueryField

	Tags []*base.Tag
}

func (m *Method) LineIdx() int {
	return m.line - 1
}

func (m *Method) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(m.line, err)
}

// dynamic type enum
const (
	DynamicTypeConst = iota
	DynamicTypeIf
	DynamicTypeFor
)

type DynamicPart struct {
	Type int

	State *Statement

	IfCond string

	ForEle   string
	ForSlice string
	ForJoin  string
}

type Statement struct {
	Sql string

	Replaces []string
	Prepares []string

	phs []*placeholder
}

type placeholder struct {
	pre  bool
	name string
}

func (ph *placeholder) String() string {
	if ph.pre {
		return "$" + ph.name
	}
	return "#" + ph.name
}

type Var struct {
}

type QueryField struct {
	Table string
	Name  string
	Alias string

	IsCount bool
}
