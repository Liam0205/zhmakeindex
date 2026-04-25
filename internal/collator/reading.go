package collator

import (
	"unicode"
	"unicode/utf8"

	"github.com/leo-liu/zhmakeindex/CJK"
	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

type ReadingIndexCollator struct{}

func (_ ReadingIndexCollator) InitGroups(style *style.OutputStyle) []index.IndexGroup {
	groups := make([]index.IndexGroup, 2+26)
	if style.HeadingsFlag > 0 {
		groups[0].Name = style.SymheadPositive
		groups[1].Name = style.NumheadPositive
		for alph, i := 'A', 2; alph <= 'Z'; alph++ {
			groups[i].Name = string(alph)
			i++
		}
	} else if style.HeadingsFlag < 0 {
		groups[0].Name = style.SymheadNegative
		groups[1].Name = style.NumheadNegative
		for alph, i := 'a', 2; alph <= 'z'; alph++ {
			groups[i].Name = string(alph)
			i++
		}
	}
	return groups
}

func (_ ReadingIndexCollator) Group(entry *index.IndexEntry) int {
	if len(entry.Level) == 0 {
		return 0
	}
	first, _ := utf8.DecodeRuneInString(entry.Level[0].Key)
	first = unicode.ToLower(first)
	switch {
	case index.IsNumString(entry.Level[0].Key):
		return 1
	case 'a' <= first && first <= 'z':
		return 2 + int(first) - 'a'
	case CJK.Readings[first] != "":
		reading_first := int(CJK.Readings[first][0])
		return 2 + reading_first - 'a'
	default:
		return 0
	}
}

func (_ ReadingIndexCollator) RuneCmp(a, b rune) int {
	a_reading, b_reading := CJK.Readings[a], CJK.Readings[b]
	switch {
	case a_reading == "" && b_reading == "":
		return index.RuneCmpIgnoreCases(a, b)
	case a_reading == "" && b_reading != "":
		return -1
	case a_reading != "" && b_reading == "":
		return 1
	case a_reading < b_reading:
		return -1
	case a_reading > b_reading:
		return 1
	default:
		return int(a - b)
	}
}

func (_ ReadingIndexCollator) IsLetter(r rune) bool {
	r = unicode.ToLower(r)
	switch {
	case 'a' <= r && r <= 'z':
		return true
	case CJK.Readings[r] != "":
		return true
	default:
		return false
	}
}
