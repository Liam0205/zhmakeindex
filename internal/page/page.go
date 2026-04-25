package page

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var ScanSyntaxError = errors.New("索引项语法错误")

const MaxInt = int(^uint(0) >> 1)

type RangeType int

const (
	PAGE_UNKNOWN RangeType = iota
	PAGE_OPEN
	PAGE_NORMAL
	PAGE_CLOSE
)

func (rt RangeType) String() string {
	switch rt {
	case PAGE_UNKNOWN:
		return "?"
	case PAGE_OPEN:
		return "("
	case PAGE_NORMAL:
		return "."
	case PAGE_CLOSE:
		return ")"
	default:
		panic("区间格式错误")
	}
}

type Page struct {
	Numbers    []PageNumber
	Compositor string
	Encap      string
	Rangetype  RangeType
}

func (p *Page) Empty() *Page {
	return &Page{
		Numbers:    nil,
		Compositor: p.Compositor,
		Encap:      p.Encap,
		Rangetype:  PAGE_UNKNOWN,
	}
}

func (p *Page) String() string {
	var page_str []string
	for _, pn := range p.Numbers {
		page_str = append(page_str, pn.String())
	}
	return strings.Join(page_str, p.Compositor)
}

func (p *Page) Compatible(other *Page) bool {
	if len(p.Numbers) != len(other.Numbers) {
		return false
	}
	for i := 0; i < len(p.Numbers); i++ {
		if p.Numbers[i].Format != other.Numbers[i].Format {
			return false
		}
	}
	return true
}

func (p *Page) Diff(other *Page) int {
	if !p.Compatible(other) {
		return -1
	}
	depth := len(p.Numbers)
	for i := 0; i < depth-1; i++ {
		if p.Numbers[i].Num != other.Numbers[i].Num {
			return MaxInt
		}
	}
	abs := func(x int) int {
		if x >= 0 {
			return x
		} else {
			return -x
		}
	}
	return abs(p.Numbers[depth-1].Num - other.Numbers[depth-1].Num)
}

func (p *Page) Cmp(other *Page, precedence map[NumFormat]int) int {
	for i := 0; i < len(p.Numbers) && i < len(other.Numbers); i++ {
		a, b := p.Numbers[i], other.Numbers[i]
		if precedence[a.Format] != precedence[b.Format] {
			return precedence[a.Format] - precedence[b.Format]
		} else if a.Num != b.Num {
			return a.Num - b.Num
		}
	}
	if len(p.Numbers) != len(other.Numbers) {
		return len(p.Numbers) - len(other.Numbers)
	}
	return 0
}

type PageNumber struct {
	Format NumFormat
	Num    int
}

func (p PageNumber) String() string {
	return p.Format.FormatNum(p.Num)
}

type NumFormat int

const (
	NUM_UNKNOWN NumFormat = iota
	NUM_ARABIC
	NUM_ROMAN_LOWER
	NUM_ROMAN_UPPER
	NUM_ALPH_LOWER
	NUM_ALPH_UPPER
)

func ScanPage(token []rune, compositor string) ([]PageNumber, error) {
	numstr_list := strings.Split(string(token), compositor)
	var nums []PageNumber
	for _, numstr := range numstr_list {
		pn, err := ScanNumber([]rune(numstr))
		if err != nil {
			return nil, err
		}
		nums = append(nums, pn)
	}
	return nums, nil
}

func ScanNumber(token []rune) (PageNumber, error) {
	if len(token) == 0 {
		return PageNumber{}, ScanSyntaxError
	}
	if r := token[0]; unicode.IsDigit(r) {
		num, err := scanArabic(token)
		return PageNumber{Format: NUM_ARABIC, Num: num}, err
	} else if romanLowerValue[r] != 0 {
		num, err := scanRomanLower(token)
		return PageNumber{Format: NUM_ROMAN_LOWER, Num: num}, err
	} else if romanUpperValue[r] != 0 {
		num, err := scanRomanUpper(token)
		return PageNumber{Format: NUM_ROMAN_UPPER, Num: num}, err
	} else if 'a' <= r && r <= 'z' {
		num, err := scanAlphLower(token)
		return PageNumber{Format: NUM_ALPH_LOWER, Num: num}, err
	} else if 'A' <= r && r <= 'Z' {
		num, err := scanAlphUpper(token)
		return PageNumber{Format: NUM_ALPH_UPPER, Num: num}, err
	}
	return PageNumber{}, ScanSyntaxError
}

func scanArabic(token []rune) (int, error) {
	num, err := strconv.Atoi(string(token))
	if err != nil {
		err = ScanSyntaxError
	}
	return num, err
}

func scanRomanLower(token []rune) (int, error) {
	return scanRoman(token, romanLowerValue)
}

func scanRomanUpper(token []rune) (int, error) {
	return scanRoman(token, romanUpperValue)
}

func scanRoman(token []rune, romantable map[rune]int) (int, error) {
	num := 0
	for i, r := range token {
		if romantable[r] == 0 {
			return 0, ScanSyntaxError
		}
		if i == 0 || romantable[r] <= romantable[token[i-1]] {
			num += romantable[r]
		} else {
			num += romantable[r] - 2*romantable[token[i-1]]
		}
	}
	return num, nil
}

var romanLowerValue = map[rune]int{
	'i': 1, 'v': 5, 'x': 10, 'l': 50, 'c': 100, 'd': 500, 'm': 1000,
}
var romanUpperValue = map[rune]int{
	'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000,
}

func scanAlphLower(token []rune) (int, error) {
	if len(token) != 1 || token[0] < 'a' || token[0] > 'z' {
		return 0, ScanSyntaxError
	}
	return int(token[0]-'a') + 1, nil
}

func scanAlphUpper(token []rune) (int, error) {
	if len(token) != 1 || token[0] < 'A' || token[0] > 'Z' {
		return 0, ScanSyntaxError
	}
	return int(token[0]-'A') + 1, nil
}

func (numfmt NumFormat) FormatNum(num int) string {
	switch numfmt {
	case NUM_UNKNOWN:
		return "?"
	case NUM_ARABIC:
		return fmt.Sprint(num)
	case NUM_ALPH_LOWER:
		return string(rune('a' + num))
	case NUM_ALPH_UPPER:
		return string(rune('A' + num))
	case NUM_ROMAN_LOWER:
		return RomanNumString(num, false)
	case NUM_ROMAN_UPPER:
		return RomanNumString(num, true)
	default:
		panic("数字格式错误")
	}
}

func RomanNumString(num int, upper bool) string {
	if num < 1 {
		return ""
	}
	type pair struct {
		symbol string
		value  int
	}
	var romanTable = []pair{
		{"m", 1000}, {"cm", 900}, {"d", 500}, {"cd", 400}, {"c", 100}, {"xc", 90},
		{"l", 50}, {"xl", 40}, {"x", 10}, {"ix", 9}, {"v", 5}, {"iv", 4}, {"i", 1},
	}
	var numstr []rune
	for _, p := range romanTable {
		for num >= p.value {
			numstr = append(numstr, []rune(p.symbol)...)
			num -= p.value
		}
	}
	if upper {
		return strings.ToUpper(string(numstr))
	} else {
		return string(numstr)
	}
}
