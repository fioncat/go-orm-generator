package pmock

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/fioncat/go-gendb/compile/scan/smock"
	"github.com/fioncat/go-gendb/database/rdb"
)

// Result is the result of parsing the mock toml file.
// Its structure is very similar to smock.Result. The
// difference is that the smock.Body is replaced by a
// producer, and the data can be mocked through the
// Next function.
type Result struct {
	// ExecWorker represents the number of workers that
	// execute sql statements concurrently.
	ExecWorker int

	// MockWorker represents the number of workers that
	// execute mock concurrently.
	MockWorker int

	// Epoch represents the total number of epochs that
	// need to be executed.
	Epoch int

	// Entities represents all entities to be mocked.
	Entities []*Entity
}

// The global empty execArg, it will be submitted to
// Num producers for use. Because at the time of mock
// num, the mock has not yet started.
var emptyExecArg = new(ExecArg)

// Entity represents the entity to be mocked, and its
// result is very similar to smock.Entity. The difference
// is that smock.Body is replaced with producer to produce
// mock data.
type Entity struct {
	Name   string
	Fields []*Field

	num []producer
}

// Num produce the number of records to mock. It will be
// converted to an integer and returned, and an error will
// be returned directly if the conversion fails.
func (e *Entity) Num() (int64, error) {
	s := next(e.num, emptyExecArg)
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(`%s: parse num `+
			`failed: "%s" is not a number`, e.Name, s)
	}
	return num, nil
}

// Field represents the field to be mocked. It contains a
// set of producers, which can be used to produce mock data
// by calling Next().
type Field struct {
	Name string

	table string
	ftype int
	ps    []producer
}

// Next produce the next mock data. The format of  the specific
// data is configured by mock rules containing placeholders.
// The field may specify an int or float type. If specified,
// the function will also convert the produced data to the
// corresponding type and return. If the conversion fails, an
// error will be returned.
func (f *Field) Next(arg *ExecArg) (interface{}, error) {
	s := next(f.ps, arg)
	switch f.ftype {
	case smock.FieldTypeStr:
		return s, nil

	case smock.FieldTypeInt:
		intVal, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(`%s: int field "%s"`+
				` value "%s" is not a number`, f.table, f.Name, s)
		}
		return intVal, nil

	case smock.FieldTypeFloat:
		floatVal, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf(`%s: float field "%s"`+
				` value "%s" is not a number`, f.table, f.Name, s)
		}
		return floatVal, nil
	}
	return s, nil
}

var initOnce sync.Once

// Do convert smock.Result to Result and return. This process
// will really parse the mock placeholders. The results can be
// directly used in the production of mock data.
func Do(sr *smock.Result) (*Result, error) {
	var err error
	if sr.Conn != "" {
		err = rdb.Init(sr.Conn, sr.DbType)
		if err != nil {
			return nil, err
		}
	}
	initOnce.Do(initProducer)

	r := new(Result)
	r.ExecWorker = sr.ExecWorker
	r.MockWorker = sr.MockWorker
	r.Epoch = sr.Epoch

	r.Entities = make([]*Entity, len(sr.Entities))
	for idx, se := range sr.Entities {
		e := new(Entity)
		e.Name = se.Name
		e.num, err = bodys2producers(se.Num)
		if err != nil {
			return nil, fmt.Errorf("%s:%s:_num_: %v",
				sr.Path, e.Name, err)
		}
		e.Fields = make([]*Field, len(se.Fields))
		for j, sf := range se.Fields {
			f := new(Field)
			f.Name = sf.Name
			f.table = e.Name
			f.ftype = sf.Type
			f.ps, err = bodys2producers(sf.Bodys)
			if err != nil {
				return nil, fmt.Errorf("%s:%s:%s(line:%d): %v",
					sr.Path, e.Name, sf.Name, sf.Line, err)
			}
			e.Fields[j] = f
		}
		r.Entities[idx] = e
	}

	return r, nil
}

func bodys2producers(bodys []*smock.Body) ([]producer, error) {
	ps := make([]producer, len(bodys))
	for idx, body := range bodys {
		if body.Const != "" {
			ps[idx] = newConstProducer(body.Const)
			continue
		}
		p, err := newProducer(body)
		if err != nil {
			return nil, err
		}
		ps[idx] = p
	}
	return ps, nil
}
