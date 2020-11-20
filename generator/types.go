package generator

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/misc/col"
)

type Task struct {
	Path string
	Type string
	Pkg  string

	Imports []Import

	Structs    []TaskStruct
	Interfaces []TaskInterface
}

type Import struct {
	Name string
	Path string
}

func (i *Import) String() string {
	return fmt.Sprintf(`%s "%s"`, i.Name, i.Path)
}

type TaskStruct struct {
	Options []TaskOption
	Name    string
	Args    []string
}

func (t *TaskStruct) OptionNames() []string {
	return OptionNames(t.Options)
}

type TaskInterface struct {
	Options []TaskOption
	Name    string
	Methods []TaskMethod
	Args    []string
}

func (t *TaskInterface) OptionNames() []string {
	return OptionNames(t.Options)
}

type TaskMethod struct {
	Name string

	Params     string
	ParamNames col.Set

	ReturnType string

	Pkgs col.Set

	Options []TaskOption
}

func (m *TaskMethod) String() string {
	return fmt.Sprintf("%s(%s) (%s, error)",
		m.Name, m.Params, m.ReturnType)
}

func (m *TaskMethod) OptionNames() []string {
	return OptionNames(m.Options)
}

type TaskOption struct {
	Line int
	Tag  string
	Args []string
}

func OptionNames(opts []TaskOption) []string {
	names := make([]string, len(opts))
	for i, opt := range opts {
		name := fmt.Sprintf("%s[%s]", opt.Tag, strings.Join(opt.Args, ","))
		names[i] = name
	}
	return names
}
