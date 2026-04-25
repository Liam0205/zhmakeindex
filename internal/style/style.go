package style

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/transform"

	"github.com/leo-liu/zhmakeindex/kpathsea"
)

type InputStyle struct {
	Keyword        string
	ArgOpen        rune
	ArgClose       rune
	Actual         rune
	Encap          rune
	Escape         rune
	Level          rune
	Quote          rune
	PageCompositor string
	RangeOpen      rune
	RangeClose     rune
	Comment        rune
}

func NewInputStyle() *InputStyle {
	return &InputStyle{
		Keyword:        "\\indexentry",
		ArgOpen:        '{',
		ArgClose:       '}',
		Actual:         '@',
		Encap:          '|',
		Escape:         '\\',
		Level:          '!',
		Quote:          '"',
		PageCompositor: "-",
		RangeOpen:      '(',
		RangeClose:     ')',
		Comment:        '%',
	}
}

type OutputStyle struct {
	Preamble                string
	Postamble               string
	SetpagePrefix           string
	SetpageSuffix           string
	GroupSkip               string
	HeadingsFlag            int
	HeadingPrefix           string
	HeadingSuffix           string
	SymheadPositive         string
	SymheadNegative         string
	NumheadPositive         string
	NumheadNegative         string
	StrokePrefix            string
	StrokeSuffix            string
	RadicalPrefix           string
	RadicalSuffix           string
	RadicalSimplifiedFlag   int
	RadicalSimplifiedPrefix string
	RadicalSimplifiedSuffix string
	Item0                   string
	Item1                   string
	Item2                   string
	Item01                  string
	ItemX1                  string
	Item12                  string
	ItemX2                  string
	Delim0                  string
	Delim1                  string
	Delim2                  string
	DelimN                  string
	DelimR                  string
	DelimT                  string
	EncapPrefix             string
	EncapInfix              string
	EncapSuffix             string
	PagePrecedence          string
	LineMax                 int
	IndentSpace             string
	IndentLength            int
	Suffix2p                string
	Suffix3p                string
	SuffixMp                string
}

func NewOutputStyle() *OutputStyle {
	return &OutputStyle{
		Preamble:                "\\begin{theindex}\n",
		Postamble:               "\n\n\\end{theindex}\n",
		SetpagePrefix:           "\n  \\setcounter{page}{",
		SetpageSuffix:           "}\n",
		GroupSkip:               "\n\n  \\indexspace\n",
		HeadingsFlag:            0,
		HeadingPrefix:           "",
		HeadingSuffix:           "",
		SymheadPositive:         "Symbols",
		SymheadNegative:         "symbols",
		NumheadPositive:         "Numbers",
		NumheadNegative:         "numbers",
		StrokePrefix:            "",
		StrokeSuffix:            " 画",
		RadicalPrefix:           "",
		RadicalSuffix:           "部",
		RadicalSimplifiedFlag:   1,
		RadicalSimplifiedPrefix: "（",
		RadicalSimplifiedSuffix: "）",
		Item0:           "\n  \\item ",
		Item1:           "\n    \\subitem ",
		Item2:           "\n      \\subsubitem ",
		Item01:          "\n    \\subitem ",
		ItemX1:          "\n    \\subitem ",
		Item12:          "\n      \\subsubitem ",
		ItemX2:          "\n      \\subsubitem ",
		Delim0:          ", ",
		Delim1:          ", ",
		Delim2:          ", ",
		DelimN:          ", ",
		DelimR:          "--",
		DelimT:          "",
		EncapPrefix:     "\\",
		EncapInfix:      "{",
		EncapSuffix:     "}",
		PagePrecedence:  "rnaRA",
		LineMax:         72,
		IndentSpace:     "\t\t",
		IndentLength:    16,
		Suffix2p:       "",
		Suffix3p:       "",
		SuffixMp:       "",
	}
}

func LoadStyles(stylePath string, decoder transform.Transformer) (*InputStyle, *OutputStyle) {
	in := NewInputStyle()
	out := NewOutputStyle()

	if stylePath == "" {
		return in, out
	}
	if filepath.Ext(stylePath) == "" {
		stylePath += ".ist"
	}
	stylePath = kpathsea.FindFile(stylePath)
	if stylePath == "" {
		log.Fatalln("找不到格式文件。")
	}
	styleFile, err := os.Open(stylePath)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer styleFile.Close()
	scanner := bufio.NewScanner(transform.NewReader(styleFile, decoder))
	scanner.Split(ScanStyleTokens)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Println(err.Error())
		}
		key := scanner.Text()
		if !scanner.Scan() {
			log.Println("格式文件不完整")
		}
		if err := scanner.Err(); err != nil {
			log.Println(err.Error())
		}
		value := scanner.Text()
		switch key {
		case "keyword":
			in.Keyword = Unquote(value)
		case "arg_open":
			in.ArgOpen = UnquoteChar(value)
		case "arg_close":
			in.ArgClose = UnquoteChar(value)
		case "actual":
			in.Actual = UnquoteChar(value)
		case "encap":
			in.Encap = UnquoteChar(value)
		case "escape":
			in.Escape = UnquoteChar(value)
		case "level":
			in.Level = UnquoteChar(value)
		case "quote":
			in.Quote = UnquoteChar(value)
		case "page_compositor":
			in.PageCompositor = Unquote(value)
		case "range_open":
			in.RangeOpen = UnquoteChar(value)
		case "range_close":
			in.RangeClose = UnquoteChar(value)
		case "comment":
			in.Comment = UnquoteChar(value)
		case "preamble":
			out.Preamble = Unquote(value)
		case "postamble":
			out.Postamble = Unquote(value)
		case "setpage_prefix":
			out.SetpagePrefix = Unquote(value)
		case "setpage_suffix":
			out.SetpageSuffix = Unquote(value)
		case "group_skip":
			out.GroupSkip = Unquote(value)
		case "headings_flag", "lethead_flag":
			out.HeadingsFlag = ParseInt(value)
		case "heading_prefix", "lethead_prefix":
			out.HeadingPrefix = Unquote(value)
		case "heading_suffix", "lethead_suffix":
			out.HeadingSuffix = Unquote(value)
		case "symhead_positive":
			out.SymheadPositive = Unquote(value)
		case "symhead_negative":
			out.SymheadNegative = Unquote(value)
		case "numhead_positive":
			out.NumheadPositive = Unquote(value)
		case "numhead_negative":
			out.NumheadNegative = Unquote(value)
		case "stroke_prefix":
			out.StrokePrefix = Unquote(value)
		case "stroke_suffix":
			out.StrokeSuffix = Unquote(value)
		case "radical_prefix":
			out.RadicalPrefix = Unquote(value)
		case "radical_suffix":
			out.RadicalSuffix = Unquote(value)
		case "radical_simplify_flag":
			out.RadicalSimplifiedFlag = ParseInt(value)
		case "radical_simplified_prefix":
			out.RadicalSimplifiedPrefix = Unquote(value)
		case "radical_simplified_suffix":
			out.RadicalSimplifiedSuffix = Unquote(value)
		case "item_0":
			out.Item0 = Unquote(value)
		case "item_1":
			out.Item1 = Unquote(value)
		case "item_2":
			out.Item2 = Unquote(value)
		case "item_01":
			out.Item01 = Unquote(value)
		case "item_x1":
			out.ItemX1 = Unquote(value)
		case "item_12":
			out.Item12 = Unquote(value)
		case "item_x2":
			out.ItemX2 = Unquote(value)
		case "delim_0":
			out.Delim0 = Unquote(value)
		case "delim_1":
			out.Delim1 = Unquote(value)
		case "delim_2":
			out.Delim2 = Unquote(value)
		case "delim_n":
			out.DelimN = Unquote(value)
		case "delim_r":
			out.DelimR = Unquote(value)
		case "delim_t":
			out.DelimT = Unquote(value)
		case "encap_prefix":
			out.EncapPrefix = Unquote(value)
		case "encap_infix":
			out.EncapInfix = Unquote(value)
		case "encap_suffix":
			out.EncapSuffix = Unquote(value)
		case "line_max":
			out.LineMax = ParseInt(value)
		case "indent_space":
			out.IndentSpace = Unquote(value)
		case "indent_length":
			out.IndentLength = ParseInt(value)
		case "suffix_2p":
			out.Suffix2p = Unquote(value)
		case "suffix_3p":
			out.Suffix3p = Unquote(value)
		case "suffix_mp":
			out.SuffixMp = Unquote(value)
		default:
			log.Printf("忽略未知格式 %s\n", key)
		}
	}
	return in, out
}

func Unquote(src string) string {
	if src == "" {
		return ""
	}
	if src[0] == '"' {
		src = strings.Replace(src, "\n", "\\n", -1)
	}
	dst, err := strconv.Unquote(src)
	if err != nil {
		log.Println(err.Error())
	}
	return dst
}

func UnquoteChar(src string) rune {
	src = Unquote(src)
	dst, _, tail, err := strconv.UnquoteChar(src, 0)
	if tail != "" {
		err = strconv.ErrSyntax
	}
	if err != nil {
		log.Println(err.Error())
	}
	return dst
}

func ParseInt(src string) int {
	i, err := strconv.ParseInt(src, 0, 0)
	if err != nil {
		log.Println(err.Error())
	}
	return int(i)
}

func ScanStyleTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0
	in_comment := false
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !in_comment && r == '%' {
			in_comment = true
		} else if in_comment && r == '\n' {
			in_comment = false
		} else if !in_comment && !unicode.IsSpace(r) {
			break
		}
	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	switch first, firstwidth := utf8.DecodeRune(data[start:]); first {
	case '\'', '"', '`':
		for width, i := 0, start+firstwidth; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if r == '\\' {
				_, newwidth := utf8.DecodeRune(data[i+width:])
				width += newwidth
			} else if r == first {
				return i + width, data[start : i+width], nil
			}
		}
	default:
		for width, i := 0, start; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if unicode.IsSpace(r) || r == '%' {
				return i, data[start:i], nil
			}
		}
	}
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	return 0, nil, nil
}
