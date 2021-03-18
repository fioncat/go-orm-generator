package golang

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/token"
)

type DeepType struct {
	Slice *DeepType `json:"slice,omitempty"`
	Map   *DeepMap  `json:"map,omitempty"`

	Pointer bool `json:"pointer"`
	Simple  bool `json:"simple"`

	Package string `json:"package,omitempty"`

	Full string `json:"full,omitempty"`
	Name string `json:"name,omitempty"`
}

type DeepMap struct {
	Full  string    `json:"full"`
	Key   string    `json:"key"`
	Value *DeepType `json:"value"`
}

func ParseDeepType(ts string) (*DeepType, error) {
	t := new(DeepType)
	err := parseDeepType(ts, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func parseDeepType(ts string, t *DeepType) error {
	t.Full = ts
	s := token.NewScanner(ts, _typeTokens)
	var e token.Element
	ok := s.Next(&e)
	if !ok {
		return typeBadFormat("type", ts)
	}
	switch e.Token {
	case token.LBRACK:
		ok = s.Next(&e)
		if !ok || e.Token != token.RBRACK {
			return typeBadFormat("slice", ts)
		}
		t.Slice = new(DeepType)
		subStr := strings.TrimPrefix(ts, "[]")
		return parseDeepType(subStr, t.Slice)

	case _map:
		t.Map = new(DeepMap)
		return parseMapType(ts, s, t.Map)

	case token.MUL:
		t.Pointer = true
		ok = s.Next(&e)
		if !ok {
			return typeBadFormat("type", ts)
		}
	}
	if !e.Indent {
		return typeBadFormat("type", ts)
	}

	first := e.Get()
	ok = s.Next(&e)
	if ok {
		t.Package = first
		if e.Token != token.PERIOD {
			return typeBadFormat("type", ts)
		}
		ok = s.Next(&e)
		if !ok || !e.Indent {
			return typeBadFormat("type", ts)
		}
		t.Name = e.Get()
	} else {
		if !e.Indent {
			return typeBadFormat("type", ts)
		}
		t.Name = e.Get()
	}

	t.Simple = IsSimpleType(t.Name)
	return nil
}

func parseMapType(ts string, s *token.Scanner, dm *DeepMap) error {
	dm.Full = ts
	var e token.Element
	ok := s.Next(&e)
	if !ok || e.Token != token.LBRACK {
		return typeBadFormat("map", ts)
	}
	ok = s.Next(&e)
	if !ok || !e.Indent {
		return typeBadFormat("map", ts)
	}
	dm.Key = e.Get()
	ok = s.Next(&e)
	if !ok || e.Token != token.RBRACK {
		return typeBadFormat("map", ts)
	}
	prefix := fmt.Sprintf("map[%s]", dm.Key)
	subStr := strings.TrimPrefix(ts, prefix)
	dm.Value = new(DeepType)
	return parseDeepType(subStr, dm.Value)
}

func typeBadFormat(t, s string) error {
	return fmt.Errorf(`%s: "%s" is bad format`, t, s)
}
