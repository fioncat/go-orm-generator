package orm

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/misc/errors"
)

type File struct {
	Db string

	SqlPath string

	Results []*Result
}

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
	AutoIncr bool

	Comment string

	GoName string
	GoType string

	DbName string
	DbType string
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

func Parse(gfile *golang.File) (*File, error) {
	file, err := parse(gfile)
	if err != nil {
		err = errors.Trace(gfile.Path, err)
		err = errors.OnCompile(gfile.Path, gfile.Lines, err)
		return nil, err
	}
	return file, nil
}

func parse(gfile *golang.File) (*File, error) {
	file := new(File)
	for _, opt := range gfile.Options {
		switch opt.Key {
		case "import_table":
			var table string
			var name string
			tmp := strings.Split(opt.Value, ",")
			switch len(tmp) {
			case 1:
				table = tmp[0]
				name = coder.GoName(table)

			case 2:
				table = tmp[0]
				name = tmp[1]

			default:
				return nil, opt.FmtError(`import `+
					`table "%s" is bad format`, opt.Value)
			}
			r, err := fromDatabase(table, name)
			if err != nil {
				return nil, opt.FmtError(err.Error())
			}
			file.Results = append(file.Results, r)

		case "db":
			file.Db = opt.Value

		case "sql_path":
			file.SqlPath = opt.Value
		}
	}

	for _, stc := range gfile.Structs {
		r, err := parseStruct(stc)
		if err != nil {
			return nil, err
		}
		file.Results = append(file.Results, r)
	}

	if len(file.Results) == 0 {
		return nil, fmt.Errorf("no orm to generate")
	}

	return file, nil
}

func parseStruct(stc *golang.Struct) (*Result, error) {
	r := newResult(stc.Name, len(stc.Fields))
	r.line = stc.Line
	r.Comment = stc.Comment
	err := parseTags(r, nil, stc.Tags, parseStructOption)
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
		err = parseTags(r, rf, gf.Tags, parseFieldOption)
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

type parseAction func(r *Result, f *Field, opt *base.Option) error

func parseTags(r *Result, f *Field, tags []*base.Tag, parse parseAction) error {
	for _, tag := range tags {
		if tag.Name != "orm" {
			continue
		}
		for _, opt := range tag.Options {
			opt := opt
			err := parse(r, f, &opt)
			if err != nil {
				return opt.Trace(err)
			}
		}
	}
	return nil
}

func parseStructOption(r *Result, _ *Field, opt *base.Option) error {
	switch opt.Key {
	case "table":
		r.Table = opt.Value

	case "name":
		r.Name = opt.Value

	case "index":
		arrs, err := parse2larray(opt.Value)
		if err != nil {
			return err
		}
		for _, names := range arrs {
			r.addIdx(opt.Line, names)
		}

	case "unique":
		arrs, err := parse2larray(opt.Value)
		if err != nil {
			return err
		}
		for _, names := range arrs {
			r.addUnique(opt.Line, names)
		}

	case "pk":
		arr, err := parse1larray(opt.Value)
		if err != nil {
			return err
		}
		for _, name := range arr {
			r.addPk(opt.Line, name)
		}

	default:
		return fmt.Errorf(`unknown option "%s"`, opt.Key)
	}
	return nil
}

func parseFieldOption(r *Result, f *Field, opt *base.Option) error {
	switch opt.Key {
	case "flags":
		flags, err := parse1larray(opt.Value)
		if err != nil {
			return err
		}
		for _, flag := range flags {
			switch flag {
			case "primary":
				r.addPk(opt.Line, f.GoName)

			case "auto-incr":
				f.AutoIncr = true

			case "index":
				r.addIdx(opt.Line, []string{f.GoName})

			case "unique":
				r.addUnique(opt.Line, []string{f.GoName})

			default:
				return fmt.Errorf(`unknown flag "%s"`, flag)
			}
		}

	case "type":
		f.DbType = strings.ToUpper(opt.Value)

	case "name":
		f.DbName = opt.Value

	default:
		return fmt.Errorf(`unknown option "%s"`, opt.Key)
	}
	return nil
}

var arrTokens = []token.Token{
	token.LBRACK, token.RBRACK,
	token.COMMA,
}

func parse2larray(val string) ([][]string, error) {
	s := token.NewScanner(val, arrTokens)
	var arrs [][]string
	var e token.Element
	for s.Next(&e) {
		if e.Token == token.COMMA {
			continue
		}
		if e.Token != token.LBRACK {
			return nil, e.NotMatch("LBRACK")
		}
		var names []string
		for {
			ok := s.Next(&e)
			if !ok {
				return nil, s.EarlyEnd("RBRACK")
			}
			if e.Token == token.RBRACK {
				break
			}
			if e.Token == token.COMMA {
				continue
			}
			names = append(names, e.Get())
		}
		arrs = append(arrs, names)
	}
	return arrs, nil
}

func parse1larray(val string) ([]string, error) {
	s := token.NewScanner(val, arrTokens)
	var arr []string
	var e token.Element
	ok := s.Next(&e)
	if !ok {
		return nil, s.EarlyEnd("LBRACK")
	}
	if e.Token != token.LBRACK {
		return nil, e.NotMatch("LBRACK")
	}

	for {
		ok = s.Next(&e)
		if !ok {
			return nil, s.EarlyEnd("RBRACK")
		}
		if e.Token == token.RBRACK {
			break
		}
		if e.Token == token.COMMA {
			continue
		}
		arr = append(arr, e.Get())
	}
	return arr, nil
}
