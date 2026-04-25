package index

import (
	"testing"

	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

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
