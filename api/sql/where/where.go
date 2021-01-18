package where

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	Eq = "="
	Ne = "!="
	Gt = ">"
	Lt = "<"
	Ge = ">="
	Le = "<="
	In = "IN"
)

type Where struct {
	where string
	vs    []interface{}
}

func New(n int) *Where {
	w := new(Where)
	w.where = "%s"
	w.vs = make([]interface{}, 0, n)
	return w
}

func (w *Where) Add(name, symbol string, val interface{}) *Where {
	w.add("", name, symbol, val)
	return w
}

func (w *Where) And(name, symbol string, val interface{}) *Where {
	w.add(" AND", name, symbol, val)
	return w
}

func (w *Where) Or(name, symbol string, val interface{}) *Where {
	w.add(" OR", name, symbol, val)
	return w
}

func (w *Where) add(prefix, name, symbol string, val interface{}) {
	exp := w.exp(prefix, name, symbol, val)
	w.where = fmt.Sprintf(w.where, exp+"%s")
}

func (w *Where) AndAll() *Where {
	w.where = strings.Replace(w.where, "%s", " AND (%s)", 1)
	return w
}

func (w *Where) OrAll() *Where {
	w.where = strings.Replace(w.where, "%s", " OR (%s)", 1)
	return w
}

func (w *Where) End() *Where {
	w.where = strings.Replace(w.where, "%s)", ")%s", 1)
	return w
}

func (w *Where) exp(prefix, name, symbol string, val interface{}) string {
	if symbol == In {
		vs := toslice(val)
		valstr := strings.Repeat("?,", len(vs))
		valstr = valstr[:len(valstr)-1]
		exp := fmt.Sprintf("%s %s IN (%s)", prefix, name, valstr)

		w.vs = append(w.vs, vs...)
		return exp
	}

	exp := fmt.Sprintf("%s %s%s?", prefix, name, symbol)
	w.vs = append(w.vs, val)
	return exp
}

func (w *Where) Get() (string, []interface{}) {
	s := strings.Replace(w.where, "%s", "", 1)
	return strings.TrimSpace(s), w.vs
}

func toslice(v interface{}) []interface{} {
	slicev := reflect.ValueOf(v)
	if slicev.Kind() != reflect.Slice {
		return []interface{}{v}
	}
	length := slicev.Len()
	vs := make([]interface{}, length)
	for i := 0; i < length; i++ {
		ele := slicev.Index(i).Interface()
		vs[i] = ele
	}
	return vs
}
