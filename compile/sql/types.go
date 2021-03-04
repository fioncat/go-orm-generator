package sql

type Method struct {
	Inter string
	Name  string

	State *Statement

	Exec bool

	Dyn bool
	Dps []*DynamicPart

	Fields []*QueryField
}

// dynamic type enum
const (
	DynamicTypeConst = iota
	DynamicTypeIf
	DynamicTypeFor
)

type DynamicPart struct {
	Type int

	State *Statement

	IfCond string

	ForEle   string
	ForSlice string
	ForJoin  string
}

type Statement struct {
	Sql string

	Replaces []string
	Prepares []string

	phs []*placeholder
}

type placeholder struct {
	pre  bool
	name string
}

func (ph *placeholder) String() string {
	if ph.pre {
		return "$" + ph.name
	}
	return "#" + ph.name
}

type Var struct {
}

type QueryField struct {
	Table string
	Name  string
	Alias string

	IsCount bool
}
