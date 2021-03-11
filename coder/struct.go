package coder

import (
	"fmt"
	"strconv"
	"strings"
)

type StructGroup struct {
	ss []*Struct
}

func (sg *StructGroup) Add() *Struct {
	s := new(Struct)
	sg.ss = append(sg.ss, s)
	return s
}

func (sg *StructGroup) Gets() []*Struct {
	return sg.ss
}

type Struct struct {
	comm string
	name string
	fs   []*Field
}

func (s *Struct) code(c *Coder) bool {
	if s.comm != "" {
		c.P(0, "// ", s.name, " ", s.comm)
	}
	if len(s.fs) == 0 {
		c.P(0, "type ", s.name, " struct {}")
		return true
	}

	c.P(0, "type ", s.name, " struct {")

	var nameLen int
	var typeLen int
	for _, f := range s.fs {
		if len(f.name) > nameLen {
			nameLen = len(f.name)
		}
		if len(f._type) > typeLen {
			typeLen = len(f._type)
		}
	}
	nameFmt := "%-" + strconv.Itoa(nameLen) + "s"
	typeFmt := "%-" + strconv.Itoa(typeLen) + "s"

	for _, f := range s.fs {
		name := fmt.Sprintf(nameFmt, f.name)
		_type := fmt.Sprintf(typeFmt, f._type)

		var tagStr string
		if len(f.tags) > 0 {
			ts := make([]string, len(f.tags))
			for idx, t := range f.tags {
				_s := fmt.Sprintf(`%s:"%s"`, t.key, t.val)
				ts[idx] = _s
			}
			tagStr = strings.Join(ts, " ")
		}
		c.P(1, name, " ", _type, " `", tagStr, "`")
	}
	c.P(0, "}")

	return true
}

func (s *Struct) Comment(str string, vs ...interface{}) {
	s.comm = fmt.Sprintf(str, vs...)
}

func (s *Struct) SetName(name string) {
	s.name = name
}

func (s *Struct) AddField() *Field {
	f := new(Field)
	s.fs = append(s.fs, f)
	return f
}

type Field struct {
	name  string
	_type string
	tags  []*Tag
}

func (f *Field) Set(name, _type string) {
	f.name = name
	f._type = _type
}

func (f *Field) AddTag(key, val string) {
	t := &Tag{key: key, val: val}
	f.tags = append(f.tags, t)
}

type Tag struct {
	key string
	val string
}
