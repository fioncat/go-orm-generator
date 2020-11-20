package coder

type Coder struct {
	Source string

	Pkg string

	Imports []Import
	Vars    []Var
	Consts  []Var
	Structs []Struct

	Contents []string
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
