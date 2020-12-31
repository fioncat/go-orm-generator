package errors

import "fmt"

type CompileError struct {
	Path    string
	LineNum int
	CharNum int
	Msg     string
}

func (e *CompileError) Error() string {
	if e.CharNum < 0 {
		return fmt.Sprintf("%s:%d: %s", e.Path, e.LineNum, e.Msg)
	}
	return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.LineNum, e.CharNum, e.Msg)
}

func (e *CompileError) WithCharNum(cn int) *CompileError {
	e.CharNum = cn
	return e
}

func NewComp(path string, lineNum int, msg string, vs ...interface{}) *CompileError {
	return &CompileError{
		Path:    path,
		LineNum: lineNum,
		CharNum: -1,
		Msg:     fmt.Sprintf(msg, vs...),
	}
}
