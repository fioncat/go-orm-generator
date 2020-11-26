package scanner

import "strings"

type lineScanner struct {
	idx   int
	lines []string
}

func newLines(content string) *lineScanner {
	lines := strings.Split(content, "\n")
	return &lineScanner{lines: lines}
}

func (s *lineScanner) next() (string, int) {
	if s.idx >= len(s.lines) {
		return "", -1
	}
	line := s.lines[s.idx]
	num := s.idx + 1
	s.idx += 1
	return strings.TrimSpace(line), num
}

func (s *lineScanner) pick() string {
	if s.idx+1 >= len(s.lines) {
		return ""
	}
	return s.lines[s.idx+1]
}

type fieldScanner struct {
	idx    int
	fields []string
}

func newFields(content string) *fieldScanner {
	fields := strings.Fields(content)
	return &fieldScanner{fields: fields}
}

func (s *fieldScanner) next() (string, bool) {
	if s.idx >= len(s.fields) {
		return "", false
	}
	field := s.fields[s.idx]
	s.idx += 1
	return field, true
}

type charScanner struct {
	ch    rune
	idx   int
	runes []rune
}

func newChars(content string) *charScanner {
	return &charScanner{runes: []rune(content)}
}

func (s *charScanner) next() (rune, bool) {
	if s.idx >= len(s.runes) {
		return 0, false
	}
	rune := s.runes[s.idx]
	s.idx += 1
	return rune, true
}
