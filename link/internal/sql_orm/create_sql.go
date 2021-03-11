package sql_orm

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/orm"
	"github.com/fioncat/go-gendb/version"
)

func writeCreateTables(db, path string, rs []*orm.Result) error {
	c := new(coder.Coder)
	c.P(0, "-- ------------------------------------------------------------")
	c.P(0, "-- go-gendb v", version.Short)
	c.P(0, "--")
	c.P(0, "-- https://github.com/fioncat/go-gendb")
	c.P(0, "--")
	c.P(0, "-- This file is auto generated. DO NOT EDIT.")
	c.P(0, "-- ------------------------------------------------------------")
	if db != "" {
		c.P(0, "USE ", db, ";")
	}
	for _, r := range rs {
		c.Empty()
		createTable(c, r)
	}

	return c.WriteFile(path)
}

func createTable(c *coder.Coder, r *orm.Result) {
	c.P(0, "CREATE TABLE `", r.Table, "` (")
	c.Empty()
	idMap := make(map[string]struct{}, len(r.PrimaryKey.Fields))
	for _, f := range r.PrimaryKey.Fields {
		idMap[f.GoName] = struct{}{}
	}
	for _, f := range r.Fields {
		_, isId := idMap[f.GoName]
		sb := new(stringBuilder)
		sb.Append("`" + f.DbName + "`")
		sb.Append(f.DbType)
		if f.NotNull || isId {
			sb.Append("NOT NULL")
		}
		var fieldDefault string
		if f.Default != "" {
			fieldDefault = f.Default
		} else {
			if f.NotNull {
				if f.GoType == "string" {
					fieldDefault = "''"
				} else {
					fieldDefault = "0"
				}
			} else {
				fieldDefault = "NULL"
			}
		}
		if !isId && !f.AutoIncr {
			sb.Append("DEFAULT " + fieldDefault)
		}
		if f.AutoIncr {
			sb.Append("AUTO_INCREMENT")
		}
		if f.Comment != "" {
			sb.Append("COMMENT '" + f.Comment + "'")
		}
		c.P(0, "  ", sb.Get(), ",")
	}
	c.Empty()

	pks := make([]string, len(r.PrimaryKey.Fields))
	for idx, f := range r.PrimaryKey.Fields {
		pks[idx] = fmt.Sprintf("`%s`", f.DbName)
	}
	if len(r.UniqueKeys) > 0 || len(r.Indexes) > 0 {
		c.P(0, "  PRIMARY KEY (", strings.Join(pks, ","), "),")
	} else {
		c.P(0, "  PRIMARY KEY (", strings.Join(pks, ","), ")")
	}
	c.Empty()

	if len(r.UniqueKeys) > 0 {
		for i, key := range r.UniqueKeys {
			names := make([]string, len(key.Fields))
			vals := make([]string, len(key.Fields))
			for idx, f := range key.Fields {
				names[idx] = f.GoName
				vals[idx] = fmt.Sprintf("`%s`", f.DbName)
			}
			idx := fmt.Sprintf("`unique_%s_%s`(%s),",
				r.Table, strings.Join(names, "_"),
				strings.Join(vals, ","))
			if i == len(r.UniqueKeys)-1 && len(r.Indexes) == 0 {
				idx = idx[:len(idx)-1]
			}
			c.P(0, "  UNIQUE INDEX ", idx)
		}
		c.Empty()
	}
	if len(r.Indexes) > 0 {
		for i, key := range r.Indexes {
			names := make([]string, len(key.Fields))
			vals := make([]string, len(key.Fields))
			for idx, f := range key.Fields {
				names[idx] = f.GoName
				vals[idx] = fmt.Sprintf("`%s`", f.DbName)
			}
			idx := fmt.Sprintf("`index_%s_%s`(%s),",
				r.Table, strings.Join(names, "_"),
				strings.Join(vals, ","))
			if i == len(r.Indexes)-1 {
				idx = idx[:len(idx)-1]
			}
			c.P(0, "  INDEX ", idx)
		}
		c.Empty()
	}
	c.P(0, ") ENGINE=InnoDB COMMENT '", r.Comment, "';")
}

type stringBuilder struct {
	strs []string
}

func (sb *stringBuilder) Append(s string) {
	sb.strs = append(sb.strs, s)
}

func (sb *stringBuilder) Get() string {
	return strings.Join(sb.strs, " ")
}
