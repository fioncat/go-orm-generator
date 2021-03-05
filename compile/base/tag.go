package base

import (
	"fmt"
	"strings"
)

type Tag struct {
	Line int

	Name    string
	Options []Option
}

type Option struct {
	Key   string
	Value string
}

// +gen:xxx
const tagPrefix = "+gen:"

func ParseTag(idx int, prefix, line string) (*Tag, error) {
	if !strings.HasPrefix(line, prefix) {
		return nil, nil
	}
	line = strings.TrimPrefix(line, prefix)
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, tagPrefix) {
		return nil, nil
	}
	line = strings.TrimPrefix(line, tagPrefix)
	tmp := strings.Split(line, " ")
	if len(tmp) == 0 {
		return nil, fmt.Errorf("tag body is empty")
	}
	tag := new(Tag)
	tag.Name = tmp[0]
	tag.Line = idx + 1
	tmp = tmp[1:]
	for _, optStr := range tmp {
		optTmp := strings.Split(optStr, "=")
		var opt Option
		switch len(optTmp) {
		case 1:
			opt.Value = optTmp[0]

		case 2:
			opt.Key = optTmp[0]
			opt.Value = optTmp[1]

		default:
			return nil, fmt.Errorf(`option "%s" is `+
				`bad format`, optStr)
		}
		tag.Options = append(tag.Options, opt)
	}
	return tag, nil
}
