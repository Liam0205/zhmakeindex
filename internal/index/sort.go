package index

import "unicode/utf8"

type IndexEntrySlice struct {
	Entries  []IndexEntry
	Colattor IndexCollator
}

func (s IndexEntrySlice) Len() int {
	return len(s.Entries)
}

func (s IndexEntrySlice) Swap(i, j int) {
	s.Entries[i], s.Entries[j] = s.Entries[j], s.Entries[i]
}

func (s IndexEntrySlice) Strcmp(a, b string) int {
	atype, btype := GetStringType(s.Colattor, a), GetStringType(s.Colattor, b)
	if atype < btype {
		return -1
	} else if atype > btype {
		return 1
	}
	if cmp := DecimalStrcmp(a, b); cmp != 0 {
		return cmp
	}
	a_rune, b_rune := []rune(a), []rune(b)
	for i := range a_rune {
		if i >= len(b_rune) {
			return 1
		}
		cmp := s.Colattor.RuneCmp(a_rune[i], b_rune[i])
		if cmp != 0 {
			return cmp
		}
	}
	if len(a_rune) < len(b_rune) {
		return -1
	}
	if a < b {
		return -1
	} else if a > b {
		return 1
	} else {
		return 0
	}
}

func (s IndexEntrySlice) Less(i, j int) bool {
	a, b := s.Entries[i], s.Entries[j]
	for i := range a.Level {
		if i >= len(b.Level) {
			return false
		}
		keycmp := s.Strcmp(a.Level[i].Key, b.Level[i].Key)
		if keycmp < 0 {
			return true
		} else if keycmp > 0 {
			return false
		}
		textcmp := s.Strcmp(a.Level[i].Text, b.Level[i].Text)
		if textcmp < 0 {
			return true
		} else if textcmp > 0 {
			return false
		}
	}
	if len(a.Level) < len(b.Level) {
		return true
	}
	return false
}

type StringType int

const (
	EMPTY_STR      StringType = iota
	SYMBOL_STR
	NUM_SYMBOL_STR
	NUM_STR
	LETTER_STR
)

func GetStringType(collator IndexCollator, s string) StringType {
	if len(s) == 0 {
		return EMPTY_STR
	}
	r, _ := utf8.DecodeRuneInString(s)
	switch {
	case IsNumRune(r):
		if IsNumString(s) {
			return NUM_STR
		} else {
			return NUM_SYMBOL_STR
		}
	case collator.IsLetter(r):
		return LETTER_STR
	default:
		return SYMBOL_STR
	}
}
