package golang

import "github.com/fioncat/go-gendb/compile/base"

type Import struct {
	Name string
	Path string
}

type Interface struct {
	Name    string
	Tag     *base.Tag
	Methods []Method
}

type Method struct {
	Tags []*base.Tag

	Name    string
	Imports []string

	RetSlice   bool
	RetPointer bool
	RetSimple  bool

	RetType string

	Def string
}
