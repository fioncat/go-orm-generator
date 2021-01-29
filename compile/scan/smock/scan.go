package smock

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fioncat/go-gendb/build"
	"github.com/fioncat/go-gendb/compile/scan/stoml"
	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/misc/iter"
)

type Result struct {
	Conn   string `json:"conn"`
	Worker int    `json:"worker"`
	Epoch  int    `json:"epoch"`

	Entities []*Entity `json:"entities"`
}

type Entity struct {
	Name   string   `json:"name"`
	Num    []*Body  `json:"num"`
	Fields []*Field `json:"fields"`
}

const (
	FieldTypeStr = iota
	FieldTypeInt
	FieldTypeFloat
)

type Field struct {
	Type  int     `json:"type"`
	Name  string  `json:"name"`
	Bodys []*Body `json:"bodys"`
}

type Body struct {
	Const string   `json:"const"`
	Name  string   `json:"name"`
	Args  []string `json:"args"`
}

const (
	defaultEpoch = 1

	numKey = "_num_"

	intSuffix   = ":int"
	floatSuffix = ":float"
)

func Do(path, content string) (*Result, error) {
	tr, err := stoml.Do(path, content)
	if err != nil {
		return nil, err
	}

	if len(tr.Sections) == 0 {
		return nil, fmt.Errorf("%s: empty entity", path)
	}

	r := new(Result)
	// Conn, Worker, Epoch
	for _, opt := range tr.Options {
		switch opt.Key {
		case "conn":
			r.Conn = opt.Value

		case "worker", "epoch":
			n, err := strconv.Atoi(opt.Value)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("%s:%d: number bad format",
					path, opt.Line)
			}

			if opt.Key == "worker" {
				r.Worker = n
			}
			if opt.Key == "epoch" {
				r.Epoch = n
			}

		default:
			return nil, fmt.Errorf(`%s:%d: unknown mock option "%s"`,
				path, opt.Line, opt.Key)
		}
	}
	if r.Conn == "" {
		return nil, fmt.Errorf(`%s: missing "conn" option`, path)
	}
	if r.Worker <= 0 {
		r.Worker = build.N_WORKERS
	}
	if r.Epoch <= 0 {
		r.Epoch = defaultEpoch
	}

	r.Entities = make([]*Entity, len(tr.Sections))
	for idx, sec := range tr.Sections {
		e, err := entity(path, sec)
		if err != nil {
			return nil, err
		}
		r.Entities[idx] = e
	}

	return r, nil
}

func entity(path string, sec *stoml.Section) (*Entity, error) {
	e := new(Entity)
	e.Fields = make([]*Field, 0, len(sec.Options))

	var err error
	for _, opt := range sec.Options {
		if opt.Key == numKey {
			e.Num, err = bodys(path, opt)
			if err != nil {
				return nil, err
			}
			continue
		}
		f := new(Field)
		var suffix string
		if strings.HasSuffix(opt.Key, intSuffix) {
			f.Type = FieldTypeInt
			suffix = intSuffix
		} else if strings.HasSuffix(opt.Key, floatSuffix) {
			f.Type = FieldTypeFloat
			suffix = floatSuffix
		} else {
			f.Type = FieldTypeStr
		}

		f.Name = strings.TrimSuffix(opt.Key, suffix)
		f.Bodys, err = bodys(path, opt)
		if err != nil {
			return nil, err
		}

		e.Fields = append(e.Fields, f)
	}

	return e, nil
}

func bodys(path string, opt *stoml.Option) ([]*Body, error) {
	be := &bodyErr{path: path, line: opt.Line}
	iter := iter.New([]rune(opt.Value))

	bs := make([]*Body, 0, 1)

	bucket := make([]rune, 0, 2)
	flush := func() {
		if len(bucket) > 0 {
			b := new(Body)
			b.Const = string(bucket)
			bs = append(bs, b)
			bucket = make([]rune, 0, 1)
		}
	}

	var ch rune
	var idx int
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			flush()
			break
		}

		if token.LBRACE.EqRune(ch) {
			flush()
			ph := new(Body)
			err := _ph(be, iter, ph)
			if err != nil {
				return nil, err
			}
			bs = append(bs, ph)
			continue
		}
		bucket = append(bucket, ch)
	}

	if len(bs) == 0 {
		return nil, be.get(iter.Len(), "empty body")
	}

	return bs, nil
}

func _ph(be *bodyErr, iter *iter.Iter, body *Body) error {
	var ch rune
	var idx int
	bucket := make([]rune, 0, 1)
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			return be.get(iter.Len(), "expect '}'")
		}
		if token.COLON.EqRune(ch) {
			break
		}
		bucket = append(bucket, ch)
	}
	name := string(bucket)
	if name == "" {
		return be.get(idx, "name is empty")
	}
	body.Name = name
	bucket = make([]rune, 0, 1)

	phEnd := false
	for !phEnd {
		idx = iter.NextP(&ch)
		if idx < 0 {
			return be.get(iter.Len(), "expect '}'")
		}
		switch ch {
		case token.RBRACE.Rune():
			phEnd = true
			fallthrough

		case token.COLON.Rune():
			arg := string(bucket)
			if arg != "" {
				body.Args = append(body.Args, arg)
				bucket = make([]rune, 0, 1)
			}

		default:
			bucket = append(bucket, ch)
		}
	}

	return nil
}

type bodyErr struct {
	path string
	line int
}

func (be *bodyErr) get(idx int, msg string) error {
	if idx <= 0 {
		return fmt.Errorf("%s:%d: %s", be.path, be.line, msg)
	}
	return fmt.Errorf("%s:%d:%d: %s", be.path, be.line, idx, msg)
}
