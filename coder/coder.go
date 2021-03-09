package coder

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
)

type Coder struct {
	lines []string

	subs []SubCoder
}

func (c *Coder) P(n int, vs ...interface{}) {
	s := joins(vs)
	if n > 0 {
		prefix := strings.Repeat("\t", n)
		s = prefix + s
	}
	c.lines = append(c.lines, s)
	return
}

func (c *Coder) Empty() {
	c.lines = append(c.lines, "")
}

func (c *Coder) AddSub(sc SubCoder) {
	c.subs = append(c.subs, sc)
}

func (c *Coder) Body() {
	var hasContent bool
	for _, sub := range c.subs {
		if hasContent {
			c.Empty()
		}
		hasContent = sub.code(c)
	}
}

func (c *Coder) WriteFile(path string) error {
	lines := make([]string, len(c.lines))
	for idx, line := range c.lines {
		lines[idx] = strings.TrimRightFunc(line, unicode.IsSpace)
	}
	data := []byte(strings.Join(lines, "\n"))
	return ioutil.WriteFile(path, data, 0644)
}

type SubCoder interface {
	code(c *Coder) bool
}

// Quote the string in double quotes
func Quote(ss ...string) string {
	s := strings.Join(ss, "")
	return fmt.Sprintf(`"%s"`, s)
}

// GoName converts an underscore name into a camel case
// name with an initial capital letter.
func GoName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.Split(name, "_")
	for i := range parts {
		parts[i] = Export(parts[i])
	}

	return strings.Join(parts, "")
}

// Export capitalizes the first letter of the name.
func Export(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return string(unicode.ToLower(rune(s[0])))
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

func joins(vs []interface{}) string {
	strs := make([]string, len(vs))
	for idx, v := range vs {
		strs[idx] = fmt.Sprint(v)
	}
	return strings.Join(strs, "")
}
