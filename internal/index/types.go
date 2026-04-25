package index

import (
	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
	"github.com/yasushi-saito/rbtree"
)

type IndexEntry struct {
	Input    string
	Level    []IndexEntryLevel
	Pagelist []*page.Page
}

// 实现 rbtree.CompareFunc
func CompareIndexEntry(a, b rbtree.Item) int {
	x := a.(*IndexEntry)
	y := b.(*IndexEntry)
	for i := range x.Level {
		if i >= len(y.Level) {
			return 1
		}
		if x.Level[i].Key < y.Level[i].Key {
			return -1
		} else if x.Level[i].Key > y.Level[i].Key {
			return 1
		}
		if x.Level[i].Text < y.Level[i].Text {
			return -1
		} else if x.Level[i].Text > y.Level[i].Text {
			return 1
		}
	}
	if len(x.Level) < len(y.Level) {
		return -1
	}
	return 0
}

type IndexEntryLevel struct {
	Key  string
	Text string
}

type InputIndex []IndexEntry

// 对应不同的分类排序方式
type IndexCollator interface {
	InitGroups(style *style.OutputStyle) []IndexGroup
	Group(entry *IndexEntry) int
	RuneCmp(a, b rune) int
	IsLetter(r rune) bool
}

type IndexGroup struct {
	Name  string
	Items []IndexItem
}

type IndexItem struct {
	Level int
	Text  string
	Page  []PageRange
}
