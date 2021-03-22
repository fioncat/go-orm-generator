package base

import "github.com/fioncat/go-gendb/compile/token"

// tokens uses in arrary string.
var arrTokens = []token.Token{
	token.LBRACK, token.RBRACK,
	token.COMMA,
}

// Arr2 compiles a two-dimensional array and converts it
// into a built-in two-dimensional slice.
// The format is "[[a, b, c], [d, e], [f]]"
func Arr2(val string) ([][]string, error) {
	s := token.NewScanner(val, arrTokens)
	var arrs [][]string
	var e token.Element
	for s.Next(&e) {
		if e.Token == token.COMMA {
			continue
		}
		if e.Token != token.LBRACK {
			return nil, e.NotMatch("LBRACK")
		}
		var names []string
		for {
			ok := s.Next(&e)
			if !ok {
				return nil, s.EarlyEnd("RBRACK")
			}
			if e.Token == token.RBRACK {
				break
			}
			if e.Token == token.COMMA {
				continue
			}
			names = append(names, e.Get())
		}
		arrs = append(arrs, names)
	}
	return arrs, nil
}

// Arr1 compiles the array and converts it to the built-in
// slice. The format is "[a, b, c, d, e]"
func Arr1(val string) ([]string, error) {
	s := token.NewScanner(val, arrTokens)
	var arr []string
	var e token.Element
	ok := s.Next(&e)
	if !ok {
		return nil, s.EarlyEnd("LBRACK")
	}
	if e.Token != token.LBRACK {
		return nil, e.NotMatch("LBRACK")
	}

	for {
		ok = s.Next(&e)
		if !ok {
			return nil, s.EarlyEnd("RBRACK")
		}
		if e.Token == token.RBRACK {
			break
		}
		if e.Token == token.COMMA {
			continue
		}
		arr = append(arr, e.Get())
	}
	return arr, nil
}
