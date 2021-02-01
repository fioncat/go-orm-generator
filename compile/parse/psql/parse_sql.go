package psql

import (
	"strings"

	"github.com/fioncat/go-gendb/compile/scan/sgo"
	"github.com/fioncat/go-gendb/compile/scan/ssql"
	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/iter"
)

// Statement only parses SQL statements, and the returned
// Method data does not contain data required for code
// generation. This function is generally used for operations
// on sql statements (not involving code generation).
func Statement(stat *ssql.Statement) (*Method, error) {
	if len(stat.Tokens) == 0 {
		return nil, errors.NewComp(stat.Path, stat.LineNum, "empty sql")
	}
	sm := new(sgo.Method)
	m := new(Method)
	if token.SQL_SELECT.Match(stat.Tokens[0]) {
		sm.Flag.Tags = []sgo.Tag{
			{Args: []string{"auto-ret"}},
		}
	} else {
		m.RetType = "sql.Result"
	}
	err := _sql("non-go", *stat, m, sm)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// parse sql tokens.
// If it is an execution sql statement, only parse return type.
// If it is a query, parse its SELECT cluase to get QueryFields,
// if user configured <auto-ret> option, need to parse all tables
// for the sql statement.
// TODO: At present, we do not optimize the analysis of subqueries.
// The inline table of the subquery will be returned as a normal
// table. This may cause ambiguity, which needs to be optimized later.
func _sql(goPath string, sql ssql.Statement, method *Method, sm *sgo.Method) error {

	var err error
	if !sql.IsDynamic {
		phs, err := ssql.DoPlaceholders(sql.Path, sql.LineNum, sql.Origin)
		if err != nil {
			return err
		}

		method.SQL = SQL{
			Contant:  phs.SQL,
			Prepares: phs.Prepares,
			Replaces: phs.Replaces,
		}
	} else {
		method.IsDynamic = true
		method.DynamicParts, err = dynamic(sql.Path, sql.LineNum, sql.Origin)
		if err != nil {
			return err
		}
	}

	ef := errors.NewParseFactory(sql.Path, sql.LineNum)
	iter := iter.New(sql.Tokens)
	var tk token.Token

	ok := iter.Next(&tk)
	if !ok {
		return ef.EarlyEnd("KEY")
	}

	switch tk.Get() {
	case token.SQL_INSERT.Get(),
		token.SQL_DELETE.Get(),
		token.SQL_UPDATE.Get():
		// non-query, just parse result, then returns.
		method.IsExec = true

		// If manually specify the method return
		// type, parse it and assign to Type.
		var tagLine int
		var execArg = defaultExecType
		if len(sm.Tags) >= 1 {
			tag := sm.Tags[0]
			if len(tag.Args) != 1 {
				return errors.NewComp(goPath, tag.Line,
					"exec tag bad format")
			}
			tagLine = tag.Line
			execArg = tag.Args[0]
		}

		method.Type = parseExecType(execArg)
		if method.Type == "" {
			return errors.NewComp(goPath, tagLine,
				`unknown exec return type "%s"`, execArg)
		}

		// check return type
		switch method.Type {
		case sqlExecAffect:
			fallthrough
		case sqlExecLastid:
			if method.RetType != "int64" {
				return errors.NewComp(goPath, sm.Line,
					`must return "int64", found: "%s"`, method.RetType)
			}

		case sqlExecResult:
			if method.RetType != "sql.Result" {
				return errors.NewComp(goPath, sm.Line,
					`must return "sql.Result", found: "%s"`, method.RetType)
			}
		}
		return nil

	case token.SQL_SELECT.Get():
		if method.RetSlice {
			method.Type = sqlQueryMany
		} else {
			method.Type = sqlQueryOne
		}
		method.IsExec = false

	default:
		return ef.MismatchS(0, "SELECT/INSERT/UPDATE/DELETE", tk)
	}

	// At this point, exec sql has been returned.
	// So it can only be a query.
	err = _select(iter, method, ef)
	if err != nil {
		return err
	}

	// Check out whether set auto-ret
	for _, tag := range sm.Tags {
		for _, arg := range tag.Args {
			if arg == "auto-ret" {
				method.IsAutoRet = true
				break
			}
		}
	}

	// If it is auto-ret, need to parse the tables
	if method.IsAutoRet {
		iter.Reset()
		return _tables(iter, method, ef)
	}
	return nil
}

const defaultExecType = "result"

func parseExecType(arg string) string {
	switch arg {
	case "result":
		return sqlExecResult
	case "lastid":
		return sqlExecLastid
	case "affect":
		return sqlExecAffect
	}
	return ""
}

// parse SELECT clause
func _select(sqlIter *iter.Iter, m *Method, ef *errors.ParseErrorFactory) error {
	// Extract SELECT clause
	tks := make([]token.Token, 0)
	var tk token.Token
	for {
		ok := sqlIter.Next(&tk)
		if !ok {
			return ef.EarlyEnd(token.SQL_FROM.Get())
		}
		if token.SQL_FROM.Match(tk) {
			break
		}
		tks = append(tks, tk)
	}
	if len(tks) == 0 {
		return ef.Empty(0, "SELECT clause")
	}

	sqlIter = iter.New(tks)
	var idx int
	var err error
	for {
		idx = sqlIter.Pick(&tk)
		if idx < 0 {
			return nil
		}

		var field QueryField
		if !tk.IsIndent() {
			switch {
			case token.SQL_IFNULL.Match(tk):
				sqlIter.Next(nil)
				err = _ifnull(sqlIter, &field, ef)

			case token.SQL_COUNT.Match(tk):
				sqlIter.Next(nil)
				err = _count(sqlIter, &field)

			default:
				return ef.MismatchS(idx,
					"IFNULL/COUNT/INDENT", tk)
			}
		} else {
			err = _field(sqlIter, &field, ef)
		}
		if err != nil {
			return err
		}

		m.QueryFields = append(m.QueryFields, field)
	}
}

// parse IFNULL function
func _ifnull(iter *iter.Iter, field *QueryField, ef *errors.ParseErrorFactory) error {
	var tk token.Token
	idx := iter.NextP(&tk)
	if idx < 0 {
		return ef.EarlyEnd(token.LPAREN.Get())
	}
	if !token.LPAREN.Match(tk) {
		return ef.Mismatch(idx, token.LPAREN, tk)
	}

	var next token.Token
	for {
		idx = iter.NextP(&tk)
		if idx < 0 {
			return ef.EarlyEnd(token.RPAREN.Get())
		}
		if token.RPAREN.Match(tk) {
			break
		}

		if tk.IsIndent() && field.Field == "" {
			idx = iter.Pick(&next)
			if idx < 0 {
				return ef.EarlyEnd("INDENT or .")
			}
			if token.PERIOD.Match(next) {
				field.Table = tk.Get()
				iter.Next(nil)
				idx = iter.NextP(&tk)
				if idx < 0 {
					return ef.EarlyEnd("INDENT")
				}
				field.Field = tk.Get()
				continue
			}
			field.Field = tk.Get()
		}
	}
	if field.Field == "" {
		return ef.Empty(idx, "IFNULL")
	}
	_alias(iter, field)
	return nil
}

// parse COUNT
// TODO: The current COUNT processing is too rough:
// as long as COUNT is found, it is considered to be
// a statistical sql, and all the following fields are
// ignored. More detailed COUNT parsing logic should
// be provided later.
func _count(iter *iter.Iter, field *QueryField) error {
	field.Field = "count"
	field.IsCount = true
	for {
		ok := iter.Next(nil)
		if !ok {
			return nil
		}
	}
}

// single field: [table.]field [[AS] alias]
func _field(iter *iter.Iter, field *QueryField, ef *errors.ParseErrorFactory) error {
	var tk token.Token
	idx := iter.NextP(&tk)
	if idx < 0 {
		return ef.EarlyEnd("INDENT")
	}
	if !tk.IsIndent() {
		return ef.MismatchS(idx, "INDENT", tk)
	}

	var next token.Token
	iter.Pick(&next)
	if token.PERIOD.Match(next) {
		field.Table = tk.Get()
		iter.Next(nil)
		idx = iter.NextP(&tk)
		if idx < 0 {
			return ef.EarlyEnd("INDENT")
		}
		if !tk.IsIndent() {
			return ef.MismatchS(idx, "INDENT", tk)
		}
	}
	field.Field = tk.Get()

	_alias(iter, field)
	return nil
}

// alias <previous> [[AS] alias]
func _alias(iter *iter.Iter, f *QueryField) {
	var tk token.Token
	var idx int
	for {
		idx = iter.Pick(&tk)
		if idx < 0 || token.COMMA.Match(tk) {
			var last token.Token
			iter.Previous(&last)
			if last.IsIndent() && last.Get() != f.Field {
				f.Alias = last.Get()
			}
			iter.Next(nil)
			return
		}
		iter.Next(nil)
	}
}

func _tables(iter *iter.Iter, m *Method, ef *errors.ParseErrorFactory) error {
	var tk token.Token
	var idx int
	for {
		idx = iter.NextP(&tk)
		if idx < 0 {
			return nil
		}

		if token.SQL_FROM.Match(tk) ||
			token.SQL_JOIN.Match(tk) {
			// FROM or JOIN
			var table QueryTable

			idx = iter.NextP(&tk)
			if idx < 0 {
				return ef.EarlyEnd("INDENT")
			}

			// Must be INDENT
			if !tk.IsIndent() {
				return ef.MismatchS(idx, "INDENT", tk)
			}

			table.Name = tk.Get()

			// check alias
			for {
				idx = iter.Pick(&tk)
				if idx < 0 {
					break
				}
				if token.SQL_AS.Match(tk) {
					iter.Next(nil)
					continue
				}
				if !tk.IsIndent() {
					break
				}
				table.Alias = tk.Get()
				iter.Next(nil)
				break
			}

			m.QueryTables = append(m.QueryTables, table)
		}
	}
}

func dynamic(path string, line int, sql string) ([]*DynamicPart, error) {
	scanParts, err := ssql.DoDynamic(path, line, sql)
	if err != nil {
		return nil, err
	}

	parts := make([]*DynamicPart, len(scanParts))
	for idx, scanPart := range scanParts {

		part := &DynamicPart{
			SQL: SQL{
				Contant:  scanPart.SQL.SQL,
				Prepares: scanPart.SQL.Prepares,
				Replaces: scanPart.SQL.Replaces,
			},
		}

		switch scanPart.Type {
		case ssql.DynamicTypeConst:
			part.Type = DynamicTypeConst

		case ssql.DynamicTypeIf:
			part.IfCond = scanPart.Cond
			part.Type = DynamicTypeIf

		case ssql.DynamicTypeFor:
			part.Type = DynamicTypeFor
			tmp := strings.Split(scanPart.Cond, ":")
			switch len(tmp) {
			case 1:
				part.ForSlice = tmp[0]

			case 3:
				part.ForJoin = tmp[2]
				part.ForJoin = strings.ReplaceAll(
					part.ForJoin, "'", "")
				fallthrough

			case 2:
				part.ForEle = tmp[0]
				part.ForSlice = tmp[1]

				foundEle := checkContains(
					part.SQL.Prepares, part.ForEle)
				if !foundEle {
					foundEle = checkContains(
						part.SQL.Replaces, part.ForEle)
				}
				if !foundEle {
					part.ForEle = ""
				}

			default:
				return nil, errors.NewComp(path, line,
					"for condition bad format: %s", scanPart.Cond)
			}

		default:
			// Never trigger, prevent
			return nil, errors.NewComp(path, line,
				"unknown type code=%d", scanPart.Type)
		}

		parts[idx] = part
	}

	return parts, nil
}

func checkContains(vs []string, tar string) bool {
	for _, v := range vs {
		if strings.Contains(v, tar) {
			return true
		}
	}
	return false
}
