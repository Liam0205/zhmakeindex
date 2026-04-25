package collator

import (
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/leo-liu/zhmakeindex/CJK"
	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

type StrokeIndexCollator struct{}

func (_ StrokeIndexCollator) InitGroups(style *style.OutputStyle) []index.IndexGroup {
	groups := make([]index.IndexGroup, 2+26+CJK.MAX_STROKE)
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
	for stroke, i := 1, 2+26; stroke <= CJK.MAX_STROKE; stroke++ {
		groups[i].Name = style.StrokePrefix + strconv.Itoa(stroke) + style.StrokeSuffix
		i++
	}
	return groups
}

func (_ StrokeIndexCollator) Group(entry *index.IndexEntry) int {
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
	case len(CJK.Strokes[first]) > 0:
		return 2 + 26 + (len(CJK.Strokes[first]) - 1)
	default:
		return 0
	}
}

func (_ StrokeIndexCollator) RuneCmp(a, b rune) int {
	a_strokes, b_strokes := len(CJK.Strokes[a]), len(CJK.Strokes[b])
	switch {
	case a_strokes == 0 && b_strokes == 0:
		return index.RuneCmpIgnoreCases(a, b)
	case a_strokes == 0 && b_strokes != 0:
		return -1
	case a_strokes != 0 && b_strokes == 0:
		return 1
	case a_strokes != b_strokes:
		return a_strokes - b_strokes
	case CJK.Strokes[a] < CJK.Strokes[b]:
		return -1
	case CJK.Strokes[a] > CJK.Strokes[b]:
		return 1
	default:
		return int(a - b)
	}
}

func (_ StrokeIndexCollator) IsLetter(r rune) bool {
	r = unicode.ToLower(r)
	switch {
	case 'a' <= r && r <= 'z':
		return true
	case CJK.Strokes[r] != "":
		return true
	default:
		return false
	}
}
