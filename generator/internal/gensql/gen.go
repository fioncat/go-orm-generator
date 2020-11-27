package gensql

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/parser/common"
	"github.com/fioncat/go-gendb/parser/gosql"
	"github.com/fioncat/go-gendb/store"
)

type Generator struct {
}

func (*Generator) Name() string {
	return "sql-oper"
}

type GenArg struct {
	RunnerPath string `json:"runner_path"`
	DbUse      string `json:"db_use"`
	DbImport   string `json:"db_import"`
}

func defaultArg() GenArg {
	return GenArg{
		RunnerPath: "github.com/fioncat/go-gendb/api/sqlrunner",
		DbUse:      "db",
		DbImport:   "",
	}
}

func (*Generator) Generate(c *coder.Coder, r common.Result, confPath string) error {
	or, ok := r.(*gosql.OperResult)
	if !ok {
		return errors.ErrInvalidType
	}

	arg := defaultArg()
	if confPath != "" {
		err := store.UnmarshalConf(confPath, &arg)
		if err != nil {
			return errors.Trace("read conf", err)
		}
	}

	if arg.DbImport != "" {
		name := filepath.Base(arg.DbImport)
		c.AddImport(name, arg.DbImport)
	}
	if arg.RunnerPath == "" {
		return errors.New("runner-path is empty")
	}

	c.AddImport("runner", arg.RunnerPath)

	structName := "_" + or.Name + "Impl"
	c.P(0, "// ", structName, " implement of ", or.Name)
	c.P(0, "type ", structName, " struct {")
	c.P(0, "}")
	c.Empty()

	c.P(0, "var ", or.Name, "Oper ", or.Name, " = &", structName, "{}")
	c.Empty()

	for _, m := range or.Methods {
		constName := fmt.Sprintf("sql_%s_%s", or.Name, m.Name)
		c.AddConst(constName, coder.Quote(m.SQL.String))

		runnerName := fmt.Sprintf("runner_%s_%s", or.Name, m.Name)
		c.AddVar(runnerName, "runner.New(", constName, ")")

		c.P(0, "// ", m.Name, " implement of ", or.Name, ".", m.Name)
		c.P(0, "func (*", structName, ") ", m.Def, " {")
		genMethodBody(c, m, runnerName, arg)
		c.P(0, "}")
		c.Empty()
	}

	return nil
}

func genMethodBody(c *coder.Coder, m gosql.Method, runnerName string, arg GenArg) {
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

	if m.IsExec() {
		if m.IsAffect() {
			c.P(1, "return ", runnerName, ".ExecAffect(", arg.DbUse, ", ", rep, ", ", pre, ")")
		}
		if m.IsLastId() {
			c.P(1, "return ", runnerName, ".ExecLastId(", arg.DbUse, ", ", rep, ", ", pre, ")")
		}
		if m.IsResult() {
			c.AddImport("sql", "database/sql")
			c.P(1, "return ", runnerName, ".Exec(", arg.DbUse, ", ", rep, ", ", pre, ")")
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
		c.P(1, "err := ", runnerName, ".QueryOne(", arg.DbUse, ", ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
		if m.RetPointer {
			c.P(2, "o = new(", m.RetType, ")")
		}
		if isSimple(m.RetType) {
			c.P(2, "return rows.Scan(&o)")
		} else {
			c.P(2, "return rows.Scan(", assign(rets), ")")
		}
		c.P(1, "})")
		c.P(1, "return o, err")
		return
	}

	c.P(1, "var os ", typeDef)
	c.P(1, "err := ", runnerName, ".QueryMany(", arg.DbUse, ", ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
	if m.RetPointer {
		c.P(2, "o := new(", m.RetType, ")")
	} else {
		c.P(2, "var o ", m.RetType)
	}
	if isSimple(m.RetType) {
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
