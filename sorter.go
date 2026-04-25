package main

import (
	"log"
	"sort"

	"github.com/leo-liu/zhmakeindex/internal/collator"
	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

type IndexSorter struct {
	index.IndexCollator
}

func NewIndexSorter(method string) *IndexSorter {
	switch method {
	case "bihua", "stroke":
		return &IndexSorter{
			IndexCollator: collator.StrokeIndexCollator{},
		}
	case "pinyin", "reading":
		return &IndexSorter{
			IndexCollator: collator.ReadingIndexCollator{},
		}
	case "bushou", "radical":
		return &IndexSorter{
			IndexCollator: collator.RadicalIndexCollator{},
		}
	default:
		log.Fatalln("未知排序方式")
	}
	return nil
}

func (sorter *IndexSorter) SortIndex(input *index.InputIndex, style *style.OutputStyle, option *OutputOptions) *OutputIndex {
	out := new(OutputIndex)
	out.groups = sorter.InitGroups(style)

	sort.Sort(index.IndexEntrySlice{
		Entries:  *input,
		Colattor: sorter.IndexCollator,
	})

	pagesorter := NewPageSorter(style, option)
	for _, entry := range *input {
		pageranges := pagesorter.Sort(entry)
		pageranges = pagesorter.Merge(pageranges)
		item := index.IndexItem{
			Level: len(entry.Level) - 1,
			Text:  entry.Level[len(entry.Level)-1].Text,
			Page:  pageranges,
		}
		group := sorter.Group(&entry)
		out.groups[group].Items = append(out.groups[group].Items, item)
	}

	return out
}

type PageSorter struct {
	precedence    map[page.NumFormat]int
	strict        bool
	disable_range bool
}

func NewPageSorter(style *style.OutputStyle, option *OutputOptions) *PageSorter {
	var sorter PageSorter
	sorter.precedence = make(map[page.NumFormat]int)
	for i, r := range style.PagePrecedence {
		switch r {
		case 'r':
			sorter.precedence[page.NUM_ROMAN_LOWER] = i
		case 'n':
			sorter.precedence[page.NUM_ARABIC] = i
		case 'a':
			sorter.precedence[page.NUM_ALPH_LOWER] = i
		case 'R':
			sorter.precedence[page.NUM_ROMAN_UPPER] = i
		case 'A':
			sorter.precedence[page.NUM_ALPH_UPPER] = i
		default:
			log.Println("page_precedence 语法错误，采用默认值")
			sorter.precedence = map[page.NumFormat]int{
				page.NUM_ROMAN_LOWER: 0,
				page.NUM_ARABIC:      1,
				page.NUM_ALPH_LOWER:  2,
				page.NUM_ROMAN_UPPER: 3,
				page.NUM_ALPH_UPPER:  4,
			}
		}
	}
	sorter.strict = option.strict
	sorter.disable_range = option.disable_range
	return &sorter
}

func (sorter *PageSorter) Sort(entry index.IndexEntry) []index.PageRange {
	pages := entry.Pagelist
	var out []index.PageRange
	if sorter.strict {
		sort.Sort(PageSliceStrict{
			PageSlice{pages: pages, sorter: sorter}})
	} else {
		sort.Sort(PageSliceLoose{
			PageSlice{pages: pages, sorter: sorter}})
	}
	var stack []*page.Page
	for i := 0; i < len(pages); i++ {
		p := pages[i]
		if len(stack) == 0 {
			switch p.Rangetype {
			case page.PAGE_NORMAL:
				out = append(out, index.PageRange{Begin: p, End: p})
			case page.PAGE_OPEN:
				stack = append(stack, p)
			case page.PAGE_CLOSE:
				log.Printf("条目 %s 的页码区间有误，区间末尾 %s{%s} 没有匹配的区间头。\n", entry.Input, p.Encap, p)
				out = append(out, index.PageRange{Begin: p.Empty(), End: p})
			}
		} else {
			front := stack[0]
			top := stack[len(stack)-1]
			if p.Encap != front.Encap {
				if sorter.strict {
					log.Printf("条目 %s 的页码区间可能有误，区间头 %s 没有对应的区间尾\n", entry.Input, front)
					out = append(out, index.PageRange{Begin: front, End: front.Empty()})
					stack = nil
					i--
					continue
				} else {
					if p.Rangetype == page.PAGE_NORMAL {
						out = append(out, index.PageRange{Begin: p, End: p})
					} else {
						log.Printf("条目 %s 的页码区间 %s{%s--} 内 %s%s{%s} 命令格式不同，可能丢失信息",
							entry.Input, front.Encap, front, p.Rangetype, p.Encap, p)
					}
				}
			} else if !p.Compatible(top) {
				log.Printf("条目 %s 的页码区间 %s{%s -- %s} 跨过不同的数字格式\n", entry.Input, top.Encap, top, p)
			}
			switch p.Rangetype {
			case page.PAGE_NORMAL:
			case page.PAGE_OPEN:
				stack = append(stack, p)
			case page.PAGE_CLOSE:
				if len(stack) == 1 {
					out = append(out, index.PageRange{Begin: front, End: p})
				}
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) > 0 {
		log.Printf("条目 %s 的页码区间有误，未找到与 %s{%s} 匹配的区间尾。\n", entry.Input, stack[0].Encap, stack[0])
		out = append(out, index.PageRange{Begin: stack[0], End: stack[0].Empty()})
	}
	return out
}

func (sorter *PageSorter) Merge(pages []index.PageRange) []index.PageRange {
	var out []index.PageRange
	for i, r := range pages {
		if i == 0 {
			out = append(out, r)
			continue
		}
		prev := out[len(out)-1]
		if sorter.disable_range &&
			(r.Begin.Rangetype == page.PAGE_NORMAL || prev.Begin.Rangetype == page.PAGE_NORMAL) {
			if prev.Begin == r.Begin {
				continue
			} else {
				out = append(out, r)
			}
		} else if prev.Begin.Encap == r.Begin.Encap &&
			r.Begin.Compatible(prev.Begin) &&
			r.Begin.Diff(prev.End) <= 1 {
			out[len(out)-1].End = r.End
		} else {
			out = append(out, r)
		}
	}
	for i := range out {
		if out[i].Begin.Encap == out[i].End.Encap {
			if out[i].Begin.Diff(out[i].End) == 0 {
				out[i].Begin.Rangetype = page.PAGE_NORMAL
				out[i].End.Rangetype = page.PAGE_NORMAL
			}
		}
	}
	return out
}

type PageSlice struct {
	pages  []*page.Page
	sorter *PageSorter
}

func (p PageSlice) Len() int {
	return len(p.pages)
}

func (p PageSlice) Swap(i, j int) {
	p.pages[i], p.pages[j] = p.pages[j], p.pages[i]
}

type PageSliceStrict struct {
	PageSlice
}

func (p PageSliceStrict) Less(i, j int) bool {
	a, b := p.pages[i], p.pages[j]
	if a.Encap < b.Encap {
		return true
	} else if a.Encap > b.Encap {
		return false
	}
	if cmp := a.Cmp(b, p.sorter.precedence); cmp != 0 {
		return cmp < 0
	}
	if a.Rangetype < b.Rangetype {
		return true
	} else {
		return false
	}
}

type PageSliceLoose struct {
	PageSlice
}

func (p PageSliceLoose) Less(i, j int) bool {
	a, b := p.pages[i], p.pages[j]
	if cmp := a.Cmp(b, p.sorter.precedence); cmp != 0 {
		return cmp < 0
	}
	if a.Rangetype < b.Rangetype {
		return true
	} else if a.Rangetype > b.Rangetype {
		return false
	}
	if a.Encap < b.Encap {
		return true
	} else {
		return false
	}
}
