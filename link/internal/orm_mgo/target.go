package orm_mgo

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/orm"
)

type target struct {
	dbName string

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
	ic.Add("mgo", "gopkg.in/mgo.v2")
}

func (t *target) Consts(c *coder.Var, ic *coder.Import) {
	if t.r.Db {
		table := t.r.Table
		if table == "" {
			table = t.r.Name
		}
		gp := c.NewGroup()
		gp.Add(t.r.Name+"Collection", coder.Quote(table))
		gp.Add(t.r.Name+"Database", coder.Quote(t.dbName))
	}

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
	gp := c.NewGroup()
	gp.Add(t.r.Name+"Oper", "&_", t.r.Name, "Oper{}")
}

func (t *target) Structs(c *coder.StructGroup) {
	s := c.Add()
	s.SetName(t.r.Name)
	if t.r.Db {
		f := s.AddField()
		f.Set("ID", "bson.ObjectId")
		f.AddTag("bson", "_id,omitempty")
		f.AddTag("json", "id")
	}
	for _, f := range t.r.Fields {
		gf := s.AddField()
		gf.Set(f.GoName, f.GoType)
		gf.AddTag("bson", f.DbName)
		gf.AddTag("json", f.DbName)
	}

	// Only create Oper type for DB struct.
	if !t.r.Db {
		return
	}

	s = c.Add()
	s.SetName("_" + t.r.Name + "Oper")

	s = c.Add()
	s.SetName(t.r.Name + "Query")
	f := s.AddField()
	f.Set("mq", "*mgo.Query")
	f = s.AddField()
	f.Set("sess", "*mgo.Session")
}

func (t *target) Funcs(c *coder.FunctionGroup) {
	if !t.r.Db {
		return
	}
	sessUse := t.conf["sess_use"]
	var colParam string
	if sessUse == "sess" {
		colParam = "sess"
	}

	f := c.Add()
	f.Def("Id", "(o *", t.r.Name, ") Id() string")
	f.P(0, "return o.ID.Hex()")

	operName := t.r.Name + "Oper"

	f = c.Add()
	t.funcDef(false, f, "Save", nil, []string{"*mgo.ChangeInfo", "error"})
	f.P(0, "sess, col := ", operName, ".GetCol(", colParam, ")")
	f.P(0, "defer sess.Close()")
	f.P(0, "return col.UpsertId(o.Id, o)")

	f = c.Add()
	f.Def("Select", "(q *", t.r.Name, "Query) Select(fields ...string) *", t.r.Name, "Query")
	f.P(0, "m := make(bson.M, len(fields))")
	f.P(0, "for _, field := range fields {")
	f.P(1, "m[field] = 1")
	f.P(0, "}")
	f.P(0, "q.mq.Select(m)")
	f.P(0, "return q")

	f = c.Add()
	f.Def("Limit", "(q *", t.r.Name, "Query) Limit(offset, limit int) *", t.r.Name, "Query")
	f.P(0, "if limit > 0 {")
	f.P(1, "q.mq.Limit(limit)")
	f.P(0, "}")
	f.P(0, "if offset > 0 {")
	f.P(1, "q.mq.Skip(offset)")
	f.P(0, "}")
	f.P(0, "return q")

	f = c.Add()
	f.Def("Sort", "(q *", t.r.Name, "Query) Sort(fields ...string) *", t.r.Name, "Query")
	f.P(0, "q.mq.Sort(fields...)")
	f.P(0, "return q")

	f = c.Add()
	f.Def("All", "(q *", t.r.Name, "Query) All() (os []*", t.r.Name, ", err error)")
	f.P(0, "defer q.sess.Close()")
	f.P(0, "err = q.mq.All(&os)")
	f.P(0, "return")

	f = c.Add()
	f.Def("One", "(q *", t.r.Name, "Query) One() (o *", t.r.Name, ", err error)")
	f.P(0, "defer q.sess.Close()")
	f.P(0, "err = q.mq.One(&o)")
	f.P(0, "return")

	f = c.Add()
	t.funcDef(true, f, "GetCol", nil, []string{"*mgo.Session", "*mgo.Collection"})
	f.P(0, "_sess := ", sessUse, ".Clone()")
	f.P(0, "return _sess, _sess.DB(", t.r.Name, "Database).C(", t.r.Name, "Collection)")

	f = c.Add()
	t.funcDef(true, f, "Find", []string{"query interface{}"}, []string{"*" + t.r.Name + "Query"})
	f.P(0, "_sess, col := oper.GetCol(", colParam, ")")
	f.P(0, "mq := col.Find(query)")
	f.P(0, "return &", t.r.Name, "Query{mq: mq, sess: _sess}")

	f = c.Add()
	t.funcDef(true, f, "FindById", []string{"id string"}, []string{"*" + t.r.Name, "error"})
	f.P(0, "if !bson.IsObjectIdHex(id) {")
	f.P(1, "return nil, mgo.ErrNotFound")
	f.P(0, "}")
	f.P(0, "_sess, col := oper.GetCol(", colParam, ")")
	f.P(0, "defer _sess.Close()")
	f.P(0, "var o *", t.r.Name)
	f.P(0, "err := col.FindId(bson.ObjectIdHex(id)).One(&o)")
	f.P(0, "return o, err")
}

func (t *target) funcDef(isOper bool, f *coder.Function, name string, params []string, rets []string) {
	sessUse := t.conf["sess_use"]
	var def string
	if isOper {
		def = fmt.Sprintf("(oper *%s) %s(", "_"+t.r.Name+"Oper", name)
	} else {
		def = fmt.Sprintf("(o *%s) %s(", t.r.Name, name)
	}
	if sessUse == "sess" {
		def += "sess *mgo.Session"
	}
	if len(params) > 0 {
		if sessUse == "sess" {
			def += ", "
		}
		def += strings.Join(params, ", ")
	}
	def += ") "
	switch len(rets) {
	case 1:
		def += rets[0]

	default:
		def += "(" + strings.Join(rets, ", ") + ")"
	}
	f.Def(name, def)
}

func (t *target) StructNum() int { return 0 }

func (t *target) Struct(idx int, c *coder.Struct, ic *coder.Import) {}

func (t *target) FuncNum() int { return 0 }

func (t *target) Func(idx int, c *coder.Function, ic *coder.Import) {}
