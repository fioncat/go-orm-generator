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
	Line int

	Key   string
	Value string
}

func (o *Option) Trace(err error) error {
	return errors.Trace(o.Line, err)
}

func (o *Option) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(o.Line, err)
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
		opt.Line = tag.Line
		tag.Options = append(tag.Options, opt)
	}

	return tag, nil
}

type DecodeOptionFunc func(line int, val string, vs []interface{}) error

type DecodeOptionMap map[string]DecodeOptionFunc

func DecodeTags(tags []*Tag, name string, doMap DecodeOptionMap, vs ...interface{}) error {
	for _, tag := range tags {
		if tag.Name != name {
			continue
		}
		for _, opt := range tag.Options {
			opt := opt
			doFunc := doMap[opt.Key]
			if doFunc == nil {
				return tag.FmtError(`unknown option "%s"`, opt.Key)
			}
			err := doFunc(opt.Line, opt.Value, vs)
			if err != nil {
				return opt.Trace(err)
			}
		}
	}
	return nil
}
