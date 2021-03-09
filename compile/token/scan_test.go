package token

import (
	"fmt"
	"testing"
)

var tokens = []Token{
	LPAREN,
	RPAREN,
	LBRACE,
	RBRACE,
	LBRACE,
	RBRACE,
	MUL,
	COMMA,
	PERIOD,

	Token("func"),
}

func TestScan(t *testing.T) {
	lines := []string{
		"func (e *Element) Get(a int, b int) (string, error) {",
		"func Hello(a string, b *user.Detail) (sql.Result, error) {",
		"Add(userId int64, detail *user.Detail) (int64, error)",
		`Add("I am a hero!", "My name is 'Json'")`,
	}

	for _, line := range lines {
		fmt.Println(line)
		s := NewScanner(line, tokens)
		var e Element
		for s.Next(&e) {
			fmt.Printf("%v|", e)
		}
		fmt.Println("\n----------------")
	}
}

func TestScanNoToken(t *testing.T) {
	lines := []string{
		`"database/sql"`,
		`sqlm "database/sql"`,
	}
	for _, line := range lines {
		fmt.Println(line)
		s := NewScanner(line, nil)
		var e Element
		for s.Next(&e) {
			fmt.Printf("%v|", e)
		}
		fmt.Println("\n----------------")
	}
}
