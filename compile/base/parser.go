package base

import (
	"strings"

	"github.com/fioncat/go-gendb/misc/errors"
)

type LineParser interface {
	Accept(line string) bool
	Do(line string) (interface{}, error)
}

type ScanParser interface {
	Next(idx int, line string, tags []*Tag) (bool, error)
	Get() interface{}
}

func ScanLines(prefix string, p ScanParser, lines []string, idx *int) error {
	*idx += 1
	start := *idx
	found := false
	var tags []*Tag
	for *idx < len(lines) {
		line := lines[*idx]
		line = strings.TrimSpace(line)
		if line == "" {
			*idx += 1
			continue
		}
		tag, err := ParseTag(prefix, line)
		if err != nil {
			return err
		}
		if tag != nil {
			if tag.Name == "end" {
				found = true
				break
			}
			tags = append(tags, tag)
			*idx += 1
			continue
		}
		if strings.HasPrefix(line, prefix) {
			*idx += 1
			continue
		}
		ok, err := p.Next(*idx, line, tags)
		if err != nil {
			return err
		}
		if !ok {
			found = true
			break
		}
		tags = nil
		*idx += 1
	}
	if !found {
		return errors.TraceFmt(start,
			"can not find close tag")
	}
	*idx += 1

	return nil
}
