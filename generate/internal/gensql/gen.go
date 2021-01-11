package gensql

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/compile/mediate"
	"github.com/fioncat/go-gendb/compile/parse/parsesql"
	"github.com/fioncat/go-gendb/generate/coder"
)

// implementation of sql generator

type Conf struct {
	RunnerPath string `json:"runner_path"`
	DbUse      string `json:"db_use"`
	DbImport   string `json:"db_import"`
}

type Generator struct {
}

func (*Generator) Name() string {
	return "db-oper"
}

func (*Generator) ConfType() interface{} {
	return (*Conf)(nil)
}

func (*Generator) DefaultConf() interface{} {
	return &Conf{
		RunnerPath: "github.com/fioncat/go-gendb/api/sqlrunner",
		DbUse:      "db",
		DbImport:   "",
	}
}

func (*Generator) Do(c *coder.Coder, result mediate.Result, confv interface{}) error {
	conf := confv.(*Conf)

	if conf.DbImport != "" {
		name := filepath.Base(conf.DbImport)
		c.AddImport(name, conf.DbImport)
	}

	if conf.RunnerPath == "" {
		return fmt.Errorf("runner path is empty")
	}
	c.AddImport("runner", conf.RunnerPath)

	oper := result.(*parsesql.OperResult)
	structName := "_" + oper.Name + "Impl"

	c.P(0, "// ", structName, " implement of ", oper.Name)
	c.P(0, "type ", structName, " struct {")
	c.P(0, "}")
	c.Empty()

	c.P(0, "var ", oper.Name, "Oper ", oper.Name, " = &", structName, "{}")
	c.Empty()

	for _, m := range oper.Methods {
		constName := fmt.Sprintf("sql_%s_%s", oper.Name, m.Name)
		c.AddConst(constName, coder.Quote(m.SQL.Contant))

		runnerName := fmt.Sprintf("runner_%s_%s", oper.Name, m.Name)
		c.AddVar(runnerName, "runner.New(", constName, ")")

		c.P(0, "// ", m.Name, " implement of ", oper.Name, ".", m.Name)
		c.P(0, "func (*", structName, ") ", m.Origin, " {")
		body(c, &m, runnerName, conf)
		c.P(0, "}")
		c.Empty()

		for _, imp := range m.Imports {
			c.AddImport(imp.Name, imp.Path)
		}
	}

	return nil
}

func body(c *coder.Coder, m *parsesql.Method, runnerName string, conf *Conf) {
	rep := "nil"
	if len(m.SQL.Replaces) > 0 {
		rep = strings.Join(m.SQL.Replaces, ", ")
		rep = fmt.Sprintf("[]interface{}{%s}", rep)
	}

	pre := "nil"
	if len(m.SQL.Prepares) > 0 {
		pre = strings.Join(m.SQL.Prepares, ", ")
		pre = fmt.Sprintf("[]interface{}{%s}", pre)
	}

	if m.IsExec {
		if m.IsAffect() {
			c.P(1, "return ", runnerName, ".ExecAffect(", conf.DbUse, ", ", rep, ", ", pre, ")")
		}
		if m.IsLastId() {
			c.P(1, "return ", runnerName, ".ExecLastId(", conf.DbUse, ", ", rep, ", ", pre, ")")
		}
		if m.IsResult() {
			c.AddImport("sql", "database/sql")
			c.P(1, "return ", runnerName, ".Exec(", conf.DbUse, ", ", rep, ", ", pre, ")")
		}
		return
	}

	c.AddImport("sql", "database/sql")

	var typeDef = m.RetType
	if m.RetPointer {
		typeDef = "*" + typeDef
	}
	if m.RetSlice {
		typeDef = "[]" + typeDef
	}

	rets := make([]string, len(m.QueryFields))
	for i, f := range m.QueryFields {
		var name string
		if f.Alias != "" {
			name = f.Alias
		} else {
			name = coder.GoName(f.Field)
		}
		rets[i] = name
	}

	if m.IsQueryOne() {
		c.P(1, "var o ", typeDef)
		c.P(1, "err := ", runnerName, ".QueryOne(", conf.DbUse, ", ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
		if m.RetPointer {
			c.P(2, "o = new(", m.RetType, ")")
		}
		if coder.IsSimpleType(m.RetType) {
			c.P(2, "return rows.Scan(&o)")
		} else {
			c.P(2, "return rows.Scan(", assign(rets), ")")
		}
		c.P(1, "})")
		c.P(1, "return o, err")
		return
	}

	c.P(1, "var os ", typeDef)
	c.P(1, "err := ", runnerName, ".QueryMany(", conf.DbUse, ", ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
	if m.RetPointer {
		c.P(2, "o := new(", m.RetType, ")")
	} else {
		c.P(2, "var o ", m.RetType)
	}
	if coder.IsSimpleType(m.RetType) {
		c.P(2, "err := rows.Scan(&o)")
	} else {
		c.P(2, "err := rows.Scan(", assign(rets), ")")
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
