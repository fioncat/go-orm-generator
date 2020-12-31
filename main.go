package main

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/iter"
)

func main() {
	ss := []string{
		"a", "b", "c",
		"asd", "nihao",
		"我是谁?",
	}

	iter := iter.New(ss)
	var s string
	for iter.Next(&s) {
		fmt.Println(s)
	}
}
