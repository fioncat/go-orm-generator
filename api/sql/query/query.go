package query

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/api/sql/where"
)

type Query struct {
	*where.Where

	fields []string

	orderby string

	offset int
	limit  int
}

func New(n int) *Query {
	q := new(Query)
	q.Where = where.New(n)
	return q
}

func (q *Query) Limit(offset, limit int) *Query {
	q.offset = offset
	q.limit = limit
	return q
}

func (q *Query) OrderBy(fields ...string) *Query {
	q.orderby = strings.Join(fields, ", ")
	return q
}

func (q *Query) Build(table string, fields []string) (string, []interface{}) {
	if len(q.fields) > 0 {
		fields = q.fields
	}
	fieldStr := strings.Join(fields, ", ")

	parts := make([]string, 1, 4)
	parts[0] = fmt.Sprintf("SELECT %s FROM %s", fieldStr, table)

	where, vs := q.Where.Get()
	if where != "" {
		parts = append(parts, fmt.Sprintf("WHERE %s", where))
	}

	if q.orderby != "" {
		parts = append(parts, fmt.Sprintf("ORDER BY %s", q.orderby))
	}

	if q.offset > 0 && q.limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d, %d",
			q.offset, q.limit))
	} else if q.limit > 0 && q.offset <= 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", q.limit))
	}

	return strings.Join(parts, " "), vs
}
