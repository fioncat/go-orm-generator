package cmdt

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fioncat/go-gendb/misc/term"
)

// ActionFunc is the type of function executed by each command.
// Its parameters represent the input of the command line,
// that is, the command line parameters passed in the terminal.
// These parameters will be converted into a structure (the
// specific structure is specified when the Command is defined
// by "Pv" field). Note that you generally need to convert "p"
// to a specific type. An error can be returned during execution.
// If the function returns an error, the program will exit with 1.
// If it returns nil, the program will exit with 0.
type ActionFunc func(p interface{}) error

// Command represents a specific terminal command. It contains
// a set of parameters and execution functions. In addition,
// there are data such as the name, Usage, and Help documents
// used to identify and describe it. These descriptive data are
// used to generate help documents to show to the command line
// tool users.
//
// For command parameters, they should be passed in via "Pv"
// in the form of "(*<struct>)nil". The field of this structure
// should be tagged with "flag" and "arg", the field tagged by
// "flag" represents the "--<flag> <value>" parameter of the
// command line; the field tagged by "arg" represents "<value>"
// parameter. In addition, the type of "flag" can only be string,
// int, bool (bool has no "<value>"), and "arg" can only be of
// string type.
type Command struct {
	// Name is the unique identifier of the command,
	// and the user needs to call the specified
	// command by name.
	Name string

	// Pv represents command line parameters, and its
	// form is "(*<struct>)nil". The parameters in
	// "<struct>" need to be marked with "flag" and "arg".
	Pv interface{}

	// Usage and Help for this command.
	Usage string
	Help  string

	// Action represents the specific function executed
	// by the command.
	Action ActionFunc

	// parse paramter type once
	parceOnce sync.Once

	// the paramter type and fieldName-fieldIndex map
	p  reflect.Type
	fm map[string]int

	// the parsed flags and args.
	flags map[string]*flag
	args  []*arg
}

// _field represents a Go field
type _field struct {
	goName string
	goType string
}

// flag represents a parsed command flag.
type flag struct {
	_field

	name     string
	_default interface{}

	set bool
}

// arg represent a parsed command arg.
type arg struct {
	_field
	name string
}

// parse the "Pv", use relfect to convert it into flags and args.
// the field must meet the requirements, otherwise the function
// will panic.
// For command, this function should be executed once during
// initialization.
func (cmd *Command) parsePv() {
	cmd.p = reflect.TypeOf(cmd.Pv).Elem()

	nFields := cmd.p.NumField()
	cmd.fm = make(map[string]int, nFields)
	cmd.flags = make(map[string]*flag)

	for i := 0; i < nFields; i++ {
		field := cmd.p.Field(i)
		// the field's index
		cmd.fm[field.Name] = i

		_field := _field{
			goName: field.Name,
			goType: field.Type.Name(),
		}

		// get "flag" tag
		flagName := field.Tag.Get("flag")
		if flagName == "" {
			// not a flag, it is an arg
			argName := field.Tag.Get("arg")
			if argName == "" {
				panic("field " + field.Name +
					" is neighter flag nor tag")
			}
			// arg's go type must be string
			if _field.goType != "string" {
				panic("arg " + field.Name +
					"'s type is not a string")
			}

			cmd.args = append(cmd.args, &arg{
				_field: _field,
				name:   argName,
			})
			continue
		}

		var flag flag
		flag.name = flagName
		// flag's type must be string,int,bool
		switch _field.goType {
		case "string", "int", "bool":
			flag._field = _field
		default:
			panic("unsupport flag type: " + _field.goType)
		}

		// parse default value(if given)
		defaultVal := field.Tag.Get("default")
		if defaultVal != "" {
			switch _field.goType {
			case "bool":
				panic("bool no support default value")

			case "string":
				flag._default = defaultVal

			case "int":
				intVal, err := strconv.Atoi(defaultVal)
				if err != nil {
					panic("parse int " + defaultVal + " failed")
				}
				flag._default = intVal
			}
		}

		cmd.flags[flag.name] = &flag
	}
}

// parseArgs parse command args into Go map.
// For example, the args is ["--enable-log", "--path", "/path/to/run", "hello"]
// The deinition of the parameter struct is:
//    type param struct {
//      EnableLog bool   `flag:"enable-log"`
//      Path      string `flag:"path"`
//      Info      string `flag:"info"`
//    }
// Then the function will return:
// map{ "EnableLog": true, "Path": "/path/to/run", "Info": "hello" }
func (cmd *Command) parseArgs(osargs []string) (map[string]interface{}, error) {
	cmd.parceOnce.Do(cmd.parsePv)

	argIdx := 0
	values := make(map[string]interface{}, len(osargs))

	for idx, osArg := range osargs {
		if osArg == "" {
			continue
		}

		if !strings.HasPrefix(osArg, "-") {
			// Does not start with "-",
			// indicating that this is an arg
			if argIdx >= len(cmd.args) {
				return argOutofRange(osArg, idx, len(cmd.args))
			}

			arg := cmd.args[argIdx]
			values[arg.goName] = osArg
			argIdx += 1

			continue
		}

		// Starts with "-"(or "--"), this arg is a flag
		// For typeof string/int, syntax is "--flag {value}"
		// For typeof bool, syntax is "--flag"
		flagName := strings.TrimLeft(osArg, "-")
		flag := cmd.flags[flagName]
		if flag == nil {
			// can not find this flag
			return unknownFlag(flagName)
		}
		flag.set = true

		if flag.goType == "bool" {
			// "--flag" means flag is true
			values[flag.goName] = true
			continue
		}

		// for string/int, we need to take out next osArg
		// it is current flag's value.
		var flagValue string
		if idx < len(osargs)-1 {
			nextValue := osargs[idx+1]
			if !strings.HasPrefix(nextValue, "-") {
				// the next osArg is a value,
				// we need to set it to empty to prevent
				// treating it as an arg in next loop
				flagValue = nextValue
				osargs[idx+1] = ""
			}
		}
		if flagValue == "" {
			// now we donot allow empty value for flag
			// TODO: maybe it is harmless to allow???
			return flagEmptyErr(flagName)
		}

		switch flag.goType {
		case "string":
			values[flag.goName] = flagValue

		case "int":
			intVal, err := strconv.Atoi(flagValue)
			if err != nil {
				return flagFormatErr(flagName)
			}
			values[flag.goName] = intVal
		}
	}

	if argIdx != len(cmd.args) {
		return missingArgErr(cmd.args[argIdx].name)
	}

	// For those flags with set=false, and has default
	// value, we need to add default value.
	for _, flag := range cmd.flags {
		if !flag.set {
			if flag._default != nil {
				values[flag.goName] = flag._default
			}
		} else {
			flag.set = false
		}
	}
	return values, nil
}

func missingArgErr(name string) (_ map[string]interface{}, err error) {
	err = fmt.Errorf(`missing arg "%s"`, name)
	return
}

func unknownFlag(name string) (_ map[string]interface{}, err error) {
	err = fmt.Errorf(`unknown flag "%s"`, name)
	return
}

func flagFormatErr(name string) (_ map[string]interface{}, err error) {
	err = fmt.Errorf(`flag "%s" format error`, name)
	return
}

func flagEmptyErr(name string) (_ map[string]interface{}, err error) {
	err = fmt.Errorf(`flag "%s" is empty`, name)
	return
}

func argOutofRange(arg string, idx, len int) (_ map[string]interface{}, err error) {
	err = fmt.Errorf(`arg "%s" index %d is out of range, arg length = %d`, arg, idx, len)
	return
}

// convert os Arg into paramter structure.
func (cmd *Command) parseP(args []string) (interface{}, error) {
	values, err := cmd.parseArgs(args)
	if err != nil {
		return nil, err
	}

	pv := reflect.New(cmd.p)
	v := pv.Elem()

	for name, value := range values {
		fieldIdx, ok := cmd.fm[name]
		if !ok {
			return nil, fmt.Errorf(`can not find flag `+
				`"%s" in struct`, name)
		}
		field := v.Field(fieldIdx)
		switch value.(type) {
		case int:
			field.SetInt(int64(value.(int)))
		case string:
			field.SetString(value.(string))
		case bool:
			field.SetBool(value.(bool))
		}
	}

	return pv.Interface(), nil
}

// ShowUsage outputs the Usage of the command in the terminal.
func (cmd *Command) ShowUsage() {
	fmt.Printf("Usage: go-gendb %s\n", cmd.Usage)
	fmt.Printf("Use \"go-gendb help %s\" "+
		"for more information\n", cmd.Name)
}

// ShowHelp outputs the help document of the command in
// the terminal, including Usage and Help data.
func (cmd *Command) ShowHelp() {
	fmt.Printf("Usage: go-gendb %s\n", cmd.Usage)
	fmt.Println(cmd.Help)
}

// Execute executes the command, regardless of the normal
// end of the command or error termination, it will directly
// exit the program. It should be the last function executed
// by main, and it should be executed after all initialization
// processes are over.
func (cmd *Command) Execute(args []string) {
	v, err := cmd.parseP(args)
	if err != nil {
		fmt.Printf("%s %v\n", term.Red("[arg error]"), err)
		cmd.ShowUsage()
		os.Exit(1)
	}

	err = cmd.Action(v)
	if err != nil {
		fmt.Printf("%s %v\n", term.Red("[error]"), err)
		os.Exit(1)
	}
	os.Exit(0)
}
