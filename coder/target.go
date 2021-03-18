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

type NoFuncNum struct{}

func (*NoFuncNum) FuncNum() int { return 0 }

func (*NoFuncNum) Func(idx int, c *Function, imp *Import) {}

type NoStructNum struct{}

func (*NoStructNum) StructNum() int { return 0 }

func (*NoStructNum) Struct(idx int, s *Struct, imp *Import) {}

type NoFuncs struct{}

func (*NoFuncs) Funcs(c *FunctionGroup) {}

type NoStructs struct{}

func (*NoStructs) Structs(c *StructGroup) {}

type NoConsts struct{}

func (*NoConsts) Consts(c *Var, imp *Import) {}

type NoVars struct{}

func (*NoVars) Vars(c *Var, imp *Import) {}

type NoImports struct{}

func (*NoImports) Vars(imp *Import) {}
