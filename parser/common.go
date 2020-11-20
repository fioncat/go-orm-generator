package parser

import "strings"

func SP(s, sep string) []string {
	return strings.Split(s, sep)
}

func FD(s string) []string {
	return strings.Fields(s)
}

func T(s, cut string) string {
	return strings.Trim(s, cut)
}

func TS(s string) string {
	return strings.TrimSpace(s)
}

func TL(s, cut string) string {
	return strings.TrimLeft(s, cut)
}

func TR(s, cut string) string {
	return strings.TrimRight(s, cut)
}

func HasL(s, p string) bool {
	return strings.HasPrefix(s, p)
}

func HasR(s, p string) bool {
	return strings.HasSuffix(s, p)
}
