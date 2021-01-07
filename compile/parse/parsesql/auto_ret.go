package parsesql

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/generate/coder"
)

var tableCache = make(map[string]rdb.Table)

func genRetStruct(m *Method) error {
	if len(m.QueryTables) == 0 {
		return fmt.Errorf(`can not find table, ` +
			`please check your sql statement`)
	}
	retName, err := extractRetType(m)
	if err != nil {
		return err
	}

	nameMap := make(map[string]string, len(m.QueryTables))
	aliasMap := make(map[string]string, len(m.QueryTables))

	for _, table := range m.QueryTables {
		nameMap[table.Name] = table.Name
		if table.Alias != "" {
			aliasMap[table.Alias] = table.Name
		}
	}

	err = rdb.MustInit()
	if err != nil {
		return err
	}

	goStruct := new(coder.Struct)
	goStruct.Name = retName
	goStruct.Comment = fmt.Sprintf("%s is an auto-generated "+
		"return type for method %s", retName, m.Name)
	goStruct.Fields = make([]coder.Field, len(m.QueryFields))

	var tableName string
	var ok bool
	for idx, field := range m.QueryFields {
		// call database to get table desc.
		if tableName, ok = nameMap[field.Table]; !ok {
			tableName = aliasMap[field.Table]
		}
		if tableName == "" {
			return fmt.Errorf(`can not find `+
				`table "%s", please check your sql statement`,
				field.Table)
		}

		table := tableCache[tableName]
		if table == nil {
			table, err = rdb.Get().Desc(tableName)
			if err != nil {
				return err
			}

			tableCache[tableName] = table
		}

		descField := table.Field(field.Field)
		if descField == nil {
			return fmt.Errorf(`can not find field "%s" in `+
				`table "%s" from remote database`, field.Field, tableName)
		}

		var goField coder.Field
		if field.Alias != "" {
			goField.Name = field.Alias
		} else {
			goField.Name = coder.GoName(field.Field)
		}
		goField.Type = rdb.Get().GoType(descField.GetType())
		goField.Comment = descField.GetComment()
		goField.AddTag("table", tableName)
		goField.AddTag("field", field.Field)

		goStruct.Fields[idx] = goField
	}

	m.RetStruct = goStruct
	return nil
}

func extractRetType(m *Method) (string, error) {
	ret := m.RetType
	if m.RetSlice {
		ret = strings.TrimLeft(ret, token.BRACKS.Get())
	}
	if m.RetPointer {
		ret = strings.TrimLeft(ret, token.MUL.Get())
	}

	if token.PERIOD.Contains(ret) {
		return "", fmt.Errorf("do not support "+
			"external struct: %s", ret)
	}

	if coder.IsSimpleType(ret) {
		return "", fmt.Errorf("do not support Go "+
			"builtin type: %s", ret)
	}

	return ret, nil
}
