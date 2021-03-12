package sql_orm

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

	operName string
	operType string
}

func (t *target) Name() string {
	return t.r.Name
}

func (t *target) Path() string {
	return t.path
}

func (t *target) Imports(ic *coder.Import) {
	ic.Add("", "database/sql")
	ic.Add(t.conf[runName], t.conf[runPath])
	ic.Add("", "strings")
	ic.Add("", "fmt")
}

func (t *target) Vars(c *coder.Var, ic *coder.Import) {
	group := c.NewGroup()
	group.Add(t.operName, "&", t.operType, "{}")
}

func (t *target) Consts(c *coder.Var, ic *coder.Import) {
	selectFields := make([]string, len(t.r.Fields))
	insertFields := make([]string, 0, len(t.r.Fields))
	// Field names
	gp := c.NewGroup()
	var valsCnt int
	for idx, f := range t.r.Fields {
		constName := fmt.Sprintf("%s%sField", t.r.Name, f.GoName)
		gp.Add(constName, coder.Quote(f.DbName))

		name := fmt.Sprintf("`%s`", f.DbName)
		selectFields[idx] = name
		if !f.AutoIncr {
			insertFields = append(insertFields, name)
			valsCnt++
		}
	}
	selectSql := fmt.Sprintf("SELECT %s FROM `%s`",
		strings.Join(selectFields, ","), t.r.Table)
	insertSql := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES",
		t.r.Table, strings.Join(insertFields, ","))
	deleteSql := fmt.Sprintf("DELETE FROM `%s`", t.r.Table)

	// sqls
	gp = c.NewGroup()

	valuesStr := strings.Repeat("?,", valsCnt)
	if len(valuesStr) >= 1 {
		valuesStr = valuesStr[:len(valuesStr)-1]
	}
	valuesStr = "(" + valuesStr + ")"

	// InsertOne
	name := fmt.Sprintf("_%s_InsertOne", t.r.Name)
	sql := fmt.Sprintf("%s %s", insertSql, valuesStr)
	gp.Add(name, coder.Quote(sql))

	// InsertBatch
	name = fmt.Sprintf("_%s_InsertBatch", t.r.Name)
	sql = insertSql + " %s"
	gp.Add(name, coder.Quote(sql))
	name = fmt.Sprintf("_%s_InsertValues", t.r.Name)
	gp.Add(name, coder.Quote(valuesStr))

	// FindById and DeleteById
	ids := make([]string, len(t.r.PrimaryKey.Fields))
	idMap := make(map[string]struct{}, len(t.r.PrimaryKey.Fields))
	for idx, f := range t.r.PrimaryKey.Fields {
		ids[idx] = fmt.Sprintf("`%s`=?", f.DbName)
		idMap[f.DbName] = struct{}{}
	}
	idCond := strings.Join(ids, " AND ")
	name = fmt.Sprintf("_%s_FindById", t.r.Name)
	sql = fmt.Sprintf("%s WHERE %s", selectSql, idCond)
	gp.Add(name, coder.Quote(sql))

	name = fmt.Sprintf("_%s_DeleteById", t.r.Name)
	sql = fmt.Sprintf("%s WHERE %s", deleteSql, idCond)
	gp.Add(name, coder.Quote(sql))

	// UpdateById
	nonIds := make([]string, 0, len(t.r.Fields))
	for _, f := range t.r.Fields {
		_, ok := idMap[f.DbName]
		if ok {
			continue
		}
		assign := fmt.Sprintf("`%s`=?", f.DbName)
		nonIds = append(nonIds, assign)
	}

	name = fmt.Sprintf("_%s_UpdateById", t.r.Name)
	sql = fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", t.r.Table,
		strings.Join(nonIds, ","), idCond)
	gp.Add(name, coder.Quote(sql))

	// Count
	name = fmt.Sprintf("_%s_Count", t.r.Name)
	sql = fmt.Sprintf("SELECT COUNT(1) FROM `%s`", t.r.Table)
	gp.Add(name, coder.Quote(sql))

	gp = c.NewGroup()
	// uniques(single)
	for _, key := range t.r.UniqueKeys {
		if len(key.Fields) != 1 {
			continue
		}
		f := key.Fields[0]
		name = fmt.Sprintf("_%s_FindOneBy%s", t.r.Name, f.GoName)
		sql = fmt.Sprintf("%s WHERE `%s`=?", selectSql, f.DbName)
		gp.Add(name, coder.Quote(sql))
	}

	gp = c.NewGroup()
	// indexes(single)
	for _, key := range t.r.Indexes {
		if len(key.Fields) != 1 {
			continue
		}
		f := key.Fields[0]
		if strings.Contains(f.GoName, "Date") {
			continue
		}
		if strings.Contains(f.GoName, "Time") {
			continue
		}
		name = fmt.Sprintf("_%s_FindManyBy%s", t.r.Name, f.GoName)
		sql = fmt.Sprintf("%s WHERE `%s`=?", selectSql, f.DbName)
		gp.Add(name, coder.Quote(sql))
	}
}

func (t *target) Structs(sg *coder.StructGroup) {
	s := sg.Add()
	s.SetName(t.operType)

	s = sg.Add()
	s.SetName(t.r.Name)
	s.Comment(t.r.Comment)
	for _, rf := range t.r.Fields {
		gf := s.AddField()
		gf.Set(rf.GoName, rf.GoType)
		gf.AddTag("field", rf.DbName)
	}
}

func (t *target) Funcs(fg *coder.FunctionGroup) {
	dbUse := t.conf[dbUse]
	runUse := t.conf[runName]

	idParams := make([]string, len(t.r.PrimaryKey.Fields))
	idNames := make([]string, len(t.r.PrimaryKey.Fields))
	idUpdate := make([]string, len(t.r.PrimaryKey.Fields))
	idMap := make(map[string]struct{}, len(t.r.PrimaryKey.Fields))
	for idx, f := range t.r.PrimaryKey.Fields {
		idParams[idx] = fmt.Sprintf("%s %s",
			coder.UnExport(f.GoName), f.GoType)
		idNames[idx] = coder.UnExport(f.GoName)
		idMap[f.GoName] = struct{}{}
		idUpdate[idx] = fmt.Sprintf("o.%s", f.GoName)
	}

	insertParams := make([]string, 0, len(t.r.Fields))
	selectFields := make([]string, len(t.r.Fields))
	updateParams := make([]string, 0, len(t.r.Fields))
	for idx, f := range t.r.Fields {
		if _, ok := idMap[f.GoName]; !ok {
			updateParams = append(updateParams,
				fmt.Sprintf("o.%s", f.GoName))
		}
		selectFields[idx] = fmt.Sprintf("&o.%s", f.GoName)
		if f.AutoIncr {
			continue
		}
		insertParams = append(insertParams,
			fmt.Sprintf("o.%s", f.GoName))
	}

	// InsertOne
	f := fg.Add()
	t.funcDef(f, "Insert", []string{"o *" + t.r.Name}, "sql.Result")
	sqlName := fmt.Sprintf("_%s_InsertOne", t.r.Name)
	f.P(0, "return run.Exec(", dbUse, ", ", sqlName,
		", nil, []interface{}{", strings.Join(insertParams, ", "), "})")

	// InsertBatch
	f = fg.Add()
	t.funcDef(f, "InsertBatch", []string{"os []*" + t.r.Name}, "sql.Result")
	sqlName = fmt.Sprintf("_%s_InsertBatch", t.r.Name)
	f.P(0, "vs := make([]interface{}, 0, ", len(t.r.Fields), "*len(os))")
	f.P(0, "valStrs := make([]string, len(os))")
	f.P(0, "for idx, o := range os {")
	f.P(1, "valStrs[idx] = _", t.r.Name, "_InsertValues")
	f.P(1, "vs = append(vs, ", strings.Join(insertParams, ", "), ")")
	f.P(0, "}")
	f.P(0, "valStr := strings.Join(valStrs, ", coder.Quote(", "), ")")
	f.P(0, "_sql := fmt.Sprintf(", sqlName, ", valStr)")
	f.P(0, "return run.Exec(", dbUse, ", _sql, nil, vs)")

	// FindById
	f = fg.Add()
	t.funcDef(f, "FindById", idParams, "*"+t.r.Name)
	sqlName = fmt.Sprintf("_%s_FindById", t.r.Name)
	f.P(0, "var o *", t.r.Name)
	f.P(0, "err := ", runUse, ".QueryOne(", dbUse, ", ", sqlName,
		", nil, []interface{}{", strings.Join(idNames, ", "),
		"}, func(rows *sql.Rows) error {")
	f.P(1, "o = new(", t.r.Name, ")")
	f.P(1, "return rows.Scan(", strings.Join(selectFields, ", "), ")")
	f.P(0, "})")
	f.P(0, "return o, err")

	// DeleteById
	f = fg.Add()
	t.funcDef(f, "DeleteById", idParams, "sql.Result")
	sqlName = fmt.Sprintf("_%s_DeleteById", t.r.Name)
	f.P(0, "return run.Exec(", dbUse, ", ", sqlName,
		", nil, []interface{}{", strings.Join(idNames, ", "), "})")

	// UpdateById
	f = fg.Add()
	t.funcDef(f, "UpdateById", []string{"o *" + t.r.Name}, "sql.Result")
	params := append(updateParams, idUpdate...)
	sqlName = fmt.Sprintf("_%s_UpdateById", t.r.Name)
	f.P(0, "return run.Exec(", dbUse, ", ", sqlName,
		", nil, []interface{}{", strings.Join(params, ", "), "})")

	// Count
	f = fg.Add()
	t.funcDef(f, "Count", []string{}, "int64")
	sqlName = fmt.Sprintf("_%s_Count", t.r.Name)
	f.P(0, "var cnt int64")
	f.P(0, "err := run.QueryOne(", dbUse, ", ", sqlName, ", nil, nil, func(rows *sql.Rows) error {")
	f.P(1, "return rows.Scan(&cnt)")
	f.P(0, "})")
	f.P(0, "return cnt, err")

	// FindByUnique(single)
	for _, key := range t.r.UniqueKeys {
		if len(key.Fields) != 1 {
			continue
		}
		field := key.Fields[0]
		sqlName = fmt.Sprintf("_%s_FindOneBy%s",
			t.r.Name, field.GoName)
		name := fmt.Sprintf("FindOneBy%s", field.GoName)
		paramName := coder.UnExport(field.GoName)
		param := fmt.Sprintf("%s %s", paramName, field.GoType)

		f = fg.Add()
		t.funcDef(f, name, []string{param}, "*"+t.r.Name)
		f.P(0, "var o *", t.r.Name)
		f.P(0, "err := run.QueryOne(", dbUse, ", ", sqlName,
			", nil, []interface{}{", paramName, "}, func(rows *sql.Rows) error {")
		f.P(1, "o = new(", t.r.Name, ")")
		f.P(1, "return rows.Scan(", strings.Join(selectFields, ", "), ")")
		f.P(0, "})")
		f.P(0, "return o, err")
	}

	// FindByIndex(single)
	for _, index := range t.r.Indexes {
		if len(index.Fields) != 1 {
			continue
		}
		field := index.Fields[0]
		if strings.Contains(field.GoName, "Date") {
			continue
		}
		if strings.Contains(field.GoName, "Time") {
			continue
		}
		sqlName = fmt.Sprintf("_%s_FindManyBy%s",
			t.r.Name, field.GoName)
		name := fmt.Sprintf("FindManyBy%s", field.GoName)
		paramName := coder.UnExport(field.GoName)
		param := fmt.Sprintf("%s %s", paramName, field.GoType)

		f = fg.Add()
		t.funcDef(f, name, []string{param}, "[]*"+t.r.Name)
		f.P(0, "var os []*", t.r.Name)
		f.P(0, "err := run.QueryMany(", dbUse, ", ", sqlName,
			", nil, []interface{}{", paramName, "}, func(rows *sql.Rows) error {")
		f.P(1, "o := new(", t.r.Name, ")")
		f.P(1, "err := rows.Scan(", strings.Join(selectFields, ", "), ")")
		f.P(1, "if err != nil {")
		f.P(2, "return err")
		f.P(1, "}")
		f.P(1, "os = append(os, o)")
		f.P(1, "return nil")
		f.P(0, "})")
		f.P(0, "return os, err")
	}

}

func (t *target) funcDef(f *coder.Function, name string, params []string, ret string) {
	dbUse := t.conf[dbUse]
	def := fmt.Sprintf("(*%s) %s(", t.operType, name)
	if dbUse == "db" {
		def += "db run.IDB"
	}
	if len(params) > 0 {
		def += ", " + strings.Join(params, ", ")
	}
	def += ") "
	def += fmt.Sprintf("(%s, error)", ret)
	f.Def(name, def)
}

func (t *target) StructNum() int { return 0 }

func (t *target) Struct(idx int, c *coder.Struct, ic *coder.Import) {}

func (t *target) FuncNum() int { return 0 }

func (t *target) Func(idx int, c *coder.Function, ic *coder.Import) {}
