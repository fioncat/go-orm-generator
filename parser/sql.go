package parser

import (
	"regexp"
	"strings"

	"github.com/fioncat/go-gendb/generator/coder"
	"github.com/fioncat/go-gendb/misc/errors"
)

var sqlParamRe = regexp.MustCompile(`[\$\#]\{[^\}]+\}`)

func SQLParam(sql string) ([]string, []string, string) {
	prepares := make([]string, 0)
	replaces := make([]string, 0)
	sql = sqlParamRe.ReplaceAllStringFunc(sql, func(ph string) string {
		name := ph[2 : len(ph)-1]
		switch ph[0] {
		case '$':
			prepares = append(prepares, name)
			return "?"
		case '#':
			replaces = append(replaces, name)
			return "%v"
		}
		return ph
	})
	return prepares, replaces, sql
}

func SQLQuery(sql string) ([]string, error) {
	up := strings.ToUpper(sql)
	selectIdx := strings.Index(up, "SELECT")
	fromIdx := strings.Index(up, "FROM")
	if selectIdx > 0 && fromIdx > 0 {
		return nil, errors.New("missing SELECT or FROM")
	}
	if selectIdx+6 >= fromIdx {
		return nil, errors.New("SELECT must be front of FROM")
	}
	selectClause := sql[selectIdx+6 : fromIdx]

	return extractFields(selectClause), nil
}

var ifnullRe = regexp.MustCompile(`ifnull\([^\)]+\)`)

func extractFields(clause string) []string {
	// Handling the special case of 'ifnull'
	clause = ifnullRe.ReplaceAllStringFunc(clause, func(in string) string {
		in = strings.TrimLeft(in, "ifnull(")
		in = strings.TrimRight(in, ")")
		tmp := strings.Split(in, ",")
		// 第一个作为名称替换掉ifnull
		return tmp[0]
	})

	fields := strings.Split(clause, ",")
	fieldNames := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}

		fieldParts := FD(field)
		hasAlias := len(fieldParts) > 1
		namePart := fieldParts[len(fieldParts)-1]

		subParts := strings.Split(namePart, ".")
		subPart := subParts[len(subParts)-1]

		if subPart == "" {
			continue
		}

		var name string
		if hasAlias {
			name = subPart
		} else {
			name = coder.GoName(subPart)
		}

		fieldNames = append(fieldNames, name)
	}
	return fieldNames
}
