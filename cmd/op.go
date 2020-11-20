package cmd

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fioncat/go-gendb/misc/errors"
)

type ActionFunc func(ctx *Context) bool

type Operation struct {
	ParamPtr interface{}

	Action ActionFunc

	parseOnce sync.Once

	paramType reflect.Type
	fieldsMap map[string]int

	flags map[string]*Flag
	args  []*Arg
}

func (o *Operation) parseType() {
	o.paramType = reflect.TypeOf(o.ParamPtr).Elem()

	nFields := o.paramType.NumField()
	o.fieldsMap = make(map[string]int, nFields)
	o.flags = make(map[string]*Flag)
	for i := 0; i < nFields; i++ {
		field := o.paramType.Field(i)
		o.fieldsMap[field.Name] = i

		flagName := field.Tag.Get("flag")
		param := param{
			goName: field.Name,
			goType: field.Type.Name(),
		}
		if flagName == "" {
			// 表示这是一个arg
			argName := field.Tag.Get("arg")
			if argName == "" {
				panic("missing arg or flag tag for " + field.Name)
			}
			if param.goType != "string" {
				panic("arg type can only be string " + field.Name)
			}
			arg := &Arg{Name: argName, param: param}
			o.args = append(o.args, arg)
			continue
		}

		flag := new(Flag)
		flag.Name = flagName
		// 这是一个flag，先检查类型，只能是int,string,bool
		switch param.goType {
		case "int", "string", "bool":
			flag.param = param
		default:
			panic("invalid type for " + field.Name)
		}

		// 默认值
		defaultVal := field.Tag.Get("default")
		flag.Default = nil
		if defaultVal != "" {
			if flag.IsBool() {
				panic("bool not support default value")
			}
			value := flag.getValue(defaultVal)
			if value == nil {
				panic("default value bad format " + field.Name)
			}
			flag.Default = value
		}

		o.flags[flag.Name] = flag
	}
}

func (o *Operation) ParseArgs(argStrs []string) (map[string]interface{}, error) {
	o.parseOnce.Do(o.parseType)
	argIdx := 0
	values := make(map[string]interface{}, len(argStrs))
	for i, argStr := range argStrs {
		if argStr == "" {
			continue
		}
		if !strings.HasPrefix(argStr, "-") {
			// 不以"-"开头，说明这是一个arg
			if argIdx >= len(o.args) {
				// arg已经达到上限了，说明当前这个是溢出的arg
				// 这里报错返回
				return argOutofRange(argStr, i, len(o.args))
			}
			arg := o.args[argIdx]
			values[arg.goName] = argStr
			argIdx += 1
			continue
		}

		flagName := strings.TrimLeft(argStr, "-")
		flag := o.flags[flagName]
		if flag == nil {
			return unknownFlag(argStr)
		}
		flag.isSet = true

		// 如果是bool值，出现了这个flag就表明直接是true
		if flag.IsBool() {
			values[flag.goName] = true
			continue
		}

		// 要从下一个argStr拿到这个flag的value字符串
		var flagValue string
		if i < len(argStrs)-1 {
			nextValue := argStrs[i+1]
			if !strings.HasPrefix(nextValue, "-") {
				// 下一个是当前flag的value, 取出它
				// 并在argStrs置空, 防止扫描到下一个
				// 的时候把它当做arg处理了
				flagValue = nextValue
				argStrs[i+1] = ""
			}
		}
		if flagValue == "" {
			return flagEmptyErr(argStr)
		}

		if value := flag.getValue(flagValue); value != nil {
			values[flag.goName] = value
		} else {
			return flagFormatErr(argStr)
		}
	}

	if argIdx != len(o.args) {
		return missingArgErr(o.args[argIdx].Name)
	}

	// 对于那些isSet=false, 并且有默认值的flag
	// 要把默认值给加上
	for _, flag := range o.flags {
		if !flag.isSet {
			if flag.Default != nil {
				values[flag.goName] = flag.Default
			}
		} else {
			//重置一下
			flag.isSet = false
		}
	}

	return values, nil
}

func (f *Flag) getValue(valStr string) interface{} {
	// 如果是int类型，还需要去解析int
	if f.IsInt() {
		intValue, err := strconv.Atoi(valStr)
		if err != nil || intValue < 0 {
			return nil
		}
		return intValue
	}
	return valStr
}

func (o *Operation) ParseParam(args []string) (interface{}, error) {
	values, err := o.ParseArgs(args)
	if err != nil {
		return nil, err
	}
	pv := reflect.New(o.paramType)
	v := pv.Elem()

	for name, value := range values {
		fieldIdx, ok := o.fieldsMap[name]
		if !ok {
			return nil, errors.Fmt("can not find "+
				"flag '%s' in struct", name)
		}
		field := v.Field(fieldIdx)
		switch value.(type) {
		case int:
			field.SetInt(int64(value.(int)))
		case string:
			field.SetString(value.(string))
		case bool:
			field.SetBool(value.(bool))
		default:
			return nil, errors.Fmt("unsupport type "+
				"for value '%v'", value)
		}
	}
	return pv.Interface(), nil
}

type param struct {
	goName string
	goType string
}

type Flag struct {
	param
	Name    string
	Default interface{}

	isSet bool
}

func (f *Flag) IsBool() bool {
	return f.goType == "bool"
}

func (f *Flag) IsInt() bool {
	return f.goType == "int"
}

type Arg struct {
	param

	Name string
}

func missingArgErr(name string) (_ map[string]interface{}, err error) {
	err = errors.Fmt(`missing arg "%s"`, name)
	return
}

func unknownFlag(name string) (_ map[string]interface{}, err error) {
	err = errors.Fmt(`unknown flag "%s"`, name)
	return
}

func flagFormatErr(name string) (_ map[string]interface{}, err error) {
	err = errors.Fmt(`flag "%s" format error`, name)
	return
}

func flagEmptyErr(name string) (_ map[string]interface{}, err error) {
	err = errors.Fmt(`flag "%s" is empty`, name)
	return
}

func argOutofRange(arg string, idx, len int) (_ map[string]interface{}, err error) {
	err = errors.Fmt(`arg "%s" index %d is out of range, arg length = %d`, arg, idx, len)
	return
}

var genop *Operation

func Register(op *Operation) {
	genop = op
}
