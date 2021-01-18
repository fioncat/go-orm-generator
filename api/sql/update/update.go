package update

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/api/sql/where"
)

type Update struct {
	Where *where.Where

	fields []string
	vs     []interface{}
}

func New(updateN, whereN int) *Update {
	u := new(Update)
	u.Where = where.New(whereN)
	u.fields = make([]string, 0, updateN)
	u.vs = make([]interface{}, 0, updateN)
	return u
}

func (u *Update) Set(name string, value interface{}) *Update {
	u.fields = append(u.fields, name)
	u.vs = append(u.vs, value)
	return u
}

func (u *Update) Build(table string) (string, []interface{}) {
	ups := make([]string, len(u.fields))
	for i := 0; i < len(u.fields); i++ {
		ups[i] = fmt.Sprintf("%s=?", u.fields[i])
	}

	parts := make([]string, 1, 2)
	parts[0] = fmt.Sprintf("UPDATE %s SET %s", table, strings.Join(ups, ", "))
	where, whereVs := u.Where.Get()
	if where != "" {
		parts = append(parts, fmt.Sprintf("WHERE %s", where))
	}

	vs := u.vs
	vs = append(vs, whereVs...)

	return strings.Join(parts, " "), vs
}
