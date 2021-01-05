package errors

import (
	"fmt"

	"github.com/fioncat/go-gendb/compile/token"
)

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

type ParseErrorFactory struct {
	path    string
	lineNum int
}

func NewParseFactory(path string, lineNum int) *ParseErrorFactory {
	return &ParseErrorFactory{
		path:    path,
		lineNum: lineNum,
	}
}

func (f *ParseErrorFactory) EarlyEnd(expect string) error {
	return NewComp(f.path, f.lineNum, `expect "%s", but reach the end`, expect)
}

func (f *ParseErrorFactory) MismatchS(chNum int, expect string, found token.Token) error {
	return NewComp(f.path, f.lineNum, `expect "%s", found %s`,
		expect, found.String()).WithCharNum(chNum)
}

func (f *ParseErrorFactory) Mismatch(chNum int, expect, found token.Token) error {
	return NewComp(f.path, f.lineNum, `expect "%s", found "%s"`,
		expect.String(), found.String()).WithCharNum(chNum)
}

func (f *ParseErrorFactory) Empty(chNum int, name string) error {
	return NewComp(f.path, f.lineNum, "%s is empty", name).WithCharNum(chNum)

}
