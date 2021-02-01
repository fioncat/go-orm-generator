package stoml

import (
	"fmt"
	"strings"
)

// Result represents the result of scanning a toml file.
// It includes multiple global options and multiple sections.
// The toml scanned here only supports the most basic format:
//
//  [section]
//    option1 = value1
//    option2 = value2
//    ...
//
type Result struct {
	// Options are external and are global options
	// without section.
	Options []*Option `json:"options"`

	// Sections represent all resolved sections. A
	// section contains a name and multiple options.
	Sections []*Section `json:"sections"`
}

// Section represents the named section of multiple
// option combinations in the toml file. There can be
// multiple such sections in the entire toml file.
//
// The section starts with the definition of "[section-name]"
// and ends with the next section definition. All options in
// the middle belong to this section.
type Section struct {
	// Name of the section
	Name string `json:"name"`

	// Options stores all the options contained in
	// the section.
	Options []*Option `json:"options"`

	// Line is the number of lines in the section.
	Line int `json:"line"`
}

// Option is a specific configuration item in the toml
// configuration file, and its format is very simple:
//   "key = value".
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

// Do scans the toml configuration file and returns the
// data in the form of Result.
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
