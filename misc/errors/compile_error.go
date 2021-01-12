package errors

import (
	"fmt"

	"github.com/fioncat/go-gendb/compile/token"
)

// CompileError represents an error generated during
// the compilation process. It will save the compiled
// file path, the line and the character position where
// the error occurred. And compiled error messages.
// When this error is output, it is easily locate the
// location of the compilation error.
type CompileError struct {
	Path    string
	LineNum int
	CharNum int
	Msg     string
}

// Error returns the compile error msg.
func (e *CompileError) Error() string {
	if e.CharNum < 0 {
		return fmt.Sprintf("%s:%d: %s", e.Path, e.LineNum, e.Msg)
	}
	return fmt.Sprintf("%s:%d:%d: %s", e.Path, e.LineNum, e.CharNum, e.Msg)
}

// WithCharNum set the char number.
func (e *CompileError) WithCharNum(cn int) *CompileError {
	e.CharNum = cn
	return e
}

// NewComp creates a new CompileError and returns. By
// default, there is no character number for compilation
// errors, and you need to manually call the WithCharNum
// setting.
func NewComp(path string, lineNum int, msg string, vs ...interface{}) *CompileError {
	return &CompileError{
		Path:    path,
		LineNum: lineNum,
		CharNum: -1,
		Msg:     fmt.Sprintf(msg, vs...),
	}
}

// ParseErrorFactory is a tool factory that saves the line
// information of a file being compiled, and can directly
// create some line compilation errors.
type ParseErrorFactory struct {
	path    string
	lineNum int
}

// NewParseFactory creates a NewParseFactory.
func NewParseFactory(path string, lineNum int) *ParseErrorFactory {
	return &ParseErrorFactory{
		path:    path,
		lineNum: lineNum,
	}
}

// EarlyEnd : The expectation was a certain token, but it ended early.
func (f *ParseErrorFactory) EarlyEnd(expect string) error {
	return NewComp(f.path, f.lineNum, `expect "%s", but reach the end`, expect)
}

// MismatchS : Expect a certain string, but get a different token.
func (f *ParseErrorFactory) MismatchS(chNum int, expect string, found token.Token) error {
	return NewComp(f.path, f.lineNum, `expect "%s", found %s`,
		expect, found.String()).WithCharNum(chNum)
}

// Mismatch : Expect a certain token, but get a different token.
func (f *ParseErrorFactory) Mismatch(chNum int, expect, found token.Token) error {
	return NewComp(f.path, f.lineNum, `expect "%s", found "%s"`,
		expect.String(), found.String()).WithCharNum(chNum)
}

// Empty : Expect not to be empty, but an empty string.
func (f *ParseErrorFactory) Empty(chNum int, name string) error {
	return NewComp(f.path, f.lineNum, "%s is empty", name).WithCharNum(chNum)
}
