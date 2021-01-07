package coder

import (
	"strings"
	"unicode"

	"github.com/fioncat/go-gendb/misc/set"
)

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
	Name string `json:"name"`
	Path string `json:"path"`
}

type Struct struct {
	Comment string  `json:"comment"`
	Name    string  `json:"name"`
	Fields  []Field `json:"fields"`
}

func (s *Struct) Type() string {
	return "struct"
}

type Field struct {
	Comment string     `json:"comment"`
	Name    string     `json:"name"`
	Type    string     `json:"type"`
	Tags    []FieldTag `json:"tags"`
}

func (f *Field) AddTag(name, val string) {
	tag := FieldTag{
		Name:  name,
		Value: val,
	}
	f.Tags = append(f.Tags, tag)
}

type FieldTag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func IsSimpleType(t string) bool {
	switch {
	case strings.HasPrefix(t, "int"):
		return true
	case strings.HasPrefix(t, "uint"):
		return true

	case t == "string":
		return true

	case t == "bool":
		return true

	case strings.HasPrefix(t, "float"):
		return true

	}

	return false
}

func GoName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.Split(name, "_")
	for i := range parts {
		parts[i] = Export(parts[i])
	}

	return strings.Join(parts, "")
}

func Export(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return string(unicode.ToLower(rune(s[0])))
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

func Unexport(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return string(unicode.ToLower(rune(s[0])))
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}
