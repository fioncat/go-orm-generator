package coder

type Target interface {
	Name() string
	Path() string

	Imports(imp *Import)

	Vars(c *Var, imp *Import)

	Consts(c *Var, imp *Import)

	Structs(c *StructGroup)
	Funcs(c *FunctionGroup)

	StructNum() int
	Struct(idx int, c *Struct, imp *Import)

	FuncNum() int
	Func(idx int, c *Function, imp *Import)
}
