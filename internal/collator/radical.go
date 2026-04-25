package collator

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/leo-liu/zhmakeindex/CJK"
	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

type RadicalIndexCollator struct{}

func (_ RadicalIndexCollator) InitGroups(style *style.OutputStyle) []index.IndexGroup {
	groups := make([]index.IndexGroup, 2+26+CJK.MAX_RADICAL)
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
	for r, i := 1, 2+26; r < CJK.MAX_RADICAL+1; r++ {
		var radicalName string
		if CJK.Radicals[r].Simplified != 0 && style.RadicalSimplifiedFlag != 0 {
			radicalName = fmt.Sprintf("%c%s%c%s",
				CJK.Radicals[r].Origin, style.RadicalSimplifiedPrefix, CJK.Radicals[r].Simplified, style.RadicalSimplifiedSuffix)
		} else {
			radicalName = string(CJK.Radicals[r].Origin)
		}
		groups[i].Name = style.RadicalPrefix + radicalName + style.RadicalSuffix
		i++
	}
	return groups
}

func (_ RadicalIndexCollator) Group(entry *index.IndexEntry) int {
	first, _ := utf8.DecodeRuneInString(entry.Level[0].Key)
	first = unicode.ToLower(first)
	switch {
	case index.IsNumString(entry.Level[0].Key):
		return 1
	case 'a' <= first && first <= 'z':
		return 2 + int(first) - 'a'
	case CJK.RadicalStrokes[first] != "":
		return 2 + 26 + (CJK.RadicalStrokes[first].Radical() - 1)
	default:
		return 0
	}
}

func (_ RadicalIndexCollator) RuneCmp(a, b rune) int {
	a_rs, b_rs := CJK.RadicalStrokes[a], CJK.RadicalStrokes[b]
	switch {
	case a_rs == "" && b_rs == "":
		return index.RuneCmpIgnoreCases(a, b)
	case a_rs == "" && b_rs != "":
		return -1
	case a_rs != "" && b_rs == "":
		return 1
	case a_rs < b_rs:
		return -1
	case a_rs > b_rs:
		return 1
	default:
		return int(a - b)
	}
}

func (_ RadicalIndexCollator) IsLetter(r rune) bool {
	r = unicode.ToLower(r)
	switch {
	case 'a' <= r && r <= 'z':
		return true
	case CJK.RadicalStrokes[r] != "":
		return true
	default:
		return false
	}
}
