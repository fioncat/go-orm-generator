package base

import (
	"strings"

	"github.com/fioncat/go-gendb/misc/errors"
)

// LineParser is used to parse the single line content in
// the file.
type LineParser interface {
	// Accept returns whether the line should be parsed by
	// the current parser.
	Accept(line string) bool

	// Do execute the parsing process.
	Do(line string) (interface{}, error)
}

// ScanParser is used to parse multiple lines of content
// in a file.
type ScanParser interface {
	// Next receives the content of the next line. If the
	// first bool returned is false, it means that the
	// multi-line parsing is over.
	Next(idx int, line string, tags []*Tag) (bool, error)

	// Get returns the parsing result.
	Get() interface{}
}

// ScanLines uses ScanPaser to scan and parse multiple lines
// of content. In the parsing process, the index will be
// continuously incremented, and finally the index will
// point to the next line of the multi-line content.
func ScanLines(prefix string, p ScanParser, lines []string, idx *int) error {
	// The first line is the definition of multi-line content,
	// it is used to create ScanParser, ScanParser can not
	// receive it, so it should be skipped. (The creation
	// function is of external concern)
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
		tag, err := ParseTag(*idx, prefix, line)
		if err != nil {
			return err
		}
		if tag != nil {
			// built-in end tag.
			// When it is found, regardless of the ScanPaser
			// return value, it is directly considered that
			// the multi-line content is over.
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
