package coder

import (
	"fmt"
	"strconv"
)

type Var struct {
	Const bool

	groups []*VarGroup
}

func (v *Var) code(c *Coder) bool {
	if len(v.groups) == 0 {
		return false
	}
	for idx, gp := range v.groups {
		if len(gp.names) == 0 {
			continue
		}
		if idx != 0 {
			c.Empty()
		}
		if gp.comm != "" {
			c.P(0, "// ", gp.comm)
		}
		var prefix string
		if v.Const {
			prefix = "const"
		} else {
			prefix = "var"
		}
		if len(gp.names) == 1 {
			name, val := gp.names[0], gp.vals[0]
			c.P(0, prefix, " ", name, " = ", val)
			continue
		}
		var nameLen int
		for _, name := range gp.names {
			if len(name) > nameLen {
				nameLen = len(name)
			}
		}
		nameFmt := "%-" + strconv.Itoa(nameLen) + "s"

		c.P(0, prefix, " (")
		for idx, name := range gp.names {
			val := gp.vals[idx]
			name = fmt.Sprintf(nameFmt, name)
			c.P(1, name, " = ", val)
		}
		c.P(0, ")")
	}
	return true
}

func (v *Var) NewGroup() *VarGroup {
	g := new(VarGroup)
	v.groups = append(v.groups, g)
	return g
}

type VarGroup struct {
	comm  string
	names []string
	vals  []string
}

func (g *VarGroup) Comment(vs ...interface{}) {
	g.comm = joins(vs)
}

func (g *VarGroup) Add(name string, vs ...interface{}) {
	val := joins(vs)
	g.names = append(g.names, name)
	g.vals = append(g.vals, val)
}
