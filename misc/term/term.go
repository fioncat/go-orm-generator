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

func EnableColor() {
	enable = true
}

func colorString(s string, cno int32) string {
	if !enable {
		return s
	}
	return fmt.Sprintf("\033[%d;1m%s\033[0m", cno, s)
}

func Warn(s string) string {
	return colorString(s, colorsMap["yellow"])
}

func Red(s string) string {
	return colorString(s, colorsMap["red"])
}

func Cyan(s string) string {
	return colorString(s, colorsMap["cyan"])
}

func Info(s string) string {
	return colorString(s, colorsMap["green"])
}
