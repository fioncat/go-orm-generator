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

	c *conf

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
	ic.Add(t.c.runName, t.c.runPath)
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
	for idx, f := range t.r.Fields {
		constName := fmt.Sprintf("%s%sField", t.r.Name, f.GoName)
		gp.Add(constName, coder.Quote(f.DbName))

		name := fmt.Sprintf("`%s`", f.DbName)
		selectFields[idx] = name
		if !f.AutoIncr {
			insertFields = append(insertFields, name)
		}
	}
	selectSql := fmt.Sprintf("SELECT %s FROM `%s`",
		strings.Join(selectFields, ","), t.r.Table)
	insertSql := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES",
		t.r.Table, strings.Join(insertFields, ","))
	deleteSql := fmt.Sprintf("DELETE FROM `%s`", t.r.Table)

	// sqls
	gp = c.NewGroup()

	valuesStr := strings.Repeat("?,", len(t.r.Fields))
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
}

func (t *target) StructNum() int { return 0 }

func (t *target) Struct(idx int, c *coder.Struct, ic *coder.Import) {}

func (t *target) FuncNum() int { return 0 }

func (t *target) Func(idx int, c *coder.Function, ic *coder.Import) {}
