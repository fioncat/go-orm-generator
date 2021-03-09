package coder

import (
	"fmt"
	"path/filepath"

	"github.com/fioncat/go-gendb/compile/golang"
)

type Import struct {
	imps []*golang.Import
}

func (i *Import) code(c *Coder) bool {
	if len(i.imps) == 0 {
		return false
	}
	if len(i.imps) == 1 {
		imp := i.imps[0]
		c.P(0, "import ", imp.Name, " ", Quote(imp.Path))
		return true
	}

	c.P(0, "import (")
	for _, imp := range i.imps {
		c.P(1, imp.Name, " ", Quote(imp.Path))
	}
	c.P(0, ")")
	return true
}

func (i *Import) Add(name, path string) {
	if name == "" {
		name = filepath.Base(path)
	}
	imp := new(golang.Import)
	imp.Name = name
	imp.Path = path

	i.imps = append(i.imps, imp)
}

func (i *Import) Check() error {
	imps := make([]*golang.Import, 0, len(i.imps))
	m := make(map[string]string, len(i.imps))
	for _, imp := range i.imps {
		if imp.Name == "" {
			imp.Name = filepath.Base(imp.Path)
		}
		path, ok := m[imp.Name]
		if ok {
			if imp.Path == path {
				continue
			}
			return fmt.Errorf(`import "%s" is `+
				`duplcate, please check your config`,
				imp.Name)
		}
		m[imp.Name] = imp.Path
		imps = append(imps, imp)
	}

	i.imps = imps
	return nil
}
