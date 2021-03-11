package golang

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
)

const commentPrefix = "//"

type acceptAction func(int, string, []*base.Tag, []string) (base.ScanParser, error)

var (
	packageParser base.LineParser = new(_packageParser)
	importParser  base.LineParser = new(_singleImportParser)

	acceptActions = []acceptAction{
		acceptInterface,
		acceptStruct,
	}
)

type File struct {
	Lines []string

	Type string
	Path string

	Package string

	Imports []*Import

	Interfaces []*Interface

	Structs []*Struct

	Options []base.Option
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
	start := time.Now()
	file := new(File)
	file.Path = path
	file.Lines = lines

	acceptNext := func(idx int, tag *base.Tag) (bool, error) {
		if tag != nil {
			file.Type = tag.Name
			for _, opt := range tag.Options {
				file.Options = append(file.Options, opt)
			}
			return true, nil
		}
		line := lines[idx]
		if packageParser.Accept(line) {
			name, err := packageParser.Do(line)
			if err != nil {
				return false, errors.Trace(idx+1, err)
			}
			file.Package = name.(string)
			return false, nil
		}
		return true, nil
	}

	idx, err := base.Accept(lines, commentPrefix, acceptNext)
	if err != nil {
		return nil, err
	}
	if file.Package == "" {
		return nil, fmt.Errorf("missing package")
	}

	var tags []*base.Tag
	var comms []string
	for idx < len(lines) {
		line := lines[idx]
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			idx++
			continue
		}
		if strings.HasPrefix(line, commentPrefix) {
			tag, err := base.ParseTag(idx, commentPrefix, line)
			if err != nil {
				return nil, errors.Trace(idx+1, err)
			}
			if tag != nil {
				tags = append(tags, tag)
			} else {
				comm := strings.TrimPrefix(line, commentPrefix)
				comm = strings.TrimSpace(comm)
				comms = append(comms, comm)
			}
			idx++
			continue
		}
		if importParser.Accept(line) {
			v, err := importParser.Do(line)
			if err != nil {
				return nil, errors.Trace(idx+1, err)
			}
			imp := v.(*Import)
			file.Imports = append(file.Imports, imp)
			idx++
			continue
		}
		if p := acceptPaths(line); p != nil {
			err := base.ScanLines(commentPrefix, p, lines, &idx)
			if err != nil {
				return nil, errors.Trace(idx+1, err)
			}
			imps := p.Get().([]*Import)
			file.Imports = append(file.Imports, imps...)
			continue
		}

		var p base.ScanParser
		var err error
		for _, accept := range acceptActions {
			p, err = accept(idx, line, tags, comms)
			if err != nil {
				return nil, errors.Trace(idx+1, err)
			}
			if p != nil {
				comms = nil
				break
			}
		}
		if p == nil {
			idx++
			continue
		}
		err = base.ScanLines(commentPrefix, p, lines, &idx)
		if err != nil {
			return nil, errors.Trace(idx+1, err)
		}
		v := p.Get()
		switch v.(type) {
		case *Struct:
			file.Structs = append(file.Structs,
				v.(*Struct))

		case *Interface:
			file.Interfaces = append(file.Interfaces,
				v.(*Interface))
		}
		tags = nil
	}
	log.Infof("[compile] [golang] [%v] %s, %d inter(s)",
		time.Since(start), path, len(file.Interfaces))

	return file, nil
}
