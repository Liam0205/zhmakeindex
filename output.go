package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/text/transform"

	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

type OutputIndex struct {
	groups []index.IndexGroup
	style  *style.OutputStyle
	option *OutputOptions
}

func NewOutputIndex(input *index.InputIndex, option *OutputOptions, style *style.OutputStyle) *OutputIndex {
	sorter := NewIndexSorter(option.sort)
	outindex := sorter.SortIndex(input, style, option)
	outindex.style = style
	outindex.option = option
	return outindex
}

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
		if group.Items == nil {
			continue
		}
		if first_group {
			first_group = false
		} else {
			fmt.Fprint(writer, o.style.GroupSkip)
		}
		if o.style.HeadingsFlag != 0 {
			fmt.Fprintf(writer, "%s%s%s", o.style.HeadingPrefix, group.Name, o.style.HeadingSuffix)
		}
		for i, item := range group.Items {
			switch item.Level {
			case 0:
				fmt.Fprintf(writer, "%s%s", o.style.Item0, item.Text)
				writePage(writer, 0, item.Page, o.style)
			case 1:
				if last := group.Items[i-1]; last.Level == 0 {
					if last.Page != nil {
						fmt.Fprint(writer, o.style.Item01)
					} else {
						fmt.Fprint(writer, o.style.ItemX1)
					}
				} else {
					fmt.Fprint(writer, o.style.Item1)
				}
				fmt.Fprint(writer, item.Text)
				writePage(writer, 1, item.Page, o.style)
			case 2:
				if last := group.Items[i-1]; last.Level == 1 {
					if last.Page != nil {
						fmt.Fprint(writer, o.style.Item12)
					} else {
						fmt.Fprint(writer, o.style.ItemX2)
					}
				} else {
					fmt.Fprint(writer, o.style.Item2)
				}
				fmt.Fprint(writer, item.Text)
				writePage(writer, 2, item.Page, o.style)
			default:
				log.Printf("索引项\u201c%s\u201d层次数过深，忽略此项\n", item.Text)
			}
		}
	}
	fmt.Fprint(writer, o.style.Postamble)
}

func writePage(out io.Writer, level int, pageranges []index.PageRange, st *style.OutputStyle) {
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
