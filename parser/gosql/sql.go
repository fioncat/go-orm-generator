package gosql

import (
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/scanner"
	"github.com/fioncat/go-gendb/scanner/token"
)

const (
	sqlExecAffect = "Exec_Affect"
	sqlExecLastid = "Exec_lastId"
	sqlExecResult = "Exec_Result"
	sqlQueryMany  = "Query_Many"
	sqlQueryOne   = "Query_One"
)

func (p *parser) sql(r *Method, m scanner.GoMethod, sql string) error {
	phs, err := scanner.SQLPlaceholders(sql)
	if err != nil {
		return errors.Fmt(`parse SQL placeholders failed: %v`, err)
	}
	r.SQL = SQL{
		String:   phs.SQL,
		Prepares: phs.Prepares,
		Replaces: phs.Replaces,
	}

	token, _ := p.next()
	switch token.Flag {
	case "INSERT", "DELETE", "UPDATE":
		execType := "result"
		var tagLine int
		switch len(m.Tags) {
		case 0:
		case 1:
			tag := m.Tags[0]
			if len(tag.Args) != 1 {
				err := errors.New(`exec tag bad format`)
				return errors.Line(err, tag.Line)
			}
			tagLine = tag.Line
			execType = tag.Args[0]
		default:
			err := errors.Fmt(`too many tags for "%s"`, r.Name)
			return errors.Line(err, p.line)
		}

		switch execType {
		case "affect":
			r.Type = sqlExecAffect
		case "lastid":
			r.Type = sqlExecLastid
		case "result":
			r.Type = sqlExecResult
		default:
			err := errors.Fmt(`unknown exec type "%s"`, execType)

			return errors.Line(err, tagLine)
		}

		if execType == "affect" || execType == "lastid" {
			if r.RetType != "int64" {
				err := errors.Fmt(`return type for "affect" or `+
					`"lastid" must be "int64", found: "%s"`, r.RetType)
				return errors.Line(err, p.line)
			}
		}

		if execType == "result" {
			if r.RetType != "sql.Result" {
				err := errors.Fmt(`return type for "result"`+
					` must be "sql.Result", found: "%s"`, r.RetType)
				return errors.Line(err, p.line)

			}
		}

		return nil
	}

	err = p.parseQuery(r)
	if err != nil {
		err = errors.Trace("parse query failed", err)
		return errors.Line(err, p.line)
	}

	if r.RetSlice {
		r.Type = sqlQueryMany
	} else {
		r.Type = sqlQueryOne
	}

	return nil
}

func (p *parser) parseQuery(r *Method) error {
	// Extract SELECT clause
	selectTokens := make([]token.Token, 0)
	for {
		token, ok := p.next()
		if !ok {
			return errors.New("can not find FROM")
		}
		if token.Flag == "FROM" {
			break
		}
		selectTokens = append(selectTokens, *token)
	}
	if len(selectTokens) == 0 {
		return errors.New("empty SELECT clause")
	}
	selectP := newParser(p.line, selectTokens)
	fields, err := selectP.parseSelect()
	if err != nil {
		return err
	}
	r.QueryFields = fields

	return nil
}

func (p *parser) parseSelect() ([]QueryField, error) {
	var fields []QueryField
	for {
		token, ok := p.pick()
		if !ok {
			return fields, nil
		}
		var field QueryField
		var err error
		if !token.IsIndent() {
			switch token.Flag {
			case "IFNULL":
				p.next()
				err = p.ifnull(&field)
			case "COUNT":
				p.next()
				err = p.count(&field)

			default:
				return nil, errors.Fmt(`unknown `+
					`field start "%s", except: `+
					`INDENT,IFNULL,COUNT`, token.Flag)
			}
		} else {
			err = p.field(&field)
		}
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
}

var (
	errIfnullBadFormat  = errors.New("IFNULL is bad format")
	errIfnullMissIndent = errors.New("IFNULL missing INDENT")
)

func (p *parser) ifnull(f *QueryField) error {
	token, _ := p.next()
	if token.Flag != "(" {
		return errIfnullBadFormat
	}

	for {
		token, ok := p.next()
		if !ok {
			return errIfnullBadFormat
		}
		if token.Flag == ")" {
			break
		}
		if token.IsIndent() && f.Field == "" {
			next, _ := p.pick()
			if next.Flag == "." {
				f.Table = token.Indent
				p.next()
				token, _ = p.next()
				if !token.IsIndent() {
					return errIfnullMissIndent
				}
				f.Field = token.Indent
				continue
			}
			f.Field = token.Indent
		}
	}
	if f.Field == "" {
		return errIfnullMissIndent
	}
	p.alias(f)
	return nil
}

func (p *parser) count(f *QueryField) error {
	f.Field = "count"
	for {
		_, ok := p.next()
		if !ok {
			return nil
		}
	}
}

var errFieldBadFormat = errors.New("field is bad format")

func (p *parser) field(f *QueryField) error {
	token, _ := p.next()
	if !token.IsIndent() {
		return errFieldBadFormat
	}
	next, _ := p.pick()
	if next.Flag == "." {
		f.Table = token.Indent
		p.next()
		token, _ = p.next()
		if !token.IsIndent() {
			return errFieldBadFormat
		}
	}
	f.Field = token.Indent

	p.alias(f)
	return nil
}

func (p *parser) alias(f *QueryField) {
	for {
		token, ok := p.pick()
		if !ok || token.Flag == "," {
			last, _ := p.last()
			if last.IsIndent() && last.Indent != f.Field {
				f.Alias = last.Indent
			}
			p.next()
			return
		}
		p.next()
	}
}
