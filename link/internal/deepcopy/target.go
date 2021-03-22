package deepcopy

import (
	"fmt"
	"path/filepath"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/deepcopy"
	"github.com/fioncat/go-gendb/compile/golang"
)

func newTarget(file *golang.File, r *deepcopy.Result) coder.Target {
	t := new(target)
	t.path = file.Path
	t.r = r

	t.importMap = make(map[string]string, len(file.Imports))
	for _, imp := range file.Imports {
		if imp.Name == "" {
			imp.Name = filepath.Base(imp.Path)
		}
		t.importMap[imp.Name] = imp.Path
	}
	return t
}

type target struct {
	coder.NoVars
	coder.NoConsts
	coder.NoStructs
	coder.NoFuncs
	coder.NoStructNum
	coder.NoImports

	path string

	r *deepcopy.Result

	importMap map[string]string
}

func (t *target) Name() string {
	return "ex.deepcopy." + t.r.Name
}

func (t *target) Path() string {
	return t.path
}

func (t *target) FuncNum() int { return 1 }

func (t *target) Func(_ int, c *coder.Function, ic *coder.Import) {
	for _, imp := range t.r.Imports {
		ic.Add("", imp)
	}

	c.Def(t.r.Method, "(o *", t.r.Name, ") ",
		"", t.r.Method, " (res *", t.r.Target, ")")
	c.P(0, "res = new(", t.r.Target, ")")
	for _, f := range t.r.Fields {
		t.field(f, c, ic)
	}
	c.P(0, "return")
}

func (t *target) field(f *deepcopy.Field, c *coder.Function, ic *coder.Import) {
	var left string
	if f.Left != "" {
		left = f.Left
	} else {
		left = f.Name
	}
	left = fmt.Sprintf("res.%s", left)
	if f.Set != "" && f.Type.Slice == nil && f.Type.Map == nil {
		right := base.ReplacePlaceholder(
			f.Set, "o."+f.Name, "o", f.Name)
		c.P(0, left, " = ", right)
		return
	}
	var pkg string
	if f.Type.Slice != nil {
		pkg = f.Type.Slice.Package
	} else if f.Type.Map != nil {
		pkg = f.Type.Map.Value.Package
	} else {
		pkg = f.Type.Package
	}
	if pkg != "" {
		path := t.importMap[pkg]
		if path != "" {
			ic.Add(f.Type.Package, path)
		}
	}
	right := fmt.Sprintf("o.%s", f.Name)
	if f.Type.Slice != nil {
		copySlice(0, c, 0, left, right,
			f.Type.Slice, f.Set, t.r.Method)
		return
	}
	if f.Type.Map != nil {
		copyMap(0, c, 0, left, right,
			f.Type.Map, f.Set, t.r.Method)
		return
	}
	if !f.Type.Simple {
		right += "." + t.r.Method + "()"
	}
	c.P(0, left, " = ", right)

}

func copySlice(nTab int,
	c *coder.Function,
	depth int,
	a, b string,
	slice *golang.DeepType,
	set, method string,
) {

	idxStr := fmt.Sprintf("idx%d", depth)
	eleStr := fmt.Sprintf("e%d", depth)
	var eq string
	if depth == 0 {
		eq = "="
	} else {
		eq = ":="
	}
	c.P(nTab, a, " ", eq, " make([]", slice.Full, ", len(", b, "))")
	c.P(nTab, "for ", idxStr, ", ", eleStr, " := range ", b, " {")
	if slice.Slice != nil {
		nexta := fmt.Sprintf("slice%d", depth)
		copySlice(nTab+1, c, depth+1, nexta, eleStr, slice.Slice, set, method)
		c.P(nTab+1, a, "[", idxStr, "]", " = ", nexta)
	} else if slice.Map != nil {
		nexta := fmt.Sprintf("m%d", depth)
		copyMap(nTab+1, c, depth+1, nexta, eleStr, slice.Map, set, method)
		c.P(nTab+1, a, "[", idxStr, "]", " = ", nexta)
	} else {
		left := getLeft(slice.Simple, eleStr, set, method)
		c.P(nTab+1, a, "[", idxStr, "]", " = ", left)
	}
	c.P(nTab, "}")
}

func copyMap(nTab int,
	c *coder.Function,
	depth int,
	a, b string,
	m *golang.DeepMap,
	set, method string,
) {

	keyStr := fmt.Sprintf("key%d", depth)
	valStr := fmt.Sprintf("val%d", depth)
	var eq string
	if depth == 0 {
		eq = "="
	} else {
		eq = ":="
	}
	c.P(nTab, a, " ", eq, " make(", m.Full, ", len(", b, "))")
	c.P(nTab, "for ", keyStr, ", ", valStr, " := range ", b, " {")
	if m.Value.Slice != nil {
		nexta := fmt.Sprintf("slice%d", depth)
		copySlice(nTab+1, c, depth+1, nexta, valStr, m.Value.Slice, set, method)
		c.P(nTab+1, a, "[", keyStr, "]", " = ", nexta)
	} else if m.Value.Map != nil {
		nexta := fmt.Sprintf("m%d", depth)
		copyMap(nTab+1, c, depth+1, nexta, valStr, m.Value.Map, set, method)
		c.P(nTab+1, a, "[", keyStr, "]", " = ", nexta)
	} else {
		left := getLeft(m.Value.Simple, valStr, set, method)
		c.P(nTab+1, a, "[", keyStr, "]", " = ", left)
	}
	c.P(nTab, "}")
}

func getLeft(isSimple bool, val, set, method string) string {
	var left string
	if isSimple {
		left = val
	} else {
		if set != "" {
			left = base.ReplacePlaceholder(set, val)
		} else {
			left = val + "." + method + "()"
		}
	}
	return left
}

func getAssign(src string, t *golang.DeepType) string {
	if t.Simple {
		return src
	}
	return src + ".DeepCopy()"
}
