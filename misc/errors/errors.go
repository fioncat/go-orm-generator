package errors

import (
	errbase "errors"
	"fmt"
)

func Fmt(s string, args ...interface{}) error {
	s = fmt.Sprintf(s, args...)
	return New(s)
}

func New(s string) error {
	return errbase.New(s)
}

func Trace(reason string, err error) error {
	return Fmt("%s: %v", reason, err)
}

func Line(path string, line int, err error) error {
	reason := fmt.Sprintf("%s:%d", path, line)
	return Trace(reason, err)
}
