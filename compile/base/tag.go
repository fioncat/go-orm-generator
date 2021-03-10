package base

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/misc/errors"
)

type Tag struct {
	Line int

	Name    string
	Options []Option
}

func (t *Tag) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(t.Line, err)
}

type Option struct {
	Key   string
	Value string
}

var tagTokens = []token.Token{
	token.EQ,
}

// +gen:xxx
const tagPrefix = "+gen:"

func ParseTag(idx int, prefix, line string) (*Tag, error) {
	if !strings.HasPrefix(line, prefix) {
		return nil, nil
	}
	line = strings.TrimPrefix(line, prefix)
	line = strings.TrimSpace(line)

	s := token.NewScanner(line, tagTokens)
	var e token.Element

	ok := s.Next(&e)
	if !ok {
		return nil, nil
	}
	if !e.Indent {
		return nil, nil
	}
	def := e.Get()
	if !strings.HasPrefix(def, tagPrefix) {
		return nil, nil
	}
	name := strings.TrimPrefix(def, tagPrefix)
	tag := new(Tag)
	tag.Name = name
	tag.Line = idx + 1

	var next token.Element
	for {
		ok = s.Next(&e)
		if !ok {
			break
		}
		var opt Option
		ok = s.Cur(&next)
		if ok && next.Token == token.EQ {
			opt.Key = e.Get()
			s.Next(nil)
			ok = s.Next(&e)
			if ok && (e.Indent || e.String) {
				opt.Value = e.Get()
			}
		} else {
			opt.Value = e.Get()
		}
		tag.Options = append(tag.Options, opt)
	}

	return tag, nil
}
