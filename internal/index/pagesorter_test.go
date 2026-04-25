package index

import (
	"testing"

	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

func TestMergeDisableRangeDedupByValue(t *testing.T) {
	s := style.NewOutputStyle()
	sorter := NewPageSorter(s, false, true)

	p1 := &page.Page{
		Numbers:    []page.PageNumber{{Format: page.NUM_ARABIC, Num: 42}},
		Compositor: "-",
		Encap:      "textbf",
		Rangetype:  page.PAGE_NORMAL,
	}
	p2 := &page.Page{
		Numbers:    []page.PageNumber{{Format: page.NUM_ARABIC, Num: 42}},
		Compositor: "-",
		Encap:      "textbf",
		Rangetype:  page.PAGE_NORMAL,
	}

	if p1 == p2 {
		t.Fatal("test setup error: p1 and p2 must be different pointers")
	}

	pages := []PageRange{
		{Begin: p1, End: p1},
		{Begin: p2, End: p2},
	}

	got := sorter.Merge(pages)
	if len(got) != 1 {
		t.Fatalf("Merge() with disable_range returned %d ranges, want 1 (duplicate not deduplicated)", len(got))
	}
}

func TestNewPageSorterPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		style   *style.OutputStyle
		strict  bool
		format  page.NumFormat
		wantPos int
	}{
		{
			name:    "default roman lower precedence first",
			style:   style.NewOutputStyle(),
			strict:  false,
			format:  page.NUM_ROMAN_LOWER,
			wantPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewPageSorter(tt.style, tt.strict, false)
			if got := sorter.precedence[tt.format]; got != tt.wantPos {
				t.Fatalf("precedence[%v] = %d, want %d", tt.format, got, tt.wantPos)
			}
		})
	}
}
