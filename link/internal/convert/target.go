package convert

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/convert"
)

type target struct {
	coder.NoVars
	coder.NoConsts
	coder.NoStructs
	coder.NoFuncs
	coder.NoStructNum

	path string

	r *convert.Result
}

func (t *target) Name() string {
	return "ex.convert." + t.r.Name
}

func (t *target) Path() string {
	return t.path
}

func (t *target) Imports(ic *coder.Import) {
	for _, imp := range t.r.Imports {
		ic.Add("", imp)
	}
}

func (t *target) FuncNum() int { return 1 }

func (t *target) Func(_ int, c *coder.Function, ic *coder.Import) {
	c.Def(t.r.Method, "(b *", t.r.Name, ") ",
		t.r.Method, "() *", t.r.Target)
	c.P(0, "a := new(", t.r.Target, ")")
	for _, f := range t.r.Fields {
		c.P(0, f.Left, " = ", f.Right)
	}
	c.P(0, "return a")
}
