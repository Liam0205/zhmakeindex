package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/text/transform"

	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

// 输出索引
type OutputIndex struct {
	groups []IndexGroup
	style  *style.OutputStyle
	option *OutputOptions
}

func NewOutputIndex(input *InputIndex, option *OutputOptions, style *style.OutputStyle) *OutputIndex {
	sorter := NewIndexSorter(option.sort)
	outindex := sorter.SortIndex(input, style, option)
	outindex.style = style
	outindex.option = option
	return outindex
}

// 按格式输出索引项
// suffix_2p, suffix_3p, suffix_mp 暂未实现
// line_max, indent_space, indent_length 未实现
func (o *OutputIndex) Output(option *OutputOptions) {
	var writer io.WriteCloser
	if o.option.output == "" {
		writer = os.Stdout
	} else {
		var err error
		writer, err = os.Create(o.option.output)
		if err != nil {
			log.Fatalln(err)
		}
		defer writer.Close()
	}
	writer = transform.NewWriter(writer, option.encoder)

	fmt.Fprint(writer, o.style.Preamble)
	first_group := true
	for _, group := range o.groups {
		if group.items == nil {
			continue
		}
		if first_group {
			first_group = false
		} else {
			fmt.Fprint(writer, o.style.GroupSkip)
		}
		if o.style.HeadingsFlag != 0 {
			fmt.Fprintf(writer, "%s%s%s", o.style.HeadingPrefix, group.name, o.style.HeadingSuffix)
		}
		for i, item := range group.items {
			switch item.level {
			case 0:
				fmt.Fprintf(writer, "%s%s", o.style.Item0, item.text)
				writePage(writer, 0, item.page, o.style)
			case 1:
				if last := group.items[i-1]; last.level == 0 {
					if last.page != nil {
						fmt.Fprint(writer, o.style.Item01)
					} else {
						fmt.Fprint(writer, o.style.ItemX1)
					}
				} else {
					fmt.Fprint(writer, o.style.Item1)
				}
				fmt.Fprint(writer, item.text)
				writePage(writer, 1, item.page, o.style)
			case 2:
				if last := group.items[i-1]; last.level == 1 {
					if last.page != nil {
						fmt.Fprint(writer, o.style.Item12)
					} else {
						fmt.Fprint(writer, o.style.ItemX2)
					}
				} else {
					fmt.Fprint(writer, o.style.Item2)
				}
				fmt.Fprint(writer, item.text)
				writePage(writer, 2, item.page, o.style)
			default:
				log.Printf("索引项\u201c%s\u201d层次数过深，忽略此项\n", item.text)
			}
		}
	}
	fmt.Fprint(writer, o.style.Postamble)
}

func writePage(out io.Writer, level int, pageranges []PageRange, st *style.OutputStyle) {
	if pageranges == nil {
		return
	}
	switch level {
	case 0:
		fmt.Fprint(out, st.Delim0)
	case 1:
		fmt.Fprint(out, st.Delim1)
	case 2:
		fmt.Fprint(out, st.Delim2)
	}
	for i, p := range pageranges {
		if i > 0 {
			fmt.Fprint(out, st.DelimN)
		}
		p.Write(out, st)
	}
	if len(pageranges) != 0 {
		fmt.Fprint(out, st.DelimT)
	}
}

// 一个输出项目组
type IndexGroup struct {
	name  string
	items []IndexItem
}

// 一个输出项，包括级别、文字、一系列页码区间
type IndexItem struct {
	level int
	text  string
	page  []PageRange
}

// 用于输出的页码区间
type PageRange struct {
	begin *page.Page
	end   *page.Page
}

func (p *PageRange) Diff() int {
	return p.end.Diff(p.begin)
}

// 输出页码区间
func (p *PageRange) Write(out io.Writer, st *style.OutputStyle) {
	var rangestr string
	switch {
	case p.Diff() == 0:
		rangestr = p.begin.String()
	case p.begin.Rangetype == page.PAGE_NORMAL && p.end.Rangetype == page.PAGE_NORMAL &&
		p.Diff() == 1 && st.Suffix2p == "":
		rangestr = p.begin.String() + st.DelimN + p.end.String()
	case p.Diff() == 1 && st.Suffix2p != "":
		rangestr = p.begin.String() + st.Suffix2p
	case p.Diff() == 2 && st.Suffix3p != "":
		rangestr = p.begin.String() + st.Suffix3p
	case p.Diff() >= 2 && st.SuffixMp != "":
		rangestr = p.begin.String() + st.SuffixMp
	default:
		rangestr = p.begin.String() + st.DelimR + p.end.String()
	}
	if p.begin.Encap == "" {
		fmt.Fprint(out, rangestr)
	} else {
		fmt.Fprint(out, st.EncapPrefix, p.begin.Encap,
			st.EncapInfix, rangestr, st.EncapSuffix)
	}
}
