package orm_mgo

import (
	"fmt"
	"strings"

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
	if !t.r.Db {
		return
	}
	ic.Add("", "gopkg.in/mgo.v2/bson")
}

func (t *target) Consts(c *coder.Var, ic *coder.Import) {
	var name string
	gp := c.NewGroup()
	if t.r.Db {
		name = fmt.Sprintf("%sFieldID", t.r.Name)
		gp.Add(name, coder.Quote("_id"))
	}
	for _, f := range t.r.Fields {
		name = fmt.Sprintf("%sField%s", t.r.Name, f.GoName)
		gp.Add(name, coder.Quote(f.DbName))
	}

	gp = c.NewGroup()
	if t.r.Db {
		name = fmt.Sprintf("%sSortIDDesc", t.r.Name)
		gp.Add(name, coder.Quote("-_id"))
	}
	for _, f := range t.r.Fields {
		if strings.Contains(f.GoName, "Time") ||
			strings.Contains(f.GoName, "Date") {
			f.Sort = true
		}
		if !f.Sort {
			continue
		}
		name = fmt.Sprintf("%sSort%sDesc", t.r.Name, f.GoName)
		gp.Add(name, coder.Quote("-"+f.DbName))
	}
}

func (t *target) Vars(c *coder.Var, ic *coder.Import) {
}

func (t *target) StructNum() int { return 1 }

func (t *target) Struct(_ int, c *coder.Struct, ic *coder.Import) {
	c.SetName(t.r.Name)
	if t.r.Db {
		f := c.AddField()
		f.Set("ID", "bson.ObjectId")
		f.AddTag("bson", "_id,omitempty")
		f.AddTag("json", "id")
	}
	for _, f := range t.r.Fields {
		gf := c.AddField()
		gf.Set(f.GoName, f.GoType)
		gf.AddTag("bson", f.DbName)
		gf.AddTag("json", f.DbName)
	}
}

func (t *target) Funcs(c *coder.FunctionGroup) {}

func (t *target) FuncNum() int { return 0 }

func (t *target) Func(idx int, c *coder.Function, ic *coder.Import) {}

func (t *target) Structs(c *coder.StructGroup) {}
