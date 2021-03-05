package sql

import (
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/sql"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/link/internal/refs"
)

type Linker struct{}

type conf struct {
	dbUse   string
	runPath string
	runName string
}

func defaultConf() *conf {
	return &conf{
		dbUse:   "db",
		runPath: "github.com/finocat/go-gendb/api/sql/run",
		runName: "run",
	}
}

func (*Linker) Do(file *golang.File) ([]coder.Target, error) {
	conf := defaultConf()
	for _, opt := range file.Options {
		switch opt.Key {
		case "db_use":
			conf.dbUse = opt.Value

		case "run_path":
			conf.runPath = opt.Value

		case "run_name":
			conf.runName = opt.Value
		}
	}

	ts := make([]coder.Target, len(file.Interfaces))
	for idx, inter := range file.Interfaces {
		t, err := createTarget(inter, file, conf)
		if err != nil {
			return nil, err
		}
		ts[idx] = t
	}
	return ts, nil
}

func createTarget(inter *golang.Interface, file *golang.File, c *conf) (
	coder.Target, error,
) {
	if inter.Tag.Name == "" {
		return nil, inter.FmtError("tag miss name")
	}
	name := inter.Tag.Name

	sqlM0 := make(map[string]*sql.Method)
	sqlM1 := make(map[string]*sql.Method)

	var sqlPaths []string
	for _, opt := range inter.Tag.Options {
		if opt.Key != "" && opt.Key != "file" {
			continue
		}
		if opt.Value == "" {
			continue
		}
		sqlPaths = append(sqlPaths, opt.Value)
	}
	if len(sqlPaths) == 0 {
		return nil, inter.FmtError("missing sql file")
	}

	for _, sqlPath := range sqlPaths {
		v, err := refs.Import(file.Path, sqlPath, "sql")
		if err != nil {
			return nil, err
		}
		sqlFile := v.(*sql.File)
		for _, sqlM := range sqlFile.Methods {
			if sqlM.Inter == name {
				sqlM0[sqlM.Name] = sqlM
				continue
			}
			if sqlM.Inter == "" {
				sqlM1[sqlM.Name] = sqlM
				continue
			}
		}
	}

	t := new(target)
	t.inter = inter
	t.file = file
	t.conf = c
	t.name = name
	t.ms = make([]*method, len(inter.Methods))

	for idx, gom := range inter.Methods {
		var sqlm *sql.Method
		sqlm = sqlM0[gom.Name]
		if sqlm == nil {
			sqlm = sqlM1[gom.Name]
		}
		if sqlm == nil {
			return nil, gom.FmtError(`can not find`+
				` method "%s" from sql file`, gom.Name)
		}

		m := &method{
			sqlm: sqlm,
			gom:  gom,
		}

		isAutoRet := false
		for _, tag := range gom.Tags {
			switch tag.Name {
			case "auto-ret":
				isAutoRet = true

			case "last-id":
				m.lastId = true
			}
		}
		if isAutoRet {
			ret, err := createRet(gom, sqlm)
			if err != nil {
				return nil, err
			}
			ret.m = m
			t.rets = append(t.rets, ret)
		}
		t.ms[idx] = m
	}

	return t, nil
}

func createRet(gom *golang.Method, sqlm *sql.Method) (*ret, error) {
	err := rdb.MustInit()
	if err != nil {
		return nil, gom.FmtError("auto-ret " +
			"require database connection")
	}

	name, err := extractRetName(gom)
	if err != nil {
		return nil, err
	}

	ret := new(ret)
	ret.name = name
	ret.fs = make([]*field, len(sqlm.Fields))
	for idx, f := range sqlm.Fields {
		fieldName := getFieldName(f)
		tableName := f.Table

		table, err := rdb.Get().Desc(tableName)
		if err != nil {
			return nil, gom.FmtError("Desc "+
				"table failed: %v", err)
		}
		dbField := table.Field(f.Name)
		if dbField == nil {
			return nil, sqlm.FmtError(`can not`+
				` find field "%s" from table "%s"`+
				` in the remote database.`,
				f.Name, f.Table)
		}

		retField := new(field)
		retField.name = fieldName
		retField.goType = dbField.GetType()
		retField.dbTable = f.Table
		retField.dbField = f.Name

		ret.fs[idx] = retField
	}

	return ret, nil
}

func extractRetName(gom *golang.Method) (string, error) {
	name := gom.RetType
	if gom.RetSlice {
		name = strings.TrimPrefix(name, "[]")
	}
	if gom.RetPointer {
		name = strings.TrimPrefix(name, "*")
	}
	if strings.Contains(name, ".") {
		return "", gom.FmtError("auto-ret's "+
			"return type must be local, found: %s", name)
	}
	return name, nil
}

func getFieldName(f *sql.QueryField) string {
	if f.Alias != "" {
		return f.Alias
	}
	return coder.GoName(f.Name)
}
