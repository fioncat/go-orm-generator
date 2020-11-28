package common

import "github.com/fioncat/go-gendb/coder"

type Result interface {
	Generator() string
	Key() string
}

type StructResult struct {
	Structs []coder.Struct
}

func (r *StructResult) Generator() string {
	return "go-struct"
}

func (r *StructResult) Key() string {
	return "struct"
}
