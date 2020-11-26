package token

type Token struct {
	Flag   string
	Indent string
}

func (t Token) String() string {
	return t.Flag
}

func (t Token) IsIndent() bool {
	return t.Flag == "INDENT"
}

func NewFlag(flag string) Token {
	return Token{Flag: flag}
}

func NewIndent(indent string) Token {
	return Token{
		Flag:   "INDENT",
		Indent: indent,
	}
}
