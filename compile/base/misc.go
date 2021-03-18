package base

import (
	"fmt"
	"strings"
)

func ReplacePlaceholder(s string, reps ...string) string {
	for idx, rep := range reps {
		ph := fmt.Sprintf("$%d", idx)
		s = strings.ReplaceAll(s, ph, rep)
	}
	return s
}
