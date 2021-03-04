package golang

import (
	"fmt"
	"testing"
)

func TestParseSingleImport(t *testing.T) {
	var lines = []string{
		`import "github.com/a/b"`,
		`import user "aa/bb/user"`,
		`import "fmt/base`,
		`var a string`,
		``,
		`const bbb int64`,
	}

	p := new(_singleImportParser)

	for idx, line := range lines {
		ok := p.Accept(line)
		if ok {
			v, err := p.Do(line)
			if err != nil {
				fmt.Printf("idx=%d, err=%v\n", idx, err)
				return
			}
			imp := v.(*Import)
			fmt.Printf("import %s %s\n", imp.Name, imp.Path)
		}
	}

}

func TestParseMultiImports(t *testing.T) {
	var lines = []string{
		"import (",
		`aaa "github.com/a/b/aab"`,
		`"fmt"`,
		`"strings"`,
		`"testing"`,
		`"model "d/a""`,
		")",
	}
	p := acceptPaths(lines[0])
	if p == nil {
		fmt.Println("imports is bad")
		return
	}
	lines = lines[1:]
	for idx, line := range lines {
		ok, err := p.Next(line, nil)
		if err != nil {
			fmt.Printf("idx=%d, err=%v\n", idx, err)
		}
		if !ok {
			break
		}
	}

	imps := p.Get().([]Import)
	for _, imp := range imps {
		fmt.Printf("%s %s\n", imp.Name, imp.Path)
	}
}

func TestParseInterface(t *testing.T) {
	var lines = []string{
		"type User interface {",
		"    Add(db *sql.DB, u *model.User) (int64, error)",
		"    FindById(db *sql.DB, id int64) (*model.User, error)",
		"    FindAll(db *sql.DB, query []string) ([]*model.User, error)",
		"    Search(query []string) ([]model.User, error)",
		"    Search(qs []string) ([]*User, error)",
		"    Do(a runner.Cond) (User, error)",
		"    Do2(b runner.Many) ([]string, error)",
		"}",
	}
	p := acceptInterface(lines[0])
	if p == nil {
		fmt.Println("interface is bad")
		return
	}
	lines = lines[1:]
	for idx, line := range lines {
		ok, err := p.Next(line, nil)
		if err != nil {
			fmt.Printf("idx: %d, error: %v\n", idx, err)
			return
		}
		if !ok {
			break
		}
	}
	inter := p.Get().(*Interface)
	fmt.Printf("name = %s\n", inter.Name)
	for _, m := range inter.Methods {
		fmt.Println(m.Def)
		fmt.Printf("MethodName=%s, Imports=%v, RetSlice=%v, RetP=%v, RetSimple=%v, RetType=%s\n",
			m.Name, m.Imports, m.RetSlice, m.RetPointer,
			m.RetSimple, m.RetType)
	}

}
