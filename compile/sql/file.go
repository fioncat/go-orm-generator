package sql

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/misc/errors"
)

const commPrefix = "--"

type File struct {
	Path string

	Methods []*Method
}

type acceptAction func(tag *base.Tag) (base.ScanParser, error)

var acceptActions = []acceptAction{
	acceptSql,
}

func ReadFile(path string) (*File, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadLines(path, strings.Split(string(data), "\n"))
}

func ReadLines(path string, lines []string) (*File, error) {
	file, err := readLines(path, lines)
	if err != nil {
		err = errors.Trace(path, err)
		return nil, errors.OnCompile(path, lines, err)
	}
	return file, nil
}

func readLines(path string, lines []string) (*File, error) {
	file := new(File)
	file.Path = path

	acceptNext := func(idx int, tag *base.Tag) (bool, error) {
		if tag == nil {
			return true, nil
		}
		return false, nil
	}

	idx, err := base.Accept(lines, commPrefix, acceptNext)
	if err != nil {
		return nil, err
	}
	idx += 1

	for idx < len(lines) {
		line := lines[idx]
		tag, err := base.ParseTag(commPrefix, line)
		if err != nil {
			return nil, err
		}
		if tag == nil {
			idx++
			continue
		}
		var p base.ScanParser
		for _, accept := range acceptActions {
			p, err = accept(tag)
			if err != nil {
				return nil, err
			}
		}
		if p == nil {
			err = fmt.Errorf(`unknown tag type "%s"`, tag.Name)
			return nil, errors.Trace(idx+1, err)
		}

		err = base.ScanLines(commPrefix, p, lines, &idx)
		if err != nil {
			return nil, err
		}

		v := p.Get()
		switch v.(type) {
		case error:
			return nil, v.(error)

		case *Method:
			file.Methods = append(file.Methods, v.(*Method))
		}
	}

	for _, m := range file.Methods {
		if m.State != nil {
			flatState(m.State)
		}
		for _, dp := range m.Dps {
			flatState(dp.State)
		}
	}

	return file, nil
}

func flatState(state *Statement) {
	for _, ph := range state.phs {
		if ph.pre {
			state.Prepares = append(state.Prepares, ph.name)
			continue
		}
		state.Replaces = append(state.Replaces, ph.name)
	}
}
