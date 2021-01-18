package gensqlorm

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/mediate"
	"github.com/fioncat/go-gendb/compile/parse/parsesql"
	"github.com/fioncat/go-gendb/generate/coder"
)

type Conf struct {
	RunPath    string
	UpdatePath string
	QueryPath  string
}

type Generator struct {
}

func (*Generator) Name() string {
	return "sql-orm"
}

func (*Generator) ConfType() interface{} {
	return (*Conf)(nil)
}

func (*Generator) DefaultConf() interface{} {
	return &Conf{
		RunPath:    "github.com/fioncat/go-gendb/api/sql/run",
		UpdatePath: "github.com/fioncat/go-gendb/api/sql/update",
		QueryPath:  "github.com/fioncat/go-gendb/api/sql/query",
	}
}

type genContext struct {
	sqlfWithid string
	sqlfNoid   string

	scanf string

	assignfWithid string
	assignfNoid   string

	valuesWithid string
	valuesNoid   string

	table string
	name  string

	idName     string
	isAutoIncr bool

	idConst string

	update string
}

func (ctx *genContext) op(name string) string {
	return fmt.Sprintf("_%s_%s", ctx.name, name)
}

func (ctx *genContext) _const(name string) string {
	return fmt.Sprintf("%sField%s", ctx.name, name)
}

func (*Generator) Do(c *coder.Coder, r mediate.Result, confv interface{}) error {
	conf := confv.(*Conf)

	c.AddImport("fmt", "fmt")
	c.AddImport("query", conf.QueryPath)
	c.AddImport("run", conf.RunPath)
	c.AddImport("sql", "database/sql")
	c.AddImport("strings", "strings")
	c.AddImport("update", conf.UpdatePath)

	result := r.(*parsesql.OrmResult)

	c.AddStruct(result.Struct)

	ctx := new(genContext)
	ctx.name = result.GoName
	ctx.table = _field(result.TableName)
	ctx.idName = _field(result.IdField.Name)
	ctx.isAutoIncr = result.IdField.IsAutoIncr

	// field Const, and build context
	fwithids := make([]string, len(result.Fields))
	fnoids := make([]string, 0, len(result.Fields)-1)
	scanfs := make([]string, len(result.Fields))
	asswithids := make([]string, len(result.Fields))
	assnoids := make([]string, 0, len(result.Fields)-1)
	updates := make([]string, 0, len(result.Fields)-1)

	c.P(0, "// all fields for ", ctx.table)
	c.P(0, "const (")
	for idx, field := range result.Fields {
		fwithids[idx] = _field(field.Name)
		scanfs[idx] = fmt.Sprintf("&o.%s", field.GoName)
		asswithids[idx] = fmt.Sprintf("o.%s", field.GoName)

		if !field.IsPrimary {
			fnoids = append(fnoids, _field(field.Name))
			assnoids = append(assnoids, fmt.Sprintf("o.%s", field.GoName))
			updates = append(updates, fmt.Sprintf("%s=?", _field(field.Name)))
		}

		c.P(1, "// ", ctx._const(field.GoName), " is the name of field ", ctx.table, ".", field.Name)
		c.P(1, ctx._const(field.GoName), " = ", coder.Quote(_field(field.Name)))
	}
	c.Empty()
	c.P(1, ctx.name, "Table = ", coder.Quote(ctx.table))
	c.P(0, ")")

	ctx.sqlfWithid = strings.Join(fwithids, ",")
	ctx.sqlfNoid = strings.Join(fnoids, ",")
	ctx.scanf = strings.Join(scanfs, ", ")
	ctx.assignfWithid = strings.Join(asswithids, ", ")
	ctx.assignfNoid = strings.Join(assnoids, ", ")
	ctx.update = strings.Join(updates, ",")

	ctx.valuesWithid = values(len(result.Fields))
	ctx.valuesNoid = values(len(result.Fields) - 1)

	c.Empty()
	c.P(0, "// auto-generated sql statements")
	c.P(0, "const (")
	// sql const
	// Insert
	var sql string
	c.P(1, "// ", ctx.op("Insert"), " is the sql for <Insert> operation.")
	if ctx.isAutoIncr {
		sql = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)",
			ctx.table, ctx.sqlfNoid, ctx.valuesNoid)
	} else {
		sql = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)",
			ctx.table, ctx.sqlfWithid, ctx.valuesWithid)
	}
	c.P(1, ctx.op("Insert"), " = ", coder.Quote(sql))

	// Batch
	c.P(1, "// ", ctx.op("Batch"), " is the sql for <Insert-Batch> operation.")
	var values string
	if ctx.isAutoIncr {
		sql = fmt.Sprintf("INSERT INTO %s(%s) VALUES",
			ctx.table, ctx.sqlfNoid)
		values = ctx.valuesNoid
	} else {
		sql = fmt.Sprintf("INSERT INTO %s(%s) VALUES",
			ctx.table, ctx.sqlfWithid)
		values = ctx.valuesWithid
	}
	sql += " %s"
	c.P(1, ctx.op("Batch_0"), " = ", coder.Quote(sql))
	c.P(1, ctx.op("Batch_1"), " = ", coder.Quote("("+values+"),"))

	// Delete
	c.P(1, "// ", ctx.op("Delete"), " is the sql for <Delete> operation.")
	sql = fmt.Sprintf("DELETE FROM %s WHERE %s=?", ctx.table, ctx.idName)
	c.P(1, ctx.op("Delete"), " = ", coder.Quote(sql))

	// Update
	c.P(1, "// ", ctx.op("Update"), " is the sql for <Update> operation.")
	sql = fmt.Sprintf("UPDATE %s SET %s WHERE %s=?", ctx.table,
		ctx.update, ctx.idName)
	c.P(1, ctx.op("Update"), " = ", coder.Quote(sql))

	// Upsert
	c.P(1, "// ", ctx.op("Upsert"), " is the sql for <Upsert> operation.")
	sql = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		ctx.table, ctx.sqlfWithid, ctx.valuesWithid, ctx.update)
	c.P(1, ctx.op("Upsert"), " = ", coder.Quote(sql))

	// FindById
	sql = fmt.Sprintf("SELECT %s FROM %s WHERE %s=?",
		ctx.sqlfWithid, ctx.table, ctx.idName)
	c.P(1, "// ", ctx.op("FindById"), " is the sql for <FindById> operation.")
	c.P(1, ctx.op("FindById"), " = ", coder.Quote(sql))

	// FindAll
	sql = fmt.Sprintf("SELECT %s FROM %s", ctx.sqlfWithid, ctx.table)
	c.P(1, "// ", ctx.op("FindAll"), " is the sql for <FindAll> operation.")
	c.P(1, ctx.op("FindAll"), " = ", coder.Quote(sql))

	c.P(0, ")")
	c.Empty()

	c.P(0, "// ", ctx.name, "Fields stores all field names for table ", ctx.table)
	c.P(0, "var ", ctx.name, "Fields = []string{")
	for _, f := range result.Fields {
		c.P(1, ctx._const(f.GoName), ",")
	}
	c.P(0, "}")
	c.Empty()

	ormName := ctx.name + "ORM"
	implName := ormName + "Impl"
	c.P(0, "// ", ormName, " is the global variable to process orm-operation for ", ctx.table)
	c.P(0, "var ", ormName, " = &", implName, "{}")
	c.Empty()

	c.P(0, "type ", implName, " struct {")
	c.P(0, "}")
	c.Empty()

	// BatchInsert
	nFields := len(result.Fields)
	if result.IdField.IsAutoIncr {
		nFields--
	}
	c.P(0, "func (*", implName, ") BatchInsert(db run.IDB, os []*", ctx.name, ") (sql.Result, error) {")
	c.P(1, "vs := make([]interface{}, 0, ", nFields, "*len(os))")
	c.P(1, "for _, o := range os {")
	if ctx.isAutoIncr {
		c.P(2, "vs = append(vs, ", ctx.assignfNoid, ")")
	} else {
		c.P(2, "vs = append(vs, ", ctx.assignfWithid, ")")
	}
	c.P(1, "}")
	c.P(1, "valuesStr := strings.Repeat(", ctx.op("Batch_1"), ", len(os))")
	c.P(1, "valuesStr = valuesStr[:len(valuesStr)-1]")
	c.P(1, "_sql := fmt.Sprintf(", ctx.op("Batch_0"), ", valuesStr)")
	c.P(1, "return run.Exec(db, _sql, nil, vs)")
	c.P(0, "}")
	c.Empty()

	c.P(0, "func (*", implName, ") Update(db run.IDB, u *update.Update) (sql.Result, error) {")
	c.P(1, "_sql, vs := u.Build(", ctx.name, "Table)")
	c.P(1, "return run.Exec(db, _sql, nil, vs)")
	c.P(0, "}")
	c.Empty()

	// FindById
	c.P(0, "func (*", implName, ") FindById(db run.IDB, id ", result.IdField.GoType, ") (o *", ctx.name, ", err error) {")
	c.P(1, "err = run.QueryOne(db, ", ctx.op("FindById"), ", nil, []interface{}{id}, func(rows *sql.Rows) error {")
	c.P(2, "o = new(", ctx.name, ")")
	c.P(2, "return rows.Scan(", ctx.scanf, ")")
	c.P(1, "})")
	c.P(1, "return")
	c.P(0, "}")
	c.Empty()

	// Search
	c.P(0, "func (*", implName, ") Search(db run.IDB, q *query.Query) (os []*", ctx.name, ", err error) {")
	c.P(1, "_sql, vs := q.Build(", ctx.name, "Table, ", ctx.name, "Fields)")
	c.P(1, "err = run.QueryMany(db, _sql, nil, vs, func(rows *sql.Rows) error {")
	c.P(2, "o := new(", ctx.name, ")")
	c.P(2, "err := rows.Scan(", ctx.scanf, ")")
	c.P(2, "if err != nil {")
	c.P(3, "return err")
	c.P(2, "}")
	c.P(2, "os = append(os, o)")
	c.P(2, "return nil")
	c.P(1, "})")
	c.P(1, "return")
	c.P(0, "}")
	c.Empty()

	// Walk
	c.P(0, "func (*", implName, ") Walk(db run.IDB, batchSize int64, walkFunc func(os []*", ctx.name, ") error) error {")
	c.P(1, "_sql := ", ctx.op("FindAll"), " + ", coder.Quote(" LIMIT ?,?"))
	c.P(1, "var offset int64 = 0")
	c.P(1, "for {")
	c.P(2, "var os []*", ctx.name)
	c.P(2, "err := run.QueryMany(db, _sql, nil, []interface{}{offset, batchSize}, func(rows *sql.Rows) error {")
	c.P(3, "o := new(", ctx.name, ")")
	c.P(3, "err := rows.Scan(", ctx.scanf, ")")
	c.P(3, "if err != nil {")
	c.P(4, "return err")
	c.P(3, "}")
	c.P(3, "os = append(os, o)")
	c.P(3, "return nil")
	c.P(2, "})")
	c.P(2, "if err != nil {")
	c.P(3, "return err")
	c.P(2, "}")
	c.P(2, "if len(os) == 0 {")
	c.P(3, "return nil")
	c.P(2, "}")
	c.P(2, "offset += batchSize")
	c.P(2, "err = walkFunc(os)")
	c.P(2, "if err != nil {")
	c.P(3, "return err")
	c.P(2, "}")
	c.P(1, "}")
	c.P(0, "}")
	c.Empty()

	// Save(Upsert)
	c.P(0, "func (o *", ctx.name, ") Save(db run.IDB) (sql.Result, error) {")
	c.P(1, "return run.Exec(db, ", ctx.op("Upsert"), ", nil, []interface{}{", ctx.assignfWithid, ", ", ctx.assignfNoid, "})")
	c.P(0, "}")
	c.Empty()

	// Add(Insert)
	var insert0 string
	if ctx.isAutoIncr {
		insert0 = ctx.assignfNoid
	} else {
		insert0 = ctx.assignfWithid
	}
	c.P(0, "func (o *", ctx.name, ") Add(db run.IDB) (sql.Result, error) {")
	c.P(1, "return run.Exec(db, ", ctx.op("Insert"), ", nil, []interface{}{", insert0, "})")
	c.P(0, "}")
	c.Empty()

	// UpdateById
	c.P(0, "func (o *", ctx.name, ") UpdateById(db run.IDB) (sql.Result, error) {")
	c.P(1, "return run.Exec(db, ", ctx.op("Update"), ", nil, []interface{}{", ctx.assignfNoid, ", o.", result.IdField.GoName, "})")
	c.P(0, "}")
	c.Empty()

	// Remove
	c.P(0, "func (o *", ctx.name, ") Remove(db run.IDB) (sql.Result, error) {")
	c.P(1, "return run.Exec(db, ", ctx.op("Delete"), ", nil, []interface{}{o.", result.IdField.GoName, "})")
	c.P(0, "}")
	c.Empty()

	return nil
}

func _field(s string) string {
	return fmt.Sprintf("`%s`", s)
}

func values(n int) string {
	s := strings.Repeat("?,", n)
	return s[:len(s)-1]
}
