package common

import "github.com/fioncat/go-gendb/coder"

type Result interface {
	Source() string
	Generator() string
	Key() string
}

type StructResult struct {
	src string

	Structs []coder.Struct
}

func (r *StructResult) Source() string {
	return r.src
}

func (r *StructResult) Generator() string {
	return "go-struct"
}

func (r *StructResult) Key() string {
	return "struct"
}
