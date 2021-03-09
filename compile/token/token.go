package token

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/misc/errors"
)

type Token string

func (t Token) PrefixOf(s string) bool {
	return strings.HasPrefix(s, string(t))
}

func (t Token) Trim(s string) string {
	s = strings.TrimLeft(s, string(t))
	s = strings.TrimSpace(s)
	return s
}

func (t Token) Equal(s string) bool {
	return s == string(t)
}

func (t Token) Get() string {
	return string(t)
}

const (
	SPACE = Token(" ")

	LPAREN = Token("(")
	RPAREN = Token(")")

	PLUS  = Token("+")
	COLON = Token(":")

	LBRACE = Token("{")
	RBRACE = Token("}")

	LBRACK = Token("[")
	RBRACK = Token("]")

	MUL    = Token("*")
	COMMA  = Token(",")
	PERIOD = Token(".")

	PERCENT = Token("%")

	EQ = Token("=")
	GT = Token(">")
	LT = Token("<")

	BREAK = Token("\n")
)

const (
	_Space = ' '
	_Quo0  = '"'
	_Quo1  = '\''
	_Quo2  = '`'
)

type Scanner struct {
	line    string
	lineIdx int

	es  []Element
	idx int

	tokens []Token

	kws map[string]int
	chs map[rune]int

	ignoreCase bool

	BreakLine bool
}

func EmptyScannerIC(tokens []Token) *Scanner {
	return newScanner(tokens, true)
}

func EmptyScanner(tokens []Token) *Scanner {
	return newScanner(tokens, false)
}

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

func (s *Scanner) Empty() bool {
	if s == nil {
		return true
	}
	if len(s.es) == 0 {
		return true
	}
	return false
}

func (s *Scanner) AddLine(lidx int, line string) {
	var bucket []rune
	var quo *rune

	s.line = line
	s.lineIdx = lidx

	rs := []rune(line)
	flush := func(idx int) {
		if len(bucket) == 0 {
			if quo == nil {
				return
			}
		}
		str := string(bucket)
		kw := str
		if s.ignoreCase {
			kw = strings.ToUpper(str)
		}
		var e Element
		e.Pos = idx - len(bucket)
		e.line = lidx + 1
		if quo != nil {
			// - STRINGs
			e.Token = Token(str)
			e.String = true
			e.StringRune = *quo
			quo = nil
		} else if tIdx, ok := s.kws[kw]; ok {
			// - KEYWORDs
			e.Token = s.tokens[tIdx]
		} else {
			// - INDENTs
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

func NewScanner(line string, tokens []Token) *Scanner {
	s := EmptyScanner(tokens)
	s.AddLine(0, line)
	return s
}

func (s *Scanner) Cur(e *Element) bool {
	if s.idx >= len(s.es) {
		return false
	}
	*e = s.es[s.idx]
	return true
}

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

func (s *Scanner) Pervious(e *Element, delta int) bool {
	idx := s.idx - delta
	if idx < 0 {
		return false
	}
	*e = s.es[idx]
	return true
}

func (s *Scanner) Stop() {
	s.idx = len(s.es)
}

func (s *Scanner) Reset() {
	s.idx = 0
}

func (s *Scanner) EarlyEnd(expect string) error {
	return errors.Trace(len(s.line),
		fmt.Errorf("expect %s, found: 'EOF'", expect))
}

func (s *Scanner) EarlyEndL(expect string) error {
	err := s.EarlyEnd(expect)
	num := s.lineIdx + 1
	if num < 0 {
		return err
	}
	return errors.Trace(s.lineIdx+1, err)
}

func (s *Scanner) Line() string {
	return s.line
}

func (s *Scanner) Gets() []Element {
	return s.es
}

type Element struct {
	Space bool

	line int

	StringRune rune

	Pos    int
	Token  Token
	Indent bool
	String bool
}

func (e Element) Get() string {
	return string(e.Token)
}

func (e Element) FmtErr(a string, b ...interface{}) error {
	if e.Pos < 0 {
		return fmt.Errorf(a, b...)
	}
	return errors.TraceFmt(e.Pos, a, b...)
}

func (e Element) FmtErrL(a string, b ...interface{}) error {
	err := e.FmtErr(a, b...)
	if e.line < 0 {
		return err
	}
	return errors.Trace(e.line, err)
}

func (e Element) NotMatch(expect string) error {
	return e.FmtErr("expect %s, found: '%s'", expect, e.Type())
}

func (e Element) NotMatchL(expect string) error {
	err := e.NotMatch(expect)
	return errors.Trace(e.line, err)
}

func (e Element) Type() string {
	if e.Indent {
		return "INDENT"
	}
	if e.String {
		return "STRING"
	}
	return string(e.Token)
}
