package orm_mgo

import (
	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/orm"
)

type target struct {
	path string

	r *orm.Result

	conf map[string]string
}

func (t *target) Name() string {
	return t.r.Name
}

func (t *target) Path() string {
	return t.path
}

func (t *target) Imports(ic *coder.Import) {
}

func (t *target) Vars(c *coder.Var, ic *coder.Import) {
}
