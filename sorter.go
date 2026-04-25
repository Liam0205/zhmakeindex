package main

import (
	"log"
	"sort"

	"github.com/leo-liu/zhmakeindex/internal/collator"
	"github.com/leo-liu/zhmakeindex/internal/index"
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

	pagesorter := index.NewPageSorter(style, option.strict, option.disable_range)
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
