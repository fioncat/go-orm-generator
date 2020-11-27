package gosql

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/dbaccess"
	"github.com/fioncat/go-gendb/dbaccess/dbtypes"
	"github.com/fioncat/go-gendb/misc/col"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/parser/common"
	"github.com/fioncat/go-gendb/scanner"
	"github.com/fioncat/go-gendb/scanner/token"
)

type OperResult struct {
	Name    string
	Methods []Method
}

type Method struct {
	Type int
	Name string
	SQL  SQL

	Def string

	RetType    string
	RetSlice   bool
	RetPointer bool

	Imports []coder.Import

	QueryFields []QueryField

	importNames col.Set
}

type QueryField struct {
	Table string
	Field string
	Alias string
}

type SQL struct {
	String   string
	Prepares []string
	Replaces []string
}

func Parse(sr *scanner.GoResult, debug bool) ([]common.Result, error) {
	dir := filepath.Dir(sr.Path)
	for _, inter := range sr.Interfaces {
		var result OperResult
		err := parseInterface(&result, sr, inter, dir)
		if err != nil {
			return nil, errors.TraceFile(err, sr.Path)
		}
		if debug {
			fmt.Printf(">>>> interface: %s\n", inter.Name)
			for _, m := range result.Methods {
				fmt.Println(m.Def)
				term.Show(m)
			}
		}
	}

	return nil, nil
}

var (
	errEmptyTag      = errors.New("tag is empty")
	errEmptyIgnore   = errors.New("ignore is empty")
	errNameBadFormat = errors.New(`name bad format, should be "{db_name} {name}"`)
	errTypeBadFormat = errors.New(`type bad format, should be "{name} {type}"`)
	errEmptyTable    = errors.New("table is empty")
)

func parseStruct(r *common.StructResult, indent scanner.GoIndent) error {
	ignoreSet := col.NewSet(0)
	nameMap := make(map[string]string)
	typeMap := make(map[string]string)

	var tableNames []string

	for i, tag := range indent.Tags {
		if i == 0 {
			tableNames = tag.Args
			if len(tableNames) == 0 {
				return errors.Line(errEmptyTable, tag.Line)
			}
		}
		if len(tag.Args) <= 1 {
			return errors.Line(errEmptyTag, tag.Line)
		}
		master := tag.Args[0]
		args := tag.Args[1:]

		switch master {
		case "ignore":
			if len(args) == 0 {
				return errors.Line(errEmptyIgnore, tag.Line)
			}
			for _, ignore := range args {
				ignoreSet.Add(ignore)
			}

		case "name":
			if len(args) != 2 {
				return errors.Line(errNameBadFormat, tag.Line)
			}
			nameMap[args[0]] = args[1]

		case "type":
			if len(args) != 2 {
				return errors.Line(errTypeBadFormat, tag.Line)
			}
			typeMap[args[0]] = args[1]

		default:
			err := errors.Fmt(`unknown tag "%s"`, master)
			return errors.Line(err, tag.Line)
		}
	}

	if len(tableNames) == 0 {
		return errors.Line(errEmptyTable, indent.Line)
	}

	tables := make([]*dbtypes.Table, len(tableNames))
	for i, tableName := range tableNames {
		table, err := dbaccess.Desc(common.GetDBType(), tableName)
		if err != nil {
			return err
		}
		tables[i] = table
	}

	structName := indent.Token
	if strings.HasPrefix(structName, "*") {
		structName = strings.TrimLeft(structName, "*")
	}

	isNew := true
	var stc *coder.Struct
	for i, s := range r.Structs {
		if s.Name == structName {
			isNew = false
			stc = &r.Structs[i]
			break
		}
	}
	if stc == nil {
		stc = new(coder.Struct)
		stc.Name = structName
		if len(tables) == 1 {
			stc.Name = tables[0].Comment
		}
	}

	for _, table := range tables {
		for _, field := range table.Fields {
			if ignoreSet.Exists(field.Name) {
				continue
			}

			var goField coder.Field
			if name, ok := nameMap[field.Name]; ok {
				goField.Name = name
			} else {
				goField.Name = coder.GoName(field.Name)
			}

			if t, ok := typeMap[field.Name]; ok {
				goField.Type = t
			} else {
				goField.Type = field.Type
			}
			goField.Comment = field.Comment

			found := false
			for _, oldField := range stc.Fields {
				if oldField.Name == goField.Name {
					found = true
					break
				}
			}
			if found {
				continue
			}

			goField.AddTag("table", table.Name)
			goField.AddTag("field", field.Name)

			stc.Fields = append(stc.Fields, goField)
		}
	}

	if isNew {
		r.Structs = append(r.Structs, *stc)
	}

	return nil
}

type sqlMap map[string]scanner.SQL

func parseInterface(r *OperResult, sr *scanner.GoResult, inter scanner.GoInterface, dir string) error {
	r.Name = inter.Name

	sqlPaths := make([]string, 0, 1)
	for _, tag := range inter.Tags {
		if len(tag.Args) == 0 {
			return errors.Line(errEmptyTag, tag.Line)
		}
		sqlPaths = append(sqlPaths, tag.Args...)
	}

	if len(sqlPaths) == 0 {
		err := errors.Fmt("empty sql file for interface %s", inter.Name)
		return errors.Line(err, inter.Line)
	}

	sm := make(sqlMap)
	for _, path := range sqlPaths {
		path = filepath.Join(dir, path)
		sqlFile, err := scanner.SQLFile(path, false)
		if err != nil {
			return err
		}
		for _, sql := range sqlFile.Sqls {
			if _, ok := sm[sql.Name]; ok {
				err := errors.Fmt("sql is duplcate: %s", sql.Name)
				err = errors.Line(err, sql.Line)
				err = errors.TraceFile(err, sqlFile.Path)
				return err
			}
			sm[sql.Name] = sql
		}
	}

	for _, method := range inter.Methods {
		mr, err := parseMethod(method, sr, sm)
		if err != nil {
			return err
		}
		r.Methods = append(r.Methods, *mr)
	}
	return nil
}

func parseMethod(method scanner.GoMethod, sr *scanner.GoResult, sm sqlMap) (*Method, error) {
	p := newParser(method.Line, method.Tokens)
	r, err := p.method()
	if err != nil {
		return nil, err
	}
	r.Def = method.Def
	for _, importName := range r.importNames.Slice() {
		found := false
		for _, imp := range sr.Imports {
			if imp.Name == importName {
				r.Imports = append(r.Imports, coder.Import{
					Name: imp.Name,
					Path: imp.Path,
				})
				found = true
				break
			}
		}
		if !found {
			err := errors.Fmt(`can not find `+
				`import name "%s"`, importName)
			return nil, errors.Line(err, method.Line)
		}
	}
	sql, ok := sm[r.Name]
	if !ok {
		err := errors.Fmt(`can not find sql for method "%s"`, r.Name)
		return nil, errors.Line(err, method.Line)
	}
	p = newParser(sql.Line, sql.Tokens)
	err = p.sql(r, method, sql.SQL)
	if err != nil {
		return nil, errors.TraceFile(err, sql.Path)
	}

	return r, nil
}

type parser struct {
	line   int
	idx    int
	tokens []token.Token
}

func newParser(line int, tokens []token.Token) *parser {
	return &parser{
		line:   line,
		tokens: tokens,
	}
}

func (p *parser) pick() (*token.Token, bool) {
	if p.idx >= len(p.tokens) {
		return new(token.Token), false
	}
	return &p.tokens[p.idx], true
}

func (p *parser) last() (*token.Token, bool) {
	i := p.idx - 1
	if i < 0 {
		return new(token.Token), false
	}
	return &p.tokens[i], true
}

func (p *parser) next() (*token.Token, bool) {
	t, ok := p.pick()
	if ok {
		p.idx++
		return t, true
	}
	return new(token.Token), false
}

func (p *parser) method() (*Method, error) {
	m := new(Method)
	// First is method name
	nameToken, _ := p.next()
	if !nameToken.IsIndent() {
		err := errors.Fmt(`except method name INDENT, found "%s"`,
			nameToken.Flag)
		return nil, errors.Line(err, p.line)
	}
	m.Name = nameToken.Indent
	if m.Name == "" {
		err := errors.New("method name is empty")
		return nil, errors.Line(err, p.line)
	}

	// Params
	token, _ := p.next()
	if token.Flag != "(" {
		err := errors.Fmt(`except "(" , found "%s"`, token.Flag)
		return nil, errors.Line(err, p.line)
	}

	m.importNames = col.NewSet(0)
	var ok bool
	for {
		token, ok = p.pick()
		if !ok {
			err := errors.New("param list bad format")
			return nil, errors.Line(err, p.line)
		}
		if token.Flag == ")" {
			p.next()
			break
		}

		if token.Flag == "." {
			// The previous is import name
			pre, ok := p.last()
			if !ok {
				err := errors.Fmt(`the token before "." is missing`)
				return nil, errors.Line(err, p.line)
			}
			if !pre.IsIndent() {
				err := errors.Fmt(`except INDENT before`+
					`".", found: "%s"`, pre.Flag)
				return nil, errors.Line(err, p.line)
			}
			m.importNames.Add(pre.Indent)
		}

		p.next()
	}

	// Returns
	token, _ = p.next()
	if token.Flag != "(" {
		err := errors.Fmt(`method should have returns `+
			`starts with "(", found: "%s"`, token.Flag)
		return nil, errors.Line(err, p.line)
	}
	// The first return is target, parse it
	token, _ = p.pick()
	if !token.IsIndent() {
		switch token.Flag {
		case "[]":
			m.RetSlice = true
			p.next()
			token, _ = p.pick()
			if token.Flag == "*" {
				m.RetPointer = true
				p.next()
			}
		case "*":
			m.RetPointer = true
			p.next()

		default:
			err := errors.Fmt(`except "*" or "[]" in`+
				` returns start, found "%s"`, token.Flag)
			return nil, errors.Line(err, p.line)
		}
	}
	next, _ := p.pick()
	if next.Flag == "." {
		m.importNames.Add(token.Indent)
		m.RetType = token.Indent + "."
		p.next()
	}
	token, _ = p.next()
	if token.Indent == "" {
		err := errors.Fmt(`except INDENT for first `+
			`return, found: "%s"`, token.Flag)
		return nil, errors.Line(err, p.line)
	}
	m.RetType += token.Indent

	token, _ = p.next()
	if token.Flag != "," {
		err := errors.Fmt(`except ",", found: "%s"`, token.Flag)
		return nil, errors.Line(err, p.line)
	}

	token, _ = p.next()
	if token.Indent != "error" {
		err := errors.Fmt(`the second return must `+
			`be "error", found: "%s"`, token.Flag)
		return nil, errors.Line(err, p.line)
	}

	token, _ = p.next()
	if token.Flag != ")" {
		err := errors.Fmt(`the end of method must `+
			` be ")", found: "%s"`, token.Flag)
		return nil, errors.Line(err, p.line)
	}

	return m, nil
}
