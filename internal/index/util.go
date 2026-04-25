package index

import (
	"strconv"
	"unicode"
)

func RuneCmpIgnoreCases(a, b rune) int {
	la, lb := unicode.ToLower(a), unicode.ToLower(b)
	return int(la - lb)
}

func IsNumRune(r rune) bool {
	return unicode.IsNumber(r) && r != '〇'
}

func IsNumString(s string) bool {
	for _, r := range s {
		if !IsNumRune(r) {
			return false
		}
	}
	return true
}

func DecimalStrcmp(a, b string) int {
	aint, err := strconv.ParseUint(a, 10, 64)
	if err != nil {
		return 0
	}
	bint, err := strconv.ParseUint(b, 10, 64)
	if err != nil {
		return 0
	}
	switch {
	case aint < bint:
		return -1
	case aint > bint:
		return 1
	default:
		return 0
	}
}
