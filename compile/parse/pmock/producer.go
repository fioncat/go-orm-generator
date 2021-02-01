package pmock

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fioncat/go-gendb/compile/scan/smock"
	"github.com/fioncat/go-gendb/misc/rand"
	"github.com/fioncat/go-gendb/misc/term"
)

// ExecArg is the parameter of a certain epoch in the
// process of executing the mock. It is shared by all
// entities and stores some metadata. Some placeholders
// may need to be used when expanding.
// The life cycle of ExecArg is only one epoch, and the
// arg between different epochs is not shared (each epoch
// will create a brand new ExecArg). In the case of
// concurrency, multiple ExecArg will exist at the same
// time (each worker executing the epoch will hold one),
// so there is no need to worry about concurrent access
// to resources.
type ExecArg struct {
	// Vars saves the data of all mocks in the current
	// epoch. The format of the key is
	// "<entity-name>.<field_name>".
	Vars map[string]string

	// Time is the start time of the current epoch.
	Time time.Time
}

// NewExecArg creates a new empty ExecArg.
func NewExecArg() *ExecArg {
	return &ExecArg{
		Vars: make(map[string]string),
		Time: time.Now(),
	}
}

// The producer corresponds to the body of a scan (smock.Body),
// which is the result of body parsing. In order to maintain
// uniformity, const body will also use producer to indicate
// that every time the producer of const body calls next(),
// the same const result will be produced.
//
// For placeholders, you need to match them according to name()
// and call init() to initialize, so that each call to next may
// return a different value.
//
// Multiple producers can be combined to generate data in the
// format required by the user.
type producer interface {
	// Only applicable to placeholders. Pass in placeholder
	// args(smock.Body.Args) to initialize the producer.
	init(args []string) error

	// Only applicable to placeholders, return the name of
	// the placeholder corresponding to the producer, and
	// use smock.Body.Name and producer.name() to match
	// when parse.
	name() string

	// Produce a value.
	next(arg *ExecArg) string
}

// Call next() of a group of producers, and combine the
// values they produce to return.
func next(ps []producer, arg *ExecArg) string {
	strs := make([]string, len(ps))
	for idx := 0; idx < len(ps); idx++ {
		strs[idx] = ps[idx].next(arg)
	}
	return strings.Join(strs, "")
}

// Placeholder producer names and their type mapping.
// NOTE: The constant producer is not stored here,
// you should use newConstProducer to create the
// constant producer.
var pmap map[string]reflect.Type

// Register a placeholder producer in pmap.
func register(p producer) {
	pmap[p.name()] = reflect.TypeOf(p).Elem()
}

// Create a producer based on the placeholder body,
// which matches the pmap based on body.Name, and
// returns an error if there is no match. And this
// function will directly call producer.init with
// body.Args as a parameter, and an error returned
// by init will also cause the function to return
// an error.
// The producer returned by this function can directly
// use next() to produce values.
func newProducer(body *smock.Body) (producer, error) {
	ptype, ok := pmap[body.Name]
	if !ok {
		return nil, fmt.Errorf(`can not find `+
			`placeholder named "%s"`, body.Name)
	}

	p := reflect.New(ptype).Interface().(producer)
	err := p.init(body.Args)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Initialize all built-in placeholder producers.
func initProducer() {
	pmap = make(map[string]reflect.Type)
	register(&randProducer{})
	register(&incrProducer{})
	register(&inputProducer{})
	register(&varProducer{})
}

// constProducer corresponds to const body, and its
// next() will return a fixed value every time.
type constProducer struct {
	value string
}

func (p *constProducer) init([]string) error  { return nil }
func (p *constProducer) name() string         { return "const" }
func (p *constProducer) next(*ExecArg) string { return p.value }

// newConstProducer creates a new constProducer with const value.
func newConstProducer(val string) producer {
	return &constProducer{value: val}
}

// randProducer is an implementation of "{rand}" placeholder.
type randProducer struct {
	// pick dict
	dict []string

	// Whether the length is fixed, if it is false, the
	// length needs to be randomly generated according to
	// the range of lenStart and lenEnd.
	randLen bool

	lenStart, lenEnd int

	// Whether it is a range, if it is a range, generate
	// an int value from rangeStart to rangeEnd.
	isRange bool

	rangeStart, rangeEnd int
}

func (p *randProducer) name() string {
	return "rand"
}

func (p *randProducer) init(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("rand need at least 2 args")
	}
	switch args[0] {
	case "range":
		p.isRange = true
		if len(args) != 3 {
			return fmt.Errorf("rand:range need 3 args")
		}
		start, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("rand:range: range_start " +
				"bad format, need an integer.")
		}

		end, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("rand:range: range_end " +
				"bad format, need an integer.")
		}

		if start >= end {
			return fmt.Errorf("rand:range: range_start "+
				"must > range_end, found %d vs %d",
				start, end)
		}

		p.rangeStart, p.rangeEnd = start, end

	case "dict":
		p.dict = strings.Split(args[1], ",")
		if len(p.dict) == 0 {
			return fmt.Errorf("rand:dict: dict is empty")
		}
		return p.lenRange(2, args)

	default:
		dictNames := strings.Split(args[0], ",")
		for _, dictName := range dictNames {
			switch dictName {
			case "lcase":
				p.dict = append(p.dict, rand.LCASE_DICT...)
			case "ucase":
				p.dict = append(p.dict, rand.UCASE_DICT...)
			case "case":
				p.dict = append(p.dict, rand.CASE...)
			case "num":
				p.dict = append(p.dict, rand.NUM...)
			default:
				return fmt.Errorf(`unknown built-in dict "%s"`, dictName)
			}
		}
		if len(p.dict) == 0 {
			return fmt.Errorf("rand: dict is empty")
		}
		return p.lenRange(1, args)
	}

	return nil
}

// parse len:
// 3 conditions:
//   empty: start=1, randLen=false
//   1 arg: start=arg0, randLen=false
//   2 args: start=arg0, end=arg1, randLen=true
// If end=0 and rangLen=false, final length=start
func (p *randProducer) lenRange(seed int, args []string) error {
	var err error
	switch len(args) {
	case seed:
		p.randLen = false
		p.lenStart = 1

	case seed + 1:
		p.randLen = false
		p.lenStart, err = strconv.Atoi(args[seed])
		if err != nil {
			return fmt.Errorf("rand:dict: len must be a number")
		}

	case seed + 2:
		p.randLen = true
		p.lenStart, err = strconv.Atoi(args[seed])
		if err != nil {
			return fmt.Errorf("rand:dict: len must be a number")
		}
		p.lenEnd, err = strconv.Atoi(args[seed+1])
		if err != nil {
			return fmt.Errorf("rand:dict: len must be a number")
		}
		if p.lenStart >= p.lenEnd {
			return fmt.Errorf("rand: len_start must < len_end,"+
				" found: %d vs %d", p.lenStart, p.lenEnd)
		}
	}
	return nil
}

func (p *randProducer) next(_ *ExecArg) string {
	// range, directly gen value
	if p.isRange {
		return strconv.Itoa(rand.Range(p.rangeStart, p.rangeEnd))
	}

	len := p.lenStart
	if p.randLen {
		// need to randly gen length.
		len = rand.Range(p.lenStart, p.lenEnd)
	}

	// gen value from dict
	strs := make([]string, len)
	for idx := 0; idx < len; idx++ {
		strs[idx] = rand.Choose(p.dict)
	}

	return strings.Join(strs, "")
}

// incrProducer is an implementation of "{incr}" placeholder.
type incrProducer struct {
	idx, step int32
	mu        sync.Mutex
}

func (p *incrProducer) name() string {
	return "incr"
}

func (p *incrProducer) init(args []string) error {
	userStep := false
	switch len(args) {
	case 0:
		p.idx = 0
		p.step = 1

	case 2:
		userStep = true
		fallthrough

	case 1:
		idx, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("incr: idx is not a number")
		}
		p.step = 1
		p.idx = int32(idx)
	}

	if userStep {
		step, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("incr: step is not a number")
		}
		if p.step == 0 {
			return fmt.Errorf("incr: step can not be zero.")
		}
		p.step = int32(step)
	}

	return nil
}

func (p *incrProducer) next(*ExecArg) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	cur := p.idx
	p.idx += p.step
	return strconv.Itoa(int(cur))
}

// The mapping between the input name and the
// value entered by the user.
var inputs = make(map[string]string)

// inputProducer is an implementation of "{input}"
// placeholder, and its init will ask the user to
// enter a value in the terminal.
// Subsequent input will use the value entered by
// the user (same as constProducer)
type inputProducer struct {
	value string
}

func (p *inputProducer) name() string {
	return "input"
}

func (p *inputProducer) init(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("input: args length must be 1")
	}
	name := args[0]
	val, ok := inputs[name]
	if !ok {
		val = term.Input(`Please input "%s"`, name)
		inputs[name] = val
	}

	p.value = val
	return nil
}

func (p *inputProducer) next(*ExecArg) string {
	return p.value
}

// varProducer is an implementation of "{var}" placeholder.
type varProducer struct {
	key string
}

func (p *varProducer) name() string {
	return "var"
}

func (p *varProducer) init(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("var: args length must be 1")
	}
	name := args[0]
	if name == "" {
		return fmt.Errorf("var: name is empty")
	}
	p.key = name
	return nil
}

func (p *varProducer) next(arg *ExecArg) string {
	return arg.Vars[p.key]
}
