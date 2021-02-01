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

const (
	maxWorker = 300
	maxEpoch  = 50000000
)

// Result represents the result of scanning the mock file.
// A mock file is a file in toml format. Each section
// represents an entity that needs to be mocked. A file can
// contain multiple entities. All entities form a mock epoch.
// There can be multiple mock epochs in total.
//
// In addition to section, mock parameters can be specified
// in the form of global options, where these parameters will
// also be scanned as specific fields.
type Result struct {
	// Path stores the original file path of the mock.
	Path string `json:"path"`

	// Some Global configuration for mock.
	Conn       string `json:"conn"`
	DbType     string `json:"db_type"`
	ExecWorker int    `json:"exec_worker"`
	MockWorker int    `json:"mock_worker"`
	Epoch      int    `json:"epoch"`

	// Entities stores all scanned entities, each entity
	// corresponds to a toml section.
	Entities []*Entity `json:"entities"`
}

// Entity is the entity that needs to be generated in
// the mock process. In the standard RDB, it is a data
// table, which contains multiple fields, each field
// corresponds to an option in toml.
// In each round of mock, an Entity can generate multiple
// records, which is specified by the special field "_num_".
type Entity struct {
	Name   string   `json:"name"`
	Num    []*Body  `json:"num"`
	Fields []*Field `json:"fields"`
}

// Types for field.
const (
	FieldTypeStr = iota
	FieldTypeInt
	FieldTypeFloat
)

// Field represents a specific field, and its content
// can include placeholders to generate dynamic content
// during the mock process.
// Dynamic content will be parsed into the form of multiple
// bodies.
// The data type of the field can be specified by appending
// ":int" or ":float" suffix to the field name. If there is
// a related suffix, the data will be converted to int64 or
// float64 when mocked, and an error will be reported if the
// conversion fails.
type Field struct {
	Line  int     `json:"line"`
	Type  int     `json:"type"`
	Name  string  `json:"name"`
	Bodys []*Body `json:"bodys"`
}

// Body represents the smallest unit of mock content,
// it can be a constant value or a placeholder.
//
// For example, the configured content "aaa{var:b}ccc"
// will be scanned into 3 bodies:
//    1. the first and third are constant bodies, the
//      contents of "aaa" and "ccc" are stored.
//    2. the second is the placeholder body , Name is
//      "var", Args is ["b"].
//
// All the contents of "{xxx}" will be parsed as
// placeholders. If the value does not have a
// placeholder, then the Bodys of the entire field
// should have only one const body.
type Body struct {
	// Const is only used in const body, and judging
	// whether a body is const body can be achieved
	// by judging whether the field is empty.
	Const string `json:"const"`

	// Name is only used in the placeholder body and
	// represents the name of the placeholder.
	Name string `json:"name"`

	// Args is only used in the placeholder body to
	// indicate the configuration of the placeholder.
	Args []string `json:"args"`
}

const (
	defaultEpoch = 1

	numKey = "_num_"

	intSuffix   = ":int"
	floatSuffix = ":float"
)

// Do scans the toml file in mock format, parses the
// configuration in it, and converts it to Result to
// return.
func Do(path, content string) (*Result, error) {
	tr, err := stoml.Do(path, content)
	if err != nil {
		return nil, err
	}

	if len(tr.Sections) == 0 {
		return nil, fmt.Errorf("%s: empty entity", path)
	}

	r := new(Result)
	r.Path = path
	// Conn, Worker, Epoch
	for _, opt := range tr.Options {
		switch opt.Key {
		case "conn":
			r.Conn = opt.Value

		case "db_type":
			r.DbType = opt.Value

		case "exec_worker", "mock_worker", "epoch":
			n, err := strconv.Atoi(opt.Value)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("%s:%d: number bad format",
					path, opt.Line)
			}

			if opt.Key == "exec_worker" {
				r.ExecWorker = n
			}
			if opt.Key == "mock_worker" {
				r.MockWorker = n
			}
			if opt.Key == "epoch" {
				r.Epoch = n
			}

		default:
			return nil, fmt.Errorf(`%s:%d: unknown mock option "%s"`,
				path, opt.Line, opt.Key)
		}
	}
	if r.DbType == "" {
		r.DbType = "mysql"
	}
	if r.ExecWorker <= 0 {
		r.ExecWorker = build.N_WORKERS
	}
	if r.MockWorker <= 0 {
		r.MockWorker = 1
	}
	if r.Epoch <= 0 {
		r.Epoch = defaultEpoch
	}

	if r.Epoch > maxEpoch {
		return nil, fmt.Errorf(`epoch is too big, max %d`, maxEpoch)
	}
	if r.ExecWorker > maxWorker {
		return nil, fmt.Errorf(`exec worker is too big, max %d`, maxWorker)
	}
	if r.MockWorker > maxWorker {
		return nil, fmt.Errorf(`mock worker is too big, max %d`, maxWorker)
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

// toml section -> Entity
func entity(path string, sec *stoml.Section) (*Entity, error) {
	e := new(Entity)
	e.Fields = make([]*Field, 0, len(sec.Options))
	e.Name = sec.Name

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
		f.Line = opt.Line

		e.Fields = append(e.Fields, f)
	}

	if len(e.Num) == 0 {
		e.Num = []*Body{{Const: "1"}}
	}

	return e, nil
}

// toml option -> Bodys
func bodys(path string, opt *stoml.Option) ([]*Body, error) {
	be := &bodyErr{path: path, line: opt.Line}
	iter := iter.New([]rune(opt.Value))

	bs := make([]*Body, 0, 1)

	bucket := make([]rune, 0, 2)
	flush := func() {
		// if there is const content in front, save it
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

		// "{name:args...}" placeholder
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

// parse "{name:[args...]}" placeholder, convert it into
// Body object and append to body.
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

// bodyErr builder, build scan body compile error.
type bodyErr struct {
	path string
	line int
}

// get build the error
func (be *bodyErr) get(idx int, msg string) error {
	if idx <= 0 {
		return fmt.Errorf("%s:%d: %s", be.path, be.line, msg)
	}
	return fmt.Errorf("%s:%d:%d: %s", be.path, be.line, idx, msg)
}
