package coder

import "github.com/fioncat/go-gendb/misc/set"

type Coder struct {
	Source string

	Pkg string

	Imports []Import
	Vars    []Var
	Consts  []Var
	Structs []Struct

	Contents []string

	importsSet set.Set
}

type Var struct {
	Name  string
	Value string
}

type Import struct {
	Name string
	Path string
}

type Struct struct {
	Comment string
	Name    string
	Fields  []Field
}

func (s *Struct) Type() string {
	return "struct"
}

type Field struct {
	Comment string
	Name    string
	Type    string
	Tags    []FieldTag
}

func (f *Field) AddTag(name, val string) {
	tag := FieldTag{
		Name:  name,
		Value: val,
	}
	f.Tags = append(f.Tags, tag)
}

type FieldTag struct {
	Name  string
	Value string
}
