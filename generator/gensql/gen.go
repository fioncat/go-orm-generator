package gensql

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/generator/coder"
)

type Generator struct {
}

func (*Generator) Name() string {
	return "sql"
}

func (*Generator) Parse(task *generator.Task) ([]generator.File, error) {
	files := make([]generator.File, 0,
		len(task.Structs)+len(task.Interfaces))
	for _, ts := range task.Structs {
		file, err := parseStruct(task, ts)
		if err != nil {
			return nil, err
		}
		files = append(files, *file)
	}
	for _, it := range task.Interfaces {
		file, err := parseInter(task, it)
		if err != nil {
			return nil, err
		}
		files = append(files, *file)
	}
	return files, nil
}

func (*Generator) Generate(c *coder.Coder, file generator.File) {
	if s, ok := file.Result.(*coder.Struct); ok {
		c.AddStruct(*s)
		return
	}
	if r, ok := file.Result.(*sqlResult); ok {
		genInter(c, r)
		return
	}
}

var runnerPath = "github.com/fioncat/go-gendb/api/sqlrunner"

func SetRunnerPath(path string) {
	runnerPath = path
}

func genInter(c *coder.Coder, r *sqlResult) {
	for name, path := range r.imports {
		c.AddImport(name, path)
	}
	c.AddImport("runner", runnerPath)

	// generate interface implement
	structName := r.name + "Impl"
	c.P(0, "// ", structName, " implement of ", r.name)
	c.P(0, "type ", structName, " struct {")
	c.P(0, "}")
	c.Empty()
	c.P(0, "var ", r.name, "Oper ", r.name, " = &", structName, "{}")
	c.Empty()
	// c.AddVar(r.name+"Oper", "&", structName, "{}")

	for _, method := range r.methods {
		// sql content const
		constName := fmt.Sprintf("sql_%s_%s", r.name, method.name)
		c.AddConst(constName, coder.Quote(method.sql))

		// sql runner variable
		runnerName := fmt.Sprintf("runner_%s_%s", r.name, method.name)
		c.AddVar(runnerName, "runner.New(", constName, ")")

		// method definition
		c.P(0, "// ", method.name, " implement of ", method.name)
		c.P(0, "func (*", structName, ") ", method.definition, " {")
		genMethodBody(c, method, runnerName)
		c.P(0, "}")
		c.Empty()
	}
}

func genMethodBody(c *coder.Coder, m sqlMethod, runnerName string) {
	rep := "nil"
	if len(m.replaces) > 0 {
		rep = strings.Join(m.replaces, ", ")
		rep = fmt.Sprintf("[]interface{}{%s}", rep)
	}

	pre := "nil"
	if len(m.prepares) > 0 {
		pre = strings.Join(m.prepares, ", ")
		pre = fmt.Sprintf("[]interface{}{%s}", pre)
	}

	if !m.isQuery() && m.sqlType != sqlCount {
		switch m.sqlType {
		case sqlExecAffect:
			c.P(1, "return ", runnerName, ".ExecAffect(db, ", rep, ", ", pre, ")")
		case sqlExecLastid:
			c.P(1, "return ", runnerName, ".ExecLastId(db, ", rep, ", ", pre, ")")
		case sqlExecResult:
			c.P(1, "return ", runnerName, ".Exec(db, ", rep, ", ", pre, ")")
		}
		return
	}
	c.AddImport("sql", "database/sql")

	if m.sqlType == sqlQueryOne || m.sqlType == sqlCount {
		c.P(1, "var o ", m.retType)
		c.P(1, "err := ", runnerName, ".QueryOne(db, ", rep, ", ", pre, ", func(rows *sql.Rows) error {")

		if strings.HasPrefix(m.retType, "*") {
			retType := strings.TrimLeft(m.retType, "*")
			c.P(2, "o = new(", retType, ")")
		}

		if isSimple(m.retType) {
			c.P(2, "return rows.Scan(&o)")
		} else {
			c.P(2, "return rows.Scan(", assign(m.rets), ")")
		}

		c.P(1, "})")
		c.P(1, "return o, err")
		return
	}

	objType := strings.TrimLeft(m.retType, "[]")

	c.P(1, "var os ", m.retType)
	c.P(1, "err := ", runnerName, ".QueryMany(db, ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
	if strings.HasPrefix(objType, "*") {
		objType = strings.TrimLeft(objType, "*")
		c.P(2, "o := new(", objType, ")")
	} else {
		c.P(2, "var o ", objType)
	}
	if isSimple(objType) {
		c.P(2, "err := rows.Scan(&o)")
	} else {
		c.P(2, "err := rows.Scan(", assign(m.rets), ")")
	}
	c.P(2, "if err != nil {")
	c.P(3, "return err")
	c.P(2, "}")
	c.P(2, "os = append(os, o)")
	c.P(2, "return nil")
	c.P(1, "})")
	c.P(1, "return os, err")
}

func assign(rets []string) string {
	ss := make([]string, len(rets))
	for i, ret := range rets {
		ss[i] = fmt.Sprintf("&o.%s", ret)
	}
	return strings.Join(ss, ", ")
}

func isSimple(t string) bool {
	switch {
	case strings.HasPrefix(t, "int"):
		return true
	case strings.HasPrefix(t, "float"):
		return true
	case strings.HasPrefix(t, "uint"):
		return true
	case t == "string":
		return true
	case t == "bool":
		return true
	}
	return false
}
