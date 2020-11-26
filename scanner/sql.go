package scanner

import (
	"strings"

	"github.com/fioncat/go-gendb/misc/col"
	"github.com/fioncat/go-gendb/scanner/sql"
	"github.com/fioncat/go-gendb/scanner/token"
)

type SQLResult struct {
	Path string

	nameSet col.Set
}

type SQL struct {
	Name     string        `json:"name"`
	Tokens   []token.Token `json:"-"`
	TokenStr string        `json:"tokens"`
}

func SQLFile(path string) (*SQLResult, error) {

}

func scanSQL(s string) []token.Token {
	scanner := newChars(s)
	var bucket []rune
	var tokens []token.Token
	addToken := func() {
		str := string(bucket)
		key := strings.ToUpper(str)
		if sql.Keywords.Exists(key) {
			tokens = append(tokens, token.NewFlag(key))
		} else {
			tokens = append(tokens,
				token.NewIndent(str))
		}
		bucket = nil
	}
	for {
		rune, ok := scanner.next()
		if !ok {
			break
		}
		if rune == ' ' || rune == '\n' {
			if len(bucket) == 0 {
				continue
			}
			addToken()
			// After add keyword/indent, add a space
			tokens = append(tokens, token.NewFlag(sql.SPACE))
			bucket = nil
			continue
		}

		bucket = append(bucket, rune)
		key := strings.ToUpper(string(bucket))
		if sql.Keywords.Exists(key) {
			tokens = append(tokens, token.NewFlag(key))
			bucket = nil
		}
	}
	if len(bucket) > 0 {
		addToken()
	}
	return tokens
}
