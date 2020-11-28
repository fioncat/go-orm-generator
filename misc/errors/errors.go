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

type CompileError struct {
	File  string
	Line  int
	Pos   int
	Cause error
}

func (e *CompileError) Error() string {
	if e.Pos == 0 {
		return fmt.Sprintf("%s:%d: %v", e.File, e.Line, e.Cause)
	}
	return fmt.Sprintf("%s:%d:%d: %v",
		e.File, e.Line, e.Pos, e.Cause)
}

func Line(err error, line int) error {
	return &CompileError{Cause: err, Line: line}
}

func Pos(err error, line, pos int) error {
	return &CompileError{Cause: err, Line: line, Pos: pos}
}

func TraceFile(err error, path string) error {
	if ce, ok := err.(*CompileError); ok {
		if ce.File != "" {
			return err
		}
		ce.File = path
		return ce
	}
	return Fmt("%s: %v", path, err)
}

func Trace(msg string, err error) error {
	return Fmt("%s: %v", msg, err)
}

var ErrInvalidType = errbase.New("invalid type")
