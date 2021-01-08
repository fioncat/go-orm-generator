package mediate

import "github.com/fioncat/go-gendb/generate/coder"

type Result interface {
	Type() string
	Key() string
	GetStructs() []*coder.Struct
}
