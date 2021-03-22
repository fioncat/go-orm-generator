package token

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/misc/errors"
)

// Token is the smallest unit in grammatical analysis.
// It is the result of lexical analysis (also called
// "scanning"). It can be a symbol, keyword, variable,
// string, etc.
// All the grammatical analysis work of this project
// is done based on Token.
type Token string

// PrefixOf returns whether "s" is prefixed with the
// current Token.
func (t Token) PrefixOf(s string) bool {
	return strings.HasPrefix(s, string(t))
}

// Trim removes the token in front of s.
func (t Token) Trim(s string) string {
	s = strings.TrimLeft(s, string(t))
	s = strings.TrimSpace(s)
	return s
}

// Equal returns whether the token is equal to s.
func (t Token) Equal(s string) bool {
	return s == string(t)
}

// Get returns the token as a string.
func (t Token) Get() string {
	return string(t)
}

// All the symbols that need to be used.
const (
	SPACE = Token(" ")

	LPAREN = Token("(")
	RPAREN = Token(")")

	COLON = Token(":")

	LBRACE = Token("{")
	RBRACE = Token("}")

	LBRACK = Token("[")
	RBRACK = Token("]")

	MUL    = Token("*")
	COMMA  = Token(",")
	PERIOD = Token(".")
	PLUS   = Token("+")

	PERCENT = Token("%")

	EQ = Token("=")
	GT = Token(">")
	LT = Token("<")

	BREAK = Token("\n")
)

// Spaces and strings
const (
	_Space = ' '
	_Quo0  = '"'
	_Quo1  = '\''
	_Quo2  = '`'
)

// Scanner is used to convert code content into multiple
// tokens. What it essentially performs is the lexical
// analysis process.
// After the lexical analysis is over, tokens can be obtained
// one by one in an iterative manner.
// In order to distinguish different types of tokens
// (keywords, symbols, indents, strings) and save some meta
// information of the token (line number and location, used
// for error reporting), Scanner will wrap the token as
// Element and return it.
// Sample:
//
//    code := "var a = 32"
//    codeTokens := []Token{Token("var"), EQ}
//    s := NewScanner(code, codeTokens)
//    var e Element
//    for s.Next(&e) {
//       ...
//    }
//
// See https://github.com/fioncat/go-gendb/doc/compile
// for more usage.
type Scanner struct {
	// The original line content of the scan, if it is a
	// multi-line scan, it is the content of the last line.
	line string

	// The original line index of the scan, if it is a
	// multi-line scan, it is the line index of the last line.
	lineIdx int

	// elements
	es []Element

	// iterative index
	idx int

	// keywords and chars tokens.
	tokens []Token

	// the mapping of keywords&chars and tokens indexes.
	kws map[string]int
	chs map[rune]int

	// If true, the case will be ignored when determining
	// keywords. The default is case sensitive.
	ignoreCase bool

	BreakLine bool
}

// EmptyScannerIC is the same as EmptyScanner, the
// difference is that the Scanner it returns is
// case-insensitive to keywords.
func EmptyScannerIC(tokens []Token) *Scanner {
	return newScanner(tokens, true)
}

// EmptyScanner uses the given tokens as keywords and
// symbols to create a new Scanner. It has no content.
// Caller should add content to it later through "AddLine".
// This function is generally used to create a Scanner
// with multiple lines of content (AddLine needs to be
// called repeatedly). If there is only one line of content,
// should use NewScanner instead.
func EmptyScanner(tokens []Token) *Scanner {
	return newScanner(tokens, false)
}

// newScanner creates an empty Scanner, ignoreCase
// indicates whether it is case-insensitive.
func newScanner(tokens []Token, ignoreCase bool) *Scanner {
	s := new(Scanner)
	s.kws = make(map[string]int)
	s.chs = make(map[rune]int)
	s.ignoreCase = ignoreCase
	for idx, t := range tokens {
		if len(t) == 1 {
			s.chs[[]rune(t)[0]] = idx
			continue
		}
		str := string(t)
		if ignoreCase {
			str = strings.ToUpper(str)
		}
		s.kws[str] = idx
	}
	s.tokens = tokens
	return s

}

// CopyScanner uses os data to copy a new Scanner, and
// the elements of the new Scanner will use es instead of os.
// This function is generally used in some processes that
// require separate grammatical analysis of certain key es,
// such as analyzing the SELECT clause separately when
// parsing SQL statements
func CopyScanner(os *Scanner, es []Element) *Scanner {
	s := new(Scanner)
	s.line = os.line
	s.lineIdx = os.lineIdx

	s.tokens = os.tokens
	s.chs = os.chs
	s.kws = os.kws

	s.es = es
	s.idx = 0

	return s
}

// LinesIdx returns the line index of the scanner. If it is a
// multi-line scan, the index of the last line will be returned.
func (s *Scanner) LinesIdx() int {
	return s.lineIdx
}

// Empty returns whether the Scanner is empty.
func (s *Scanner) Empty() bool {
	if s == nil {
		return true
	}
	if len(s.es) == 0 {
		return true
	}
	return false
}

// AddLine scans a line of content and converts it into
// elements for storage. Caller can get them one by one
// through Next later.
func (s *Scanner) AddLine(lidx int, line string) {
	var bucket []rune
	var quo *rune

	s.line = line
	s.lineIdx = lidx

	rs := []rune(line)
	flush := func(idx int) {
		if len(bucket) == 0 {
			// If the bucket is empty, but quo is not empty,
			// it means that this is an "empty string" and
			// it still exists as an element, so continue
			// processing.
			if quo == nil {
				return
			}
		}
		str := string(bucket)
		kw := str
		if s.ignoreCase {
			// In the case of ignoreCase, keywords have
			// already been capitalized, so the capitalization
			// processing must be done synchronously here.
			kw = strings.ToUpper(str)
		}
		var e Element
		e.Pos = idx - len(bucket)
		e.line = lidx + 1
		if quo != nil {
			// Token is a string.
			e.Token = Token(str)
			e.String = true
			e.StringRune = *quo
			quo = nil
		} else if tIdx, ok := s.kws[kw]; ok {
			// Token is a keyword.
			e.Token = s.tokens[tIdx]
		} else {
			// Token is an indent.
			e.Token = Token(str)
			e.Indent = true
		}

		if idx >= len(rs)-1 {
			e.Space = true
		} else {
			if rs[idx] == ' ' {
				e.Space = true
			}
		}

		s.es = append(s.es, e)
		// reset for next iteration.
		bucket = nil
	}
	for idx, r := range rs {
		if quo != nil && r != *quo {
			bucket = append(bucket, r)
			continue
		}
		if tIdx, ok := s.chs[r]; ok {
			flush(idx)
			var e Element
			e.Token = s.tokens[tIdx]
			e.Pos = idx
			e.line = lidx + 1
			s.es = append(s.es, e)
			continue
		}

		switch r {
		case _Space:
			if quo != nil {
				bucket = append(bucket, r)
				break
			}
			flush(idx)

		case _Quo0, _Quo1, _Quo2:
			if quo != nil {
				if *quo == r {
					// STRING ends
					flush(idx)
					break
				}
				// not end quo in STRING
				bucket = append(bucket, r)
				break
			}
			flush(idx)
			// STRING starts
			quo = new(rune)
			*quo = r

		default:
			bucket = append(bucket, r)
		}
	}
	flush(len(rs) - 1)

	if s.BreakLine {
		var e Element
		e.Token = BREAK
		e.Pos = len(rs)
		e.line = lidx + 1

		s.es = append(s.es, e)
	}
}

// NewScanner uses the specified line to create a Scanner.
// Scanner will directly scan the content of this line and
// generate elements.
func NewScanner(line string, tokens []Token) *Scanner {
	s := EmptyScanner(tokens)
	s.AddLine(0, line)
	return s
}

// Cur assigns the current element of the iteration to e.
// Returns whether the iteration is over.
func (s *Scanner) Cur(e *Element) bool {
	if s.idx >= len(s.es) {
		return false
	}
	*e = s.es[s.idx]
	return true
}

// Next assigns the current element to e and increments
// index. Returns whether the iteration is over.
func (s *Scanner) Next(e *Element) bool {
	if e == nil {
		if s.idx >= len(s.es) {
			return false
		}
		s.idx += 1
		return true
	}
	if ok := s.Cur(e); !ok {
		return false
	}
	s.idx += 1
	return true
}

// Pervious assigns the previous element relative to the
// current value to e. delta represents the offset, and
// if it is 1, the previous element of the current element
// is used.
func (s *Scanner) Pervious(e *Element, delta int) bool {
	idx := s.idx - delta
	if idx < 0 {
		return false
	}
	*e = s.es[idx]
	return true
}

// Stop directly stops the iteration.
func (s *Scanner) Stop() {
	s.idx = len(s.es)
}

// Reset directly resets the iteration.
func (s *Scanner) Reset() {
	s.idx = 0
}

// EarlyEnd returns an early termination error. That is,
// the next element is expected to be a certain value,
// but it has reached the end. This function will add the
// EOF position to the error trace.
func (s *Scanner) EarlyEnd(expect string) error {
	return errors.Trace(len(s.line),
		fmt.Errorf("expect %s, found: 'EOF'", expect))
}

// EarlyEndL is the same as EarlyEnd, the difference is that
// the line number is added to the error trace.
func (s *Scanner) EarlyEndL(expect string) error {
	err := s.EarlyEnd(expect)
	num := s.lineIdx + 1
	if num < 0 {
		return err
	}
	return errors.Trace(s.lineIdx+1, err)
}

// Line returns the original line content of the Scanner.
// If it is a multi-line scan, only the last line will be
// returned.
func (s *Scanner) Line() string {
	return s.line
}

// Gets returns all the elements directly.
func (s *Scanner) Gets() []Element {
	return s.es
}

// Element is an encapsulation of Token. Because in the
// compilation process, in addition to the original Token,
// some surrounding information is also needed, and this
// struct binds the Token and this information together.
type Element struct {
	// Space indicates whether the element is a space
	// or a tab.
	Space bool

	// line number
	line int

	// StringRune is a string symbol. Only applicable to
	// scenarios where the element is a string. There are
	// three situations: double quotes, single quotes, and
	// back quotes.
	StringRune rune

	// Pos is the position of the element in the line.
	Pos int

	// Token is the original token of the element.
	Token Token

	// Indent indicates whether the element is an Indent,
	// that is, in addition to keywords and symbols.
	Indent bool

	// String indicates whether the element is a string.
	String bool
}

// Get returns the element in the form of a string.
func (e Element) Get() string {
	return string(e.Token)
}

// FmtErr adds the position trace of the element on the
// basis of the fmt.Errorf.
func (e Element) FmtErr(a string, b ...interface{}) error {
	if e.Pos < 0 {
		return fmt.Errorf(a, b...)
	}
	return errors.TraceFmt(e.Pos, a, b...)
}

// FmtErrL adds the line number and position trace of the
// element on the basis of the fmt.Errorf.
func (e Element) FmtErrL(a string, b ...interface{}) error {
	err := e.FmtErr(a, b...)
	if e.line < 0 {
		return err
	}
	return errors.Trace(e.line, err)
}

// NotMatch returns a mismatch error. Indicates that the
// current element is expected to be a certain value, but
// it is not.
// This function will add position trace to the error.
func (e Element) NotMatch(expect string) error {
	return e.FmtErr("expect %s, found: '%s'", expect, e.Type())
}

// NotMatchL is the same as NotMatch, except that the line
// number error trace is added.
func (e Element) NotMatchL(expect string) error {
	err := e.NotMatch(expect)
	return errors.Trace(e.line, err)
}

// Type returns the element type as a string.
func (e Element) Type() string {
	if e.Indent {
		return "INDENT"
	}
	if e.String {
		return "STRING"
	}
	return string(e.Token)
}
