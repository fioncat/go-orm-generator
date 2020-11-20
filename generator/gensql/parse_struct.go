package gensql

import (
	"fmt"
	"path/filepath"

	"github.com/fioncat/go-gendb/dbaccess"
	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/generator/coder"
	"github.com/fioncat/go-gendb/misc/col"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
)

func parseStruct(task *generator.Task, ts generator.TaskStruct) (*generator.File, error) {
	if len(ts.Args) == 0 {
		err := errors.Fmt("missing table")
		return nil, errors.Line(task.Path, ts.Line, err)
	}
	tableName := ts.Args[0]

	ignoreFields := col.NewSet(0)
	typeMap := make(map[string]string)
	nameMap := make(map[string]string)

	for _, opt := range ts.Options {
		switch opt.Tag {
		case "ignore":
			if len(opt.Args) == 0 {
				err := errors.Fmt("ignore missing fieldName")
				return nil, errors.Line(task.Path, opt.Line, err)
			}
			for _, f := range opt.Args {
				ignoreFields.Add(f)
			}

		case "type":
			if len(opt.Args) != 2 {
				err := errors.Fmt("type format error: " +
					"should be {name} {type}")
				return nil, errors.Line(task.Path, opt.Line, err)
			}
			f := opt.Args[0]
			fType := opt.Args[1]
			typeMap[f] = fType

		case "name":
			if len(opt.Args) != 2 {
				err := errors.Fmt("name format error: " +
					"should be {name} {type}")
				return nil, errors.Line(task.Path, opt.Line, err)
			}
			fieldName := opt.Args[0]
			goName := opt.Args[1]
			nameMap[fieldName] = goName

		default:
			err := errors.Fmt(`unknown option "%s" for struct`, opt.Tag)
			return nil, errors.Line(task.Path, opt.Line, err)
		}
	}
	log.Infof("parse struct %s, table=%s ignore=%v,"+
		" typeMap=%v, nameMap=%v", ts.Name, tableName,
		ignoreFields.Slice(), typeMap, nameMap)

	descTable, err := dbaccess.Desc(task.Type, tableName)
	if err != nil {
		return nil, err
	}

	s := new(coder.Struct)
	s.Name = ts.Name
	s.Comment = descTable.Comment
	for _, descField := range descTable.Fields {
		if ignoreFields.Exists(descField.Name) {
			continue
		}
		var field coder.Field
		field.Comment = descField.Comment
		if customerName, ok := nameMap[descField.Name]; ok {
			field.Name = customerName
		} else {
			field.Name = coder.GoName(descField.Name)
		}

		if customerType, ok := typeMap[descField.Name]; ok {
			field.Type = customerType
		} else {
			field.Type = descField.Type
		}
		field.AddTag("table", descTable.Name)
		field.AddTag("field", descField.Name)
		s.Fields = append(s.Fields, field)
	}

	dir := filepath.Dir(task.Path)
	file := new(generator.File)
	filename := fmt.Sprintf("zz_generated_struct_%s.go", s.Name)
	file.Path = filepath.Join(dir, filename)
	file.Result = s

	return file, nil
}
