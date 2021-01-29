package stoml

import (
	"fmt"
	"strings"
)

type Result struct {
	Options  []*Option  `json:"options"`
	Sections []*Section `json:"sections"`
}

type Section struct {
	Name    string    `json:"name"`
	Options []*Option `json:"options"`
	Line    int       `json:"line"`
}

type Option struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Line  int    `json:"line"`
}

const (
	sectionPrefix = '['
	sectionSuffix = ']'

	optionSep = "="
)

func Do(path, content string) (*Result, error) {
	lines := strings.Split(content, "\n")

	r := new(Result)

	for idx, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lineNum := idx + 1
		if line[0] == sectionPrefix {
			sec := new(Section)
			err := section(line, sec)
			if err != nil {
				return nil, fmt.Errorf(
					"%s:%d: %v", path, lineNum, err)
			}
			sec.Line = lineNum
			r.Sections = append(r.Sections, sec)
			continue
		}
		opt := new(Option)
		err := option(line, opt)
		if err != nil {
			return nil, fmt.Errorf(
				"%s:%d: %v", path, lineNum, err)
		}
		opt.Line = lineNum
		if len(r.Sections) == 0 {
			r.Options = append(r.Options, opt)
			continue
		}
		lidx := len(r.Sections) - 1
		r.Sections[lidx].Options = append(
			r.Sections[lidx].Options, opt)
	}

	return r, nil
}

func section(line string, sec *Section) error {
	if len(line) <= 2 {
		return fmt.Errorf("section name bad format")
	}
	if line[0] != sectionPrefix ||
		line[len(line)-1] != sectionSuffix {
		return fmt.Errorf(`section should starts with `+
			` '[', ends with ']', found '%c' and '%c'`,
			line[0], line[len(line)-1])
	}
	line = line[1 : len(line)-1]
	if line == "" {
		return fmt.Errorf("section name is empty")
	}
	sec.Name = line
	return nil
}

func option(line string, opt *Option) error {
	tmp := strings.Split(line, optionSep)
	if len(tmp) != 2 {
		return fmt.Errorf("option bad format")
	}
	key := strings.TrimSpace(tmp[0])
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	val := strings.TrimSpace(tmp[1])
	if val == "" {
		return fmt.Errorf("value is empty")
	}
	opt.Key = key
	opt.Value = val
	return nil
}
