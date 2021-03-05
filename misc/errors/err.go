package errors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/fioncat/go-gendb/misc/term"
)

func New(text string) error {
	return errors.New(text)
}

type traceError struct {
	stack []interface{}
	err   error
}

func (e *traceError) Error() string {
	trace := make([]string, 0, len(e.stack))
	for idx := len(e.stack) - 1; idx >= 0; idx-- {
		s := fmt.Sprint(e.stack[idx])
		trace = append(trace, s)
	}

	return fmt.Sprintf("%s: %v", strings.Join(trace, ":"), e.err)
}

func TraceFmt(prefix interface{}, a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return Trace(prefix, err)
}

func Trace(prefix interface{}, err error) error {
	if ce, ok := err.(*compileError); ok {
		ce.ori = Trace(prefix, err)
		return ce
	}
	if te, ok := err.(*traceError); ok {
		te.stack = append(te.stack, prefix)
		return te
	}
	return &traceError{
		err:   err,
		stack: []interface{}{prefix},
	}
}

func decodeTrace(err error) []string {
	if te, ok := err.(*traceError); ok {
		strs := make([]string, len(te.stack))
		for idx, v := range te.stack {
			strs[idx] = fmt.Sprint(v)
		}
		return strs
	}
	return nil
}

type compileError struct {
	lines   []string
	lineIdx int
	chIdx   int

	path string

	ori error
}

func (e *compileError) Error() string {
	return e.ori.Error()
}

func OnCompile(path string, lines []string, err error) error {
	if ce, ok := err.(*compileError); ok {
		return ce
	}
	ts := decodeTrace(err)
	if len(ts) == 0 {
		return err
	}
	var (
		lineIdxStr string
		chIdxStr   string
	)
	switch len(ts) {
	case 0:
		return err

	case 1:
		// Miss path, add path and call again.
		err = Trace(path, err)
		return OnCompile(path, lines, err)

	case 2:
		lineIdxStr = ts[0]

	default:
		lineIdxStr = ts[1]
		chIdxStr = ts[0]
	}

	lineIdx, _err := strconv.Atoi(lineIdxStr)
	if _err != nil || lineIdx < 0 {
		return err
	}

	chIdx, _err := strconv.Atoi(chIdxStr)
	if _err != nil {
		chIdx = -1
	}

	return &compileError{
		lines:   lines,
		lineIdx: lineIdx,
		chIdx:   chIdx,
		ori:     err,
		path:    path,
	}
}

const showOtherNum = 4

func ShowCompile(err error) {
	ce, ok := err.(*compileError)
	if !ok {
		return
	}

	start := ce.lineIdx - showOtherNum - 1
	end := ce.lineIdx + showOtherNum

	var numMaxLen int
	var lineMaxLen int
	for idx := start; idx <= end; idx++ {
		if idx < 0 || idx >= len(ce.lines) {
			continue
		}
		num := idx + 1
		numLen := len(strconv.Itoa(num))
		if numLen > numMaxLen {
			numMaxLen = numLen
		}
		if len(ce.lines[idx]) > lineMaxLen {
			lineMaxLen = len(ce.lines[idx])
		}
	}
	numFmt := "%" + strconv.Itoa(numMaxLen) + "d"

	fmt.Println()
	for idx := start; idx <= end; idx++ {
		if idx < 0 || idx >= len(ce.lines) {
			continue
		}
		num := idx + 1
		numStr := fmt.Sprintf(numFmt, num)
		line := ce.lines[idx]
		sep := "▕ "
		if idx+1 == ce.lineIdx {
			// error line, need to mark.
			line = markError(line, ce.chIdx)
			numStr = term.UnderLine(numStr)
		}
		fmt.Printf("%s%s%s\n", numStr, sep, line)
	}
}

func markError(line string, chIdx int) string {
	if line == "" {
		return ""
	}
	if chIdx < 0 {
		return term.Mark(line)
	}
	var space int
	var spaceRune []rune
	for _, r := range []rune(line) {
		if unicode.IsSpace(r) {
			space++
			spaceRune = append(spaceRune, r)
			continue
		}
		break
	}

	chIdx += space
	chIdx = correctIdx(chIdx, len(line))
	start := correctIdx(chIdx-2, len(line))
	end := correctIdx(chIdx+2, len(line))

	var front, after, marked string

	if start == end {
		front = line[:start]
		marked = term.Mark(string(line[start]))
		if start < len(line)-1 {
			after = line[start+1:]
		}
	} else {
		front = line[:start]
		marked = term.Mark(line[start : end+1])
		if end < len(line)-1 {
			after = line[end+1:]
		}
	}

	front = strings.TrimLeftFunc(front, unicode.IsSpace)
	after = strings.TrimRightFunc(after, unicode.IsSpace)

	front = term.UnderLine(front)
	after = term.UnderLine(after)

	prefix := string(spaceRune)

	return prefix + front + marked + after
}

func correctIdx(idx int, l int) int {
	if idx < 0 {
		return 0
	}
	if idx >= l {
		return l - 1
	}
	return idx

}
