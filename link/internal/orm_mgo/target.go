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
	ic.Add("mgoapi", t.conf["mgoapi_path"])
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
	if !t.r.Db {
		return
	}
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
	f.Set("", "*mgoapi.Query")
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

	if len(t.r.Indexes) == 0 && len(t.r.UniqueKeys) == 0 {
		goto skipIndex
	}
	// Ensure Indexes
	f = c.Add()
	if sessUse == "sess" {
		f.Def("Ensure"+t.r.Name+"Indexes",
			"Ensure"+t.r.Name+"Indexes(sess *mgo.Session) error")
	} else {
		f.Def("Ensure"+t.r.Name+"Indexes",
			"Ensure"+t.r.Name+"Indexes() error")
	}
	f.P(0, "_sess, col := ", operName, ".GetCol(", colParam, ")")
	f.P(0, "defer _sess.Close()")
	for _, idx := range t.r.Indexes {
		names := make([]string, len(idx.Fields))
		for idx, field := range idx.Fields {
			names[idx] = t.r.Name + "Field" + field.GoName
		}
		f.P(0, "if err := col.EnsureIndex(mgo.Index{")
		f.P(1, "Key:        []string{", strings.Join(names, ","), "},")
		f.P(1, "Background: true,")
		f.P(1, "Sparse:     true,")
		f.P(0, "}); err != nil {")
		f.P(1, "return err")
		f.P(0, "}")
		f.P(0, "")
	}
	for _, idx := range t.r.UniqueKeys {
		names := make([]string, len(idx.Fields))
		for idx, field := range idx.Fields {
			names[idx] = t.r.Name + "Field" + field.GoName
		}
		f.P(0, "if err := col.EnsureIndex(mgo.Index{")
		f.P(1, "Key:        []string{", strings.Join(names, ","), "},")
		f.P(1, "Background: true,")
		f.P(1, "Unique:     true,")
		f.P(1, "Sparse:     true,")
		f.P(0, "}); err != nil {")
		f.P(1, "return err")
		f.P(0, "}")
		f.P(0, "")
	}
	f.P(0, "return nil")

skipIndex:

	f = c.Add()
	t.funcDef(false, f, "Save", nil, []string{"*mgo.ChangeInfo", "error"})
	f.P(0, "sess, col := ", operName, ".GetCol(", colParam, ")")
	f.P(0, "defer sess.Close()")
	f.P(0, "return col.UpsertId(o.Id, o)")

	f = c.Add()
	f.Def("All", "(q *", t.r.Name, "Query) All() (os []*", t.r.Name, ", err error)")
	f.P(0, "err = q.MarshalAll(&os)")
	f.P(0, "return")

	f = c.Add()
	f.Def("One", "(q *", t.r.Name, "Query) One() (o *", t.r.Name, ", err error)")
	f.P(0, "err = q.MarshalOne(&o)")
	f.P(0, "return")

	f = c.Add()
	f.Def("Walk", "(q *", t.r.Name, "Query) Walk(walkFunc func(o *", t.r.Name, ") error) error")
	f.P(0, "iter := q.Iter()")
	f.P(0, "var o *", t.r.Name)
	f.P(0, "for iter.Next(&o) {")
	f.P(1, "err := walkFunc(o)")
	f.P(1, "if err != nil {")
	f.P(2, "return err")
	f.P(1, "}")
	f.P(0, "}")
	f.P(0, "return iter.Err()")

	f = c.Add()
	t.funcDef(true, f, "GetCol", nil, []string{"*mgo.Session", "*mgo.Collection"})
	f.P(0, "_sess := ", sessUse, ".Clone()")
	f.P(0, "return _sess, _sess.DB(", t.r.Name, "Database).C(", t.r.Name, "Collection)")

	f = c.Add()
	t.funcDef(true, f, "Find", []string{"query interface{}"}, []string{"*" + t.r.Name + "Query"})
	f.P(0, "_sess, col := oper.GetCol(", colParam, ")")
	f.P(0, "mq := col.Find(query)")
	f.P(0, "return &", t.r.Name, "Query{Query: mgoapi.NewQuery(mq, _sess)}")

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

	f = c.Add()
	t.funcDef(true, f, "CountE", []string{"query interface{}"}, []string{"int", "error"})
	f.P(0, "_sess, col := oper.GetCol(", colParam, ")")
	f.P(0, "defer _sess.Close()")
	f.P(0, "return col.Find(query).Count()")

	f = c.Add()
	t.funcDef(true, f, "Count", []string{"query interface{}"}, []string{"int"})
	if sessUse == "sess" {
		f.P(0, "cnt, _ := oper.CountE(sess, query)")
	} else {
		f.P(0, "cnt, _ := oper.CountE(query)")
	}
	f.P(0, "return cnt")

	f = c.Add()
	t.funcDef(true, f, "Remove", []string{"query interface{}"}, []string{"*mgo.ChangeInfo", "error"})
	f.P(0, "_sess, col := oper.GetCol(", colParam, ")")
	f.P(0, "defer _sess.Close()")
	f.P(0, "return col.RemoveAll(query)")

	f = c.Add()
	t.funcDef(true, f, "RemoveById", []string{"id string"}, []string{"error"})
	f.P(0, "_sess, col := oper.GetCol(", colParam, ")")
	f.P(0, "defer _sess.Close()")
	f.P(0, "if !bson.IsObjectIdHex(id) {")
	f.P(1, "return mgo.ErrNotFound")
	f.P(0, "}")
	f.P(0, "return col.RemoveId(bson.ObjectIdHex(id))")

	// Indexes(single)
	for _, idx := range t.r.Indexes {
		if len(idx.Fields) != 1 {
			continue
		}
		field := idx.Fields[0]
		if strings.Contains(field.GoName, "Time") ||
			strings.Contains(field.GoName, "Date") {
			continue
		}
		f = c.Add()
		pName := coder.UnExport(field.GoName)
		t.funcDef(true, f, "FindManyBy"+field.GoName, []string{
			fmt.Sprintf("%s %s", pName, field.GoType)}, []string{"[]*" + t.r.Name, "error"})
		if sessUse == "sess" {
			f.P(0, "q := oper.Find(sess, bson.M{", t.r.Name, "Field", field.GoName, ": ", pName, "})")
		} else {
			f.P(0, "q := oper.Find(bson.M{", t.r.Name, "Field", field.GoName, ": ", pName, "})")
		}
		f.P(0, "return q.All()")
	}

	// Uniques(single)
	for _, idx := range t.r.UniqueKeys {
		if len(idx.Fields) != 1 {
			continue
		}
		field := idx.Fields[0]
		if strings.Contains(field.GoName, "Time") ||
			strings.Contains(field.GoName, "Date") {
			continue
		}
		f = c.Add()
		pName := coder.UnExport(field.GoName)
		t.funcDef(true, f, "FindOneBy"+field.GoName, []string{
			fmt.Sprintf("%s %s", pName, field.GoType)}, []string{"*" + t.r.Name, "error"})
		if sessUse == "sess" {
			f.P(0, "q := oper.Find(sess, bson.M{", t.r.Name, "Field", field.GoName, ": ", pName, "})")
		} else {
			f.P(0, "q := oper.Find(bson.M{", t.r.Name, "Field", field.GoName, ": ", pName, "})")
		}
		f.P(0, "return q.One()")
	}
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
