package parser

import (
	"fmt"
	"testing"
)

func TestGoMethod(t *testing.T) {
	goodDefs := []string{
		"add() (int64, error)",
		"insert(u *User) (int64, error)",
		"queryAll() ([]*User, error)",
		"queryOne(id int64) (*User, error)",
		"Hello(a int, s string, d Detail) ([]int64, error)",
		"FindByID(db *sql.DB, id int64) (*User, error)",
		"FindAll(db runner.IDB) (*User, error)",
		"Find(db *sql.DB, id int64, name string) (*model.User, error)",
	}

	fmt.Println(">> GOOD:")
	for _, def := range goodDefs {
		fmt.Printf(">>>>>>>>>>>>>>: %s\n", def)
		m, err := MustGoMethod(def)
		if err != nil {
			fmt.Printf("parse failed: %v\n", err)
			return
		}
		fmt.Println(m)
		fmt.Printf("pkgs: %v\n", m.Pkgs.Slice())
		fmt.Printf("names: %v\n", m.ParamNames.Slice())
	}
}
