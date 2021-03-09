package common

import (
	"fmt"

	"github.com/fioncat/go-gendb/compile/sql"
)

func FindMethodTag(m *sql.Method, name string) bool {
	for _, tag := range m.Tags {
		if tag.Name == name {
			return true
		}
	}
	return false
}

type Exec struct {
	Sql  string
	Vals []interface{}
}

func Method2Exec(m *sql.Method, name string) (
	*Exec, error,
) {
	if m.Dyn {
		return nil, m.FmtError(`dynamic ` +
			`method do not support execution.`)
	}
	vals := make(map[string]string)
	for _, tag := range m.Tags {
		if tag.Name != name {
			continue
		}
		for _, opt := range tag.Options {
			_, ok := vals[opt.Key]
			if ok {
				return nil, tag.FmtError(`param`+
					` "%s" is duplcate`, opt.Key)
			}
			vals[opt.Key] = opt.Value
		}
	}

	// Handle replace values
	reps := make([]interface{}, len(m.State.Replaces))
	for idx, rep := range m.State.Replaces {
		val, ok := vals[rep]
		if !ok {
			return nil, m.FmtError(`%s: can not `+
				`find value for replace placeholder`+
				` "%s"`, name, rep)
		}
		reps[idx] = val
	}
	sql := fmt.Sprintf(m.State.Sql, reps...)

	// Handle prepare values
	pres := make([]interface{}, len(m.State.Prepares))
	for idx, pre := range m.State.Prepares {
		val, ok := vals[pre]
		if !ok {
			return nil, m.FmtError(`%s: can not `+
				`find value for prepare placeholder`+
				` "%s"`, name, pre)
		}
		pres[idx] = val
	}

	return &Exec{Sql: sql, Vals: pres}, nil
}
