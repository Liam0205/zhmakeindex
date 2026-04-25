package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/transform"

	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/reader"
	"github.com/leo-liu/zhmakeindex/internal/style"
	"github.com/yasushi-saito/rbtree"
)

type InputIndex []IndexEntry

func NewInputIndex(option *InputOptions, instyle *style.InputStyle) *InputIndex {
	inset := rbtree.NewTree(CompareIndexEntry)

	if option.stdin {
		readIdxFile(inset, os.Stdin, option, instyle)
	} else {
		for _, idxname := range option.input {
			// 文件不存在且无后缀时，加上默认后缀 .idx 再试
			if _, err := os.Stat(idxname); os.IsNotExist(err) && filepath.Ext(idxname) == "" {
				idxname = idxname + ".idx"
			}
			idxfile, err := os.Open(idxname)
			if err != nil {
				log.Fatalln(err.Error())
			}
			readIdxFile(inset, idxfile, option, instyle)
			idxfile.Close()
		}
	}

	var in InputIndex
	for iter := inset.Min(); !iter.Limit(); iter = iter.Next() {
		pentry := iter.Item().(*IndexEntry)
		in = append(in, *pentry)
	}
	return &in
}

func readIdxFile(inset *rbtree.Tree, idxfile *os.File, option *InputOptions, instyle *style.InputStyle) {
	log.Printf("读取输入文件 %s ……\n", idxfile.Name())
	accepted, rejected := 0, 0

	idxreader := reader.NewNumberdReader(transform.NewReader(idxfile, option.decoder))
	for {
		entry, err := ScanIndexEntry(idxreader, option, instyle)
		if err == io.EOF {
			break
		} else if err == page.ScanSyntaxError {
			rejected++
			log.Printf("%s:%d: %s\n", idxfile.Name(), idxreader.Line(), err.Error())
			// 跳过一行
			if err := idxreader.SkipLine(); err == io.EOF {
				break
			} else if err != nil {
				log.Fatalln(err.Error())
			}
		} else if err != nil {
			log.Fatalln(err.Error())
		} else {
			accepted++
			if old := inset.Get(entry); old != nil {
				oldentry := old.(*IndexEntry)
				oldentry.pagelist = append(oldentry.pagelist, entry.pagelist...)
			} else {
				// entry 不在集合 inset 中时，插入 entry 本身和所有祖先节点，祖先不含页码
				for len(entry.level) > 0 {
					inset.Insert(entry)
					parent := &IndexEntry{
						level:    entry.level[:len(entry.level)-1],
						pagelist: nil,
					}
					if inset.Get(parent) != nil {
						break
					} else {
						entry = parent
					}
				}
			}
		}
	}
	log.Printf("接受 %d 项，拒绝 %d 项。\n", accepted, rejected)
}

// 跳过空白符和行注释
func skipspaces(rd *reader.NumberdReader, st *style.InputStyle) error {
	for {
		r, _, err := rd.ReadRune()
		if err != nil {
			return err
		} else if r == st.Comment {
			rd.SkipLine()
		} else if !unicode.IsSpace(r) {
			rd.UnreadRune()
			break
		}
	}
	return nil
}

func ScanIndexEntry(rd *reader.NumberdReader, option *InputOptions, st *style.InputStyle) (*IndexEntry, error) {
	var entry IndexEntry
	pg := new(page.Page)
	if err := skipspaces(rd, st); err != nil {
		return nil, err
	}
	for _, r := range st.Keyword {
		new_r, _, err := rd.ReadRune()
		if err != nil {
			return nil, err
		}
		if new_r != r {
			return nil, page.ScanSyntaxError
		}
	}
	if err := skipspaces(rd, st); err != nil {
		return nil, err
	}
	const (
		SCAN_OPEN = iota
		SCAN_KEY
		SCAN_VALUE
		SCAN_COMMAND
		SCAN_PAGE
		SCAN_PAGERANGE
	)
	state := SCAN_OPEN
	quoted := false
	escaped := false
	arg_depth := 0
	var token []rune
	var entry_input []rune
	pg.Rangetype = page.PAGE_NORMAL
L_scan_kv:
	for {
		r, _, err := rd.ReadRune()
		entry_input = append(entry_input, r)
		if err != nil {
			return nil, err
		}
		switch state {
		case SCAN_OPEN:
			if !quoted && r == st.ArgOpen {
				state = SCAN_KEY
			} else {
				return nil, page.ScanSyntaxError
			}
		case SCAN_KEY:
			push_keyval := func(next int) {
				str := string(token)
				if option.compress {
					str = strings.TrimSpace(str)
				}
				entry.level = append(entry.level, IndexEntryLevel{key: str, text: str})
				token = nil
				state = next
			}
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == st.ArgOpen && !escaped {
				token = append(token, r)
				arg_depth++
			} else if r == st.ArgClose && !escaped {
				if arg_depth == 0 {
					push_keyval(0)
					break L_scan_kv
				} else {
					token = append(token, r)
					arg_depth--
				}
			} else if r == st.Actual {
				push_keyval(SCAN_VALUE)
			} else if r == st.Encap {
				push_keyval(SCAN_PAGERANGE)
			} else if r == st.Level {
				push_keyval(SCAN_KEY)
			} else if r == st.Quote && !escaped {
				quoted = true
			} else {
				token = append(token, r)
			}
			if r == st.Escape {
				escaped = true
			} else {
				escaped = false
			}
		case SCAN_VALUE:
			set_value := func(next int) {
				str := string(token)
				entry.level[len(entry.level)-1].text = str
				token = nil
				state = next
			}
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == st.ArgOpen && !escaped {
				token = append(token, r)
				arg_depth++
			} else if r == st.ArgClose && !escaped {
				if arg_depth == 0 {
					set_value(0)
					break L_scan_kv
				} else {
					token = append(token, r)
					arg_depth--
				}
			} else if r == st.Encap {
				set_value(SCAN_PAGERANGE)
			} else if r == st.Level {
				set_value(SCAN_KEY)
			} else if r == st.Quote && !escaped {
				quoted = true
			} else {
				token = append(token, r)
			}
			if r == st.Escape {
				escaped = true
			} else {
				escaped = false
			}
		case SCAN_PAGERANGE:
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == st.ArgOpen || r == st.ArgClose || r == st.Actual || r == st.Encap || r == st.Level {
				return nil, page.ScanSyntaxError
			} else if r == st.RangeOpen {
				pg.Rangetype = page.PAGE_OPEN
			} else if r == st.RangeClose {
				pg.Rangetype = page.PAGE_CLOSE
			} else if r == st.Quote {
				quoted = true
			} else {
				token = append(token, r)
			}
			state = SCAN_COMMAND
			if r == st.Escape {
				escaped = true
			} else {
				escaped = false
			}
		case SCAN_COMMAND:
			if quoted {
				token = append(token, r)
				quoted = false
				break
			} else if r == st.ArgOpen && !escaped {
				token = append(token, r)
				arg_depth++
			} else if r == st.ArgClose && !escaped {
				if arg_depth == 0 {
					pg.Encap = string(token)
					break L_scan_kv
				} else {
					token = append(token, r)
					arg_depth--
				}
			} else if r == st.Quote && !escaped {
				quoted = true
			} else {
				token = append(token, r)
			}
			if r == st.Escape {
				escaped = true
			} else {
				escaped = false
			}
		default:
			panic("扫描状态错误")
		}
	}
	entry.input = string(entry_input)
	if err := skipspaces(rd, st); err != nil {
		return nil, err
	}
	state = SCAN_OPEN
	token = nil
L_scan_page:
	for {
		r, _, err := rd.ReadRune()
		if err != nil {
			return nil, err
		}
		switch state {
		case SCAN_OPEN:
			if r == st.ArgOpen {
				state = SCAN_PAGE
			} else {
				return nil, page.ScanSyntaxError
			}
		case SCAN_PAGE:
			if r == st.ArgClose {
				pg.Numbers, err = page.ScanPage(token, st.PageCompositor)
				if err != nil {
					return nil, err
				}
				break L_scan_page
			} else if r == st.ArgOpen {
				return nil, page.ScanSyntaxError
			} else {
				token = append(token, r)
			}
		default:
			panic("扫描状态错误")
		}
	}
	pg.Compositor = st.PageCompositor
	entry.pagelist = append(entry.pagelist, pg)
	return &entry, nil
}

type IndexEntry struct {
	input    string
	level    []IndexEntryLevel
	pagelist []*page.Page
}

// 实现 rbtree.CompareFunc
func CompareIndexEntry(a, b rbtree.Item) int {
	x := a.(*IndexEntry)
	y := b.(*IndexEntry)
	for i := range x.level {
		if i >= len(y.level) {
			return 1
		}
		if x.level[i].key < y.level[i].key {
			return -1
		} else if x.level[i].key > y.level[i].key {
			return 1
		}
		if x.level[i].text < y.level[i].text {
			return -1
		} else if x.level[i].text > y.level[i].text {
			return 1
		}
	}
	if len(x.level) < len(y.level) {
		return -1
	}
	return 0
}

// 一条索引条目中的一级
type IndexEntryLevel struct {
	key  string
	text string
}
