package token

import (
	"strings"

	"github.com/fioncat/go-gendb/misc/set"
)

// Token is the smallest unit produced by lexical analysis.
// It can be keywords, symbols, variables, etc. The original
// sentence is converted into multiple tokens to better provide
// to the parser for further analysis and create intermediate
// structure.
type Token struct {
	isIndent bool
	token    string
}

// String returns the token as a string. If the token is an indent,
// "INDENT" will be returned instead of the specific indent value.
func (t Token) String() string {
	if t.isIndent {
		return "INDENT"
	}
	return t.token
}

// IsIndent returns whether the token is an indent.
func (t Token) IsIndent() bool {
	return t.isIndent
}

// Get returns the token value directly, if it is an indent, will
// return the value of the indent.
func (t Token) Get() string {
	return t.token
}

// Rune returns the token in the form of character. Limited to
// tokens with only one character. If there is more than one, return 0.
func (t Token) Rune() rune {
	if len(t.token) != 1 {
		return 0
	}
	return rune(t.token[0])
}

// EqRune compares whether the token is equal to a given character.
func (t Token) EqRune(r rune) bool {
	if len(t.token) != 1 {
		return false
	}
	return r == rune(t.token[0])
}

// EqString compares whether the token is equal to a given string.
func (t Token) EqString(s string) bool {
	return s == t.token
}

// Prefix determines whether the incoming string is prefixed with
// the current token.
func (t Token) Prefix(s string) bool {
	return strings.HasPrefix(s, t.token)
}

// Contains determines whether the incoming string contains the
// current token.
func (t Token) Contains(s string) bool {
	return strings.Contains(s, t.token)
}

func (t Token) Match(ot Token) bool {
	return t.Get() == ot.Get()
}

// Indent creates a new indent token.
func Indent(indent string) Token {
	return Token{isIndent: true, token: indent}
}

// Key creates a new key token.
func Key(key string) Token {
	return Token{isIndent: false, token: key}
}

// Bucket stores an slice of tokens and a buffer of indent characters.
// Support adding new key tokens to the tokens slice or converting
// the data in the character buffer into indent and append to the
// token slice.
// This structure is generally a temporary cache structure during
// scanning (syntax analysis).
type Bucket struct {
	tokens []Token
	buff   []rune

	keywords *set.Set
}

// NewBucket creates a new empty Bucket
func NewBucket() *Bucket {
	return new(Bucket)
}

// Append add the "c" to the characters' buffer.
func (b *Bucket) Append(c rune) {
	b.buff = append(b.buff, c)
}

// SetKeywords set some keywords. If the keyword hits when adding
// indent, it will be added in the way of key token.
func (b *Bucket) SetKeywords(kws []string) {
	for i, kw := range kws {
		kws[i] = strings.ToUpper(kw)
	}
	b.keywords = set.New(kws...)
}

// Indent force convert buffer into indent token and append it
// to the tokens slice.
// Note that if SetKeywords has been called, and the content of
// the buffer happens to be one of the keywords, it will increase
// in the way of the key token.
func (b *Bucket) Indent() {
	if len(b.buff) == 0 {
		return
	}
	s := string(b.buff)
	if b.keywords != nil {
		if b.keywords.Contains(strings.ToUpper(s)) {
			// The indent is a keyword, add as key
			b.tokens = append(b.tokens,
				Key(strings.ToUpper(s)))
			b.buff = nil
			return
		}
	}
	b.tokens = append(b.tokens, Indent(s))
	b.buff = nil
}

// Key append a new key indent to the indent slice.
func (b *Bucket) Key(key Token) {
	b.Indent()
	b.tokens = append(b.tokens, key)
}

// Get returns the indent slice.
func (b *Bucket) Get() []Token {
	return b.tokens
}

func (b *Bucket) AddToken(t Token) {
	b.tokens = append(b.tokens, t)
}

// common key tokens.
var (
	EMPTY = Key("")
	SPACE = Key(" ")

	BREAK = Key("\n")
	TABLE = Key("\t")

	QUO  = Key(`"`)
	SQUO = Key(`'`)

	LPAREN = Key("(")
	RPAREN = Key(")")

	LBRACE = Key("{")
	RBRACE = Key("}")

	LBRACK = Key("[")
	RBRACK = Key("]")
	BRACKS = Key("[]")

	MUL    = Key("*")
	COMMA  = Key(",")
	PERIOD = Key(".")
)

// go-gendb tag
var (
	TAG_PREFIX = Key("// +gendb")
	TAG_NAME   = Key("+gendb")
)

// Golang keywords
var (
	GO_IMPORT    = Key("import")
	GO_IMPORTS   = Key("import (")
	GO_PACKAGE   = Key("package")
	GO_INTERFACE = Key("interface")
	GO_TYPE      = Key("type")
	GO_COMMENT   = Key("//")
	GO_ERROR     = Key("error")
)

// SQL tags
var (
	SQL_TAG     = Key("-- !")
	SQL_COMMENT = Key("--")
)

// SQL keywords
var (
	SQL_SELECT = Key("SELECT")
	SQL_FROM   = Key("FROM")
	SQL_INNER  = Key("INNER")
	SQL_LEFT   = Key("LEFT")
	SQL_RIGHT  = Key("RIGHT")
	SQL_JOIN   = Key("JOIN")
	SQL_ON     = Key("ON")
	SQL_WHERE  = Key("WHERE")
	SQL_ORDER  = Key("ORDER")
	SQL_BY     = Key("BY")
	SQL_AS     = Key("AS")
	SQL_GROUP  = Key("GROUP")
	SQL_IFNULL = Key("IFNULL")
	SQL_LIMIT  = Key("LIMIT")
	SQL_COUNT  = Key("COUNT")

	SQL_UPDATE = Key("UPDATE")
	SQL_DELETE = Key("DELETE")
	SQL_INSERT = Key("INSERT")

	SQL_Keywords = []string{
		SQL_SELECT.Get(),
		SQL_FROM.Get(),
		SQL_INNER.Get(),
		SQL_LEFT.Get(),
		SQL_RIGHT.Get(),
		SQL_JOIN.Get(),
		SQL_ON.Get(),
		SQL_WHERE.Get(),
		SQL_ORDER.Get(),
		SQL_BY.Get(),
		SQL_AS.Get(),
		SQL_GROUP.Get(),
		SQL_IFNULL.Get(),
		SQL_LIMIT.Get(),
		SQL_COUNT.Get(),
	}
)

// SQL Plarceholders
var (
	SQLPH_PRE = Key("$")
	SQLPH_REP = Key("#")
)
