package coder

import "strings"

type FunctionGroup struct {
	funcs []*Function
}

func (g *FunctionGroup) Add() *Function {
	f := new(Function)
	g.funcs = append(g.funcs, f)
	return f
}

func (g *FunctionGroup) Gets() []*Function {
	return g.funcs
}

type Function struct {
	name  string
	comm  string
	def   string
	bodys []string
}

func (f *Function) code(c *Coder) bool {
	if f.comm != "" {
		c.P(0, "// ", f.name, " ", f.comm)
	}
	c.P(0, "func ", f.def, " {")
	for _, body := range f.bodys {
		c.P(0, body)
	}
	c.P(0, "}")
	return true
}

func (f *Function) Comment(s string) {
	f.comm = s
}

func (f *Function) Def(name string, vs ...interface{}) {
	f.name = name
	f.def = joins(vs)
}

func (f *Function) P(n int, vs ...interface{}) {
	n += 1
	prefix := strings.Repeat("\t", n)
	s := joins(vs)
	s = prefix + s
	f.bodys = append(f.bodys, s)
}
