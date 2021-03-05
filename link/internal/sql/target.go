package sql

import (
	"strconv"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/sql"
)

type target struct {
	file  *golang.File
	inter *golang.Interface

	name string

	ms []*method

	rets []*ret

	conf *conf
}

type method struct {
	gom  *golang.Method
	sqlm *sql.Method

	lastId bool
}

type ret struct {
	name string

	m  *method
	fs []*field
}

type field struct {
	dbTable string
	dbField string

	name   string
	goType string
}

func (t *target) Name() string {
	return t.inter.Tag.Name
}

func (t *target) Path() string {
	return t.file.Path
}

func (t *target) Vars(c *coder.Var, _ *coder.Import) {
	vg := c.NewGroup()
	vg.Add(t.name, "&_", t.name, "{}")
}

func (t *target) Consts(c *coder.Var, _ *coder.Import) {
	vg := c.NewGroup()

	for _, m := range t.ms {
		if !m.sqlm.Dyn {
			vg.Add("_"+t.name+"_"+m.gom.Name,
				coder.Quote(m.sqlm.State.Sql))
			continue
		}
		for idx, dp := range m.sqlm.Dps {
			vg.Add("_"+t.name+"_"+m.gom.Name+strconv.Itoa(idx),
				coder.Quote(dp.State.Sql))
		}
	}
}

func (t *target) StructNum() int {
	return len(t.rets) + 1
}

func (t *target) Struct(idx int, c *coder.Struct, _ *coder.Import) {
	if idx == len(t.rets) {
		c.SetName("_" + t.name)
		return
	}
	ret := t.rets[idx]
	c.SetName(ret.name)
	for _, f := range ret.fs {
		field := c.AddField()
		field.Set(f.name, f.goType)
		field.AddTag("table", f.dbTable)
		field.AddTag("field", f.dbField)
	}
}

func (t *target) FuncNum() int {
	return len(t.ms)
}

func (t *target) Func(idx int, c *coder.Function, imp *coder.Import) {
	m := t.ms[idx]

	imp.Add("sql", "database/sql")
	imp.Add("fmt", "fmt")
	c.Def(m.gom.Name, "(*_", t.name, ") ", m.gom.Def)
	c.P(0, "var a int")
}
