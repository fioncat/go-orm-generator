package gensql

import (
	"fmt"
	"path/filepath"
	"sort"
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
		RunnerPath: "github.com/fioncat/go-gendb/api/sql/run",
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

	sort.Slice(oper.Methods, func(i, j int) bool {
		return oper.Methods[i].Name < oper.Methods[j].Name
	})

	for _, m := range oper.Methods {
		var constName string
		var (
			pre = "nil"
			rep = "nil"

			hasPre = false
			hasRep = false
		)
		if m.IsDynamic {
			constName = "_sql"
			for idx, part := range m.DynamicParts {
				if part.ForJoin != "" {
					part.SQL.Contant += part.ForJoin
				}
				c.AddConst(fmt.Sprintf("sql_%s_%s_%d",
					oper.Name, m.Name, idx), coder.Quote(part.SQL.Contant))
				if len(part.SQL.Prepares) > 0 {
					hasPre = true
				}
				if len(part.SQL.Replaces) > 0 {
					hasRep = true
				}
			}

			if hasPre {
				pre = "pvs"
			}
			if hasRep {
				rep = "rvs"
			}
		} else {
			constName = fmt.Sprintf("sql_%s_%s", oper.Name, m.Name)
			c.AddConst(constName, coder.Quote(m.SQL.Contant))
			if len(m.SQL.Replaces) > 0 {
				rep = strings.Join(m.SQL.Replaces, ", ")
				rep = fmt.Sprintf("[]interface{}{%s}", rep)
			}
			if len(m.SQL.Prepares) > 0 {
				pre = strings.Join(m.SQL.Prepares, ", ")
				pre = fmt.Sprintf("[]interface{}{%s}", pre)
			}
		}

		c.P(0, "// ", m.Name, " implement of ", oper.Name, ".", m.Name)
		c.P(0, "func (*", structName, ") ", m.Origin, " {")
		if m.IsDynamic {
			concat(c, m, oper, hasPre, hasRep)
		}
		body(c, &m, "runner", constName, rep, pre, conf)
		c.P(0, "}")
		c.Empty()

		for _, imp := range m.Imports {
			c.AddImport(imp.Name, imp.Path)
		}
	}

	return nil
}

func concat(c *coder.Coder, m parsesql.Method, oper *parsesql.OperResult, hasPre, hasRep bool) {
	c.P(1, "// >>> concat start.")
	preCap, repCap := calcphcap(m)
	if hasPre {
		c.P(1, "pvs := make([]interface{}, 0, ", preCap, ")")
	}
	if hasRep {
		c.P(1, "rvs := make([]interface{}, 0, ", repCap, ")")
	}
	c.P(1, "slice := make([]string, 0, ", calcsqlcap(m), ")")
	initLastidx := false
	for idx, part := range m.DynamicParts {
		name := fmt.Sprintf("sql_%s_%s_%d",
			oper.Name, m.Name, idx)
		c.P(1, "// concat: part ", idx)

		switch part.Type {
		case parsesql.DynamicTypeConst:
			c.P(1, "slice = append(slice, ", name, ")")
			genph(1, c, part)

		case parsesql.DynamicTypeIf:
			c.P(1, "if ", part.IfCond, " {")
			c.P(2, "slice = append(slice, ", name, ")")
			genph(2, c, part)
			c.P(1, "}")

		case parsesql.DynamicTypeFor:
			if part.ForJoin != "" {
				if initLastidx {
					c.P(1, "lastidx = len(", name, ") - ", len(part.ForJoin))
				} else {
					c.P(1, "lastidx := len(", name, ") - ", len(part.ForJoin))
					initLastidx = true
				}
			}
			genfor(c, part)
			genph(2, c, part)
			if part.ForJoin != "" {
				c.P(2, "if i == len(", part.ForSlice, ") - 1 {")
				c.P(3, "slice = append(slice, ", name, "[:lastidx])")
				c.P(2, "} else {")
				c.P(3, "slice = append(slice, ", name, ")")
				c.P(2, "}")
			} else {
				c.P(2, "slice = append(slice, ", name, ")")
			}
			c.P(1, "}")
		}
	}
	c.AddImport("strings", "strings")
	c.P(1, "// do concat")
	c.P(1, "_sql := strings.Join(slice, ", "\" \")")
	c.P(1, "// >>> concat done.")
}

func genfor(c *coder.Coder, part *parsesql.DynamicPart) {
	hasIdx := part.ForJoin != ""
	hasEle := part.ForEle != ""

	if hasIdx && hasEle {
		c.P(1, "for i, ", part.ForEle, " := range ", part.ForSlice, " {")
	} else if hasIdx && !hasEle {
		c.P(1, "for i := range ", part.ForSlice, "{")
	} else if !hasIdx && hasEle {
		c.P(1, "for _, ", part.ForEle, " := range ", part.ForSlice, " {")
	} else {
		c.P(1, "for range ", part.ForSlice, " {")
	}
}

func genph(nTab int, c *coder.Coder, part *parsesql.DynamicPart) {
	if len(part.SQL.Prepares) > 0 {
		c.P(nTab, "pvs = append(pvs, ", strings.Join(part.SQL.Prepares, ", "), ")")
	}
	if len(part.SQL.Replaces) > 0 {
		c.P(nTab, "pvs = append(rvs, ", strings.Join(part.SQL.Replaces, ", "), ")")
	}
}

func calcphcap(m parsesql.Method) (string, string) {
	var (
		preCnt = 0
		repCnt = 0

		preSlice []string
		repSlice []string
	)
	for _, part := range m.DynamicParts {
		switch part.Type {
		case parsesql.DynamicTypeConst:
			fallthrough
		case parsesql.DynamicTypeIf:
			preCnt += len(part.SQL.Prepares)
			repCnt += len(part.SQL.Replaces)

		case parsesql.DynamicTypeFor:

			var sliceCap string
			if len(part.SQL.Prepares) > 0 {
				if len(part.SQL.Prepares) > 1 {
					sliceCap = fmt.Sprintf("%d*len(%s)",
						len(part.SQL.Prepares), part.ForSlice)
				} else {
					sliceCap = fmt.Sprintf("len(%s)",
						part.ForSlice)
				}
				preSlice = append(preSlice, sliceCap)
			}

			if len(part.SQL.Replaces) > 0 {
				if len(part.SQL.Replaces) > 1 {
					sliceCap = fmt.Sprintf("%d*len(%s)",
						len(part.SQL.Replaces), part.ForSlice)
				} else {
					sliceCap = fmt.Sprintf("len(%s)",
						part.ForSlice)
				}
				repSlice = append(repSlice, sliceCap)
			}
		}
	}
	return calccap(preCnt, preSlice),
		calccap(repCnt, repSlice)
}

func calcsqlcap(m parsesql.Method) string {
	cap := 0
	var slices []string
	for _, part := range m.DynamicParts {
		switch part.Type {
		case parsesql.DynamicTypeIf:
			fallthrough
		case parsesql.DynamicTypeConst:
			cap += 1
		case parsesql.DynamicTypeFor:
			slicecap := fmt.Sprintf("len(%s)", part.ForSlice)
			slices = append(slices, slicecap)
		}
	}
	return calccap(cap, slices)
}

func calccap(cap int, extracts []string) string {
	if cap == 0 {
		if len(extracts) == 0 {
			return "0"
		}
		return strings.Join(extracts, "+")
	}
	if len(extracts) == 0 {
		return fmt.Sprint(cap)
	}
	return fmt.Sprintf("%d+%s", cap, strings.Join(extracts, "+"))

}

func body(c *coder.Coder, m *parsesql.Method, runnerName, sqlName, rep, pre string, conf *Conf) {
	if m.IsExec {
		if m.IsAffect() {
			c.P(1, "return ", runnerName, ".ExecAffect(", conf.DbUse, ", ", sqlName, ", ", rep, ", ", pre, ")")
		}
		if m.IsLastId() {
			c.P(1, "return ", runnerName, ".ExecLastId(", conf.DbUse, ", ", sqlName, ", ", rep, ", ", pre, ")")
		}
		if m.IsResult() {
			c.AddImport("sql", "database/sql")
			c.P(1, "return ", runnerName, ".Exec(", conf.DbUse, ", ", sqlName, ", ", rep, ", ", pre, ")")
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
		c.P(1, "err := ", runnerName, ".QueryOne(", conf.DbUse, ", ", sqlName, ", ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
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
	c.P(1, "err := ", runnerName, ".QueryMany(", conf.DbUse, ", ", sqlName, ", ", rep, ", ", pre, ", func(rows *sql.Rows) error {")
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
