package base

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/misc/errors"
)

// Tag is the custom content inserted in the code file.
// It will be parsed to configure code generation. It is
// usually the content of the comment, and different code
// comments have different formats, so "prefix" must be
// passed in to the functions related to the tag to indicate
// the comment prefix. For example, "//" in golang and
// "--" in sql. Follow the comment prefix with "+gen:{name}",
// so the comment will be parsed as Tag.
type Tag struct {
	// Line represents the line number where the tag is located.
	Line int

	// Name represents the name of the tag, that is, "{name}" in
	// "+gen:{name}". It cannot be empty.
	Name string

	// Options is the configuration of the tag, it is optional,
	// and the format is "{key}={value}".
	// There can be multiple, combined into the form of key-value
	// pairs. The format of the entire tag is:
	// "+gen:{name} {key}={value} {key}={value} ...".
	// {value} can be wrapped in quotation marks to deal with the
	// occurrence of spaces.
	Options []Option
}

// FmtError adds the trace of the current tag on the
// basis of the fmt.Errorf.
func (t *Tag) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(t.Line, err)
}

// Option indicates the specific key-value pair
// configuration in the tag.
type Option struct {
	// Line is equal to the Line of the tag to which
	// the Option belongs.
	Line int

	// The key-value pair content of the Option.
	Key   string
	Value string
}

// Trace adds the trace of the current option on
// the basis of err.
func (o *Option) Trace(err error) error {
	return errors.Trace(o.Line, err)
}

// FmtError adds the trace of the current option on the
// basis of the fmt.Errorf.
func (o *Option) FmtError(a string, b ...interface{}) error {
	err := fmt.Errorf(a, b...)
	return errors.Trace(o.Line, err)
}

var tagTokens = []token.Token{
	token.EQ,
}

// +gen:xxx
const tagPrefix = "+gen:"

// ParseTag parses a line of code into Tag and returns it.
// If the line is not a tag line, "nil, nil" will be returned.
func ParseTag(idx int, prefix, line string) (*Tag, error) {
	if !strings.HasPrefix(line, prefix) {
		return nil, nil
	}
	line = strings.TrimPrefix(line, prefix)
	line = strings.TrimSpace(line)

	// From here on, the line has been considered as a tag line,
	// and it needs to be parsed.
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

	// Parse key-values for the tag.
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

// DecodeOptionFunc is a function used to parse an Option.
type DecodeOptionFunc func(line int, val string, vs []interface{}) error

// DecodeOptionMap is the mapping between Option key and
// parsing function. It is passed into the DecodeTags function
// to call different parsing functions for different options in tags.
type DecodeOptionMap map[string]DecodeOptionFunc

// DecodeTags reads all options in a set of tags, uses the
// name of the option to match the DecodeFunc in doMap, and
// calls the matched DecodeFunc. If the option name is not
// in doMap, an error will be returned.
// "vs" represents the passed parameter, it will be assigned
// to the last parameter of DecodeFunc.
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
