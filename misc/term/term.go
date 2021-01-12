package term

import "fmt"

var enable bool

var colorsMap = map[string]int32{
	"black":   30,
	"red":     31,
	"green":   32,
	"yellow":  33,
	"blue":    34,
	"fuchsia": 35,
	"cyan":    36,
	"white":   37,
}

// EnableColor enable terminal color output.
// If you don't call this method, calling all color
// methods will return the original string.
// The colors are only available for unix or linux
// terminals.
func EnableColor() {
	enable = true
}

func colorString(s string, cno int32) string {
	if !enable {
		return s
	}
	return fmt.Sprintf("\033[%d;1m%s\033[0m", cno, s)
}

// Warn returns string with 'yellow' color.
func Warn(s string) string {
	return colorString(s, colorsMap["yellow"])
}

// Red returns string with 'red' color.
func Red(s string) string {
	return colorString(s, colorsMap["red"])
}

// Cyan returns string with 'cyan' color.
func Cyan(s string) string {
	return colorString(s, colorsMap["cyan"])
}

// Info returns string with 'green' color.
func Info(s string) string {
	return colorString(s, colorsMap["green"])
}
