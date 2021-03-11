package orm

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/misc/errors"
)

type Result struct {
	Name  string
	Table string

	Comment string

	Fields []*Field

	PrimaryKey *Index

	UniqueKeys []*Index

	Indexes []*Index

	line int

	dbMap map[string]*Field
	goMap map[string]*Field

	idxNames [][]string
	idxLines []int

	uqNames [][]string
	uqLines []int

	primayNames []string
	pkLines     []int
}

type Field struct {
	NotNull  bool
	AutoIncr bool

	Comment string

	GoName string
	GoType string

	DbName string
	DbType string

	Default string
}

type Index struct {
	Fields []*Field
}

func newResult(name string, cap int) *Result {
	r := new(Result)
	r.Name = coder.GoName(name)
	r.Fields = make([]*Field, 0, cap)
	r.dbMap = make(map[string]*Field, cap)
	r.goMap = make(map[string]*Field, cap)
	return r
}

func fromDatabase(tableName, name string) (*Result, error) {
	table, err := rdb.Get().Desc(tableName)
	if err != nil {
		return nil, err
	}
	fieldNames := table.FieldNames()
	if name == "" {
		name = coder.GoName(tableName)
	}
	r := newResult(name, len(fieldNames))
	for _, fieldName := range fieldNames {
		field := table.Field(fieldName)
		if field.IsPrimaryKey() {
			r.addPk(0, fieldName)
		}

		rf := new(Field)
		rf.DbName = field.GetName()
		rf.DbType = field.GetType()
		rf.DbType = strings.ToUpper(rf.DbType)
		rf.GoName = coder.GoName(rf.DbName)
		rf.GoType = rdb.Get().GoType(rf.DbType)
		rf.AutoIncr = field.IsAutoIncr()

		r.addField(rf)
	}
	err = r.parseKeys()
	if err != nil {
		// Never trigger, prevent
		return nil, fmt.Errorf("[parseKeyInImportDb] %v", err)
	}

	return r, nil
}

func (r *Result) addIdx(line int, names []string) {
	r.idxLines = append(r.idxLines, line)
	r.idxNames = append(r.idxNames, names)
}

func (r *Result) addUnique(line int, names []string) {
	r.uqLines = append(r.uqLines, line)
	r.uqNames = append(r.uqNames, names)
}

func (r *Result) addPk(line int, name string) {
	r.pkLines = append(r.pkLines, line)
	r.primayNames = append(r.primayNames, name)
}

func (r *Result) addField(f *Field) {
	r.Fields = append(r.Fields, f)
	r.dbMap[f.DbName] = f
	r.goMap[f.GoName] = f
}

func (r *Result) fieldByName(name string) *Field {
	f := r.goMap[name]
	if f != nil {
		return f
	}
	return r.dbMap[name]
}

func (r *Result) getIndexes(keyNames [][]string, lines []int) (
	[]*Index, error,
) {
	idxes := make([]*Index, len(keyNames))
	for i, names := range keyNames {
		idx := new(Index)
		idx.Fields = make([]*Field, len(names))
		for j, name := range names {
			f := r.fieldByName(name)
			if f == nil {
				line := lines[i]
				return nil, errors.TraceFmt(line,
					`can not find field "%s"`, name)
			}
			idx.Fields[j] = f
		}
		idxes[i] = idx
	}
	return idxes, nil
}

func (r *Result) parseKeys() error {
	pk := new(Index)
	for idx, pkName := range r.primayNames {
		f := r.fieldByName(pkName)
		if f == nil {
			line := r.pkLines[idx]
			return errors.TraceFmt(line,
				`can not find field "%s"`, pkName)
		}
		pk.Fields = append(pk.Fields, f)
	}
	if len(pk.Fields) == 0 {
		return errors.TraceFmt(r.line,
			"missing primary key")
	}
	r.PrimaryKey = pk

	var err error
	r.UniqueKeys, err = r.getIndexes(r.uqNames, r.uqLines)
	if err != nil {
		return err
	}

	r.Indexes, err = r.getIndexes(r.idxNames, r.idxLines)
	if err != nil {
		return err
	}

	return nil
}

func Parse(gfile *golang.File) ([]*Result, error) {
	rs, err := parse(gfile)
	if err != nil {
		err = errors.Trace(gfile.Path, err)
		err = errors.OnCompile(gfile.Path, gfile.Lines, err)
		return nil, err
	}
	return rs, nil
}

func parse(gfile *golang.File) ([]*Result, error) {
	rs := make([]*Result, len(gfile.Structs))
	for idx, stc := range gfile.Structs {
		r, err := parseStruct(stc)
		if err != nil {
			return nil, err
		}
		rs[idx] = r
	}

	for _, opt := range gfile.Options {
		if opt.Key != "import_table" {
			continue
		}
		arrs, err := base.Arr2(opt.Value)
		if err != nil {
			return nil, opt.Trace(err)
		}
		for _, arr := range arrs {
			var table string
			var name string
			switch len(arr) {
			case 1:
				table = arr[0]
				name = coder.GoName(table)

			case 2:
				table = arr[0]
				name = arr[1]

			default:
				return nil, opt.FmtError(`import_table "%s" is bad format`)
			}
			r, err := fromDatabase(table, name)
			if err != nil {
				return nil, opt.Trace(err)
			}
			rs = append(rs, r)
		}
	}

	if len(rs) == 0 {
		return nil, fmt.Errorf("no orm to generate")
	}

	return rs, nil
}

var structOptionMap = map[string]base.DecodeOptionFunc{
	"table": func(_ int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		r.Table = val
		return nil
	},

	"name": func(_ int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		r.Name = val
		return nil
	},

	"primary": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		arr, err := base.Arr1(val)
		if err != nil {
			return err
		}
		for _, name := range arr {
			r.addPk(line, name)
		}
		return nil
	},

	"index": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		arrs, err := base.Arr2(val)
		if err != nil {
			return err
		}
		for _, arr := range arrs {
			r.addIdx(line, arr)
		}
		return nil
	},

	"unique": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		arrs, err := base.Arr2(val)
		if err != nil {
			return err
		}
		for _, arr := range arrs {
			r.addUnique(line, arr)
		}
		return nil
	},
}

var fieldOptionMap = map[string]base.DecodeOptionFunc{
	"flags": func(line int, val string, vs []interface{}) error {
		r := vs[0].(*Result)
		f := vs[1].(*Field)
		flags, err := base.Arr1(val)
		if err != nil {
			return err
		}
		for _, flag := range flags {
			switch flag {
			case "auto-incr":
				f.AutoIncr = true

			case "primary":
				r.addPk(line, f.GoName)

			case "notnull":
				f.NotNull = true

			case "index":
				r.addIdx(line, []string{f.GoName})

			case "unique":
				r.addUnique(line, []string{f.GoName})

			default:
				return fmt.Errorf(`unknown flag "%s"`, flag)
			}
		}
		return nil
	},

	"name": func(line int, val string, vs []interface{}) error {
		f := vs[1].(*Field)
		f.DbName = val
		return nil
	},

	"type": func(line int, val string, vs []interface{}) error {
		f := vs[1].(*Field)
		f.DbType = val
		return nil
	},

	"default": func(line int, val string, vs []interface{}) error {
		f := vs[1].(*Field)
		f.Default = val
		return nil
	},
}

func parseStruct(stc *golang.Struct) (*Result, error) {
	r := newResult(stc.Name, len(stc.Fields))
	r.line = stc.Line
	r.Comment = stc.Comment
	err := base.DecodeTags(stc.Tags, "orm", structOptionMap, r)
	if err != nil {
		return nil, err
	}
	if r.Table == "" {
		r.Table = coder.DbName(r.Name)
	}

	for _, gf := range stc.Fields {
		rf := new(Field)
		rf.GoName = gf.Name
		rf.GoType = gf.Type
		rf.Comment = gf.Comment
		err = base.DecodeTags(gf.Tags, "orm", fieldOptionMap, r, rf)
		if err != nil {
			return nil, err
		}
		if rf.DbName == "" {
			rf.DbName = coder.DbName(rf.GoName)
		}
		if rf.DbType == "" {
			rf.DbType = rdb.Get().SqlType(rf.GoType)
		}
		r.addField(rf)
	}

	err = r.parseKeys()
	if err != nil {
		return nil, err
	}

	return r, nil
}
