package rand

import (
	"math/rand"
	"sync"
	"time"
)

var (
	LCASE_DICT = []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
		"u", "v", "w", "x", "y", "z",
	}

	UCASE_DICT = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
		"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T",
		"U", "V", "W", "X", "Y", "Z",
	}

	CASE = append(LCASE_DICT, UCASE_DICT...)

	NUM = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}

	seedOnce sync.Once
)

func Range(min, max int) int {
	seedOnce.Do(func() {
		rand.Seed(time.Now().Unix())
	})
	if min >= max {
		return max
	}
	return rand.Intn(max-min) + min
}

func Choose(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	idx := Range(0, len(slice))
	return slice[idx]
}
