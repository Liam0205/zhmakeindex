package main

import (
	"io"
	"strings"
	"testing"

	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/reader"
	"github.com/leo-liu/zhmakeindex/internal/style"
	"github.com/yasushi-saito/rbtree"
)

func TestScanIndexEntry(t *testing.T) {
	style := style.NewInputStyle()
	option := &InputOptions{}

	tests := []struct {
		name      string
		input     string
		levels    []index.IndexEntryLevel
		encap     string
		rangetype page.RangeType
		pageNum   int
		pageFmt   page.NumFormat
	}{
		{
			name:      "simple entry",
			input:     `\indexentry{foo}{1}`,
			levels:    []index.IndexEntryLevel{{Key: "foo", Text: "foo"}},
			rangetype: page.PAGE_NORMAL,
			pageNum:   1,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:  "multi-level entry",
			input: `\indexentry{A!B!C}{5}`,
			levels: []index.IndexEntryLevel{
				{Key: "A", Text: "A"},
				{Key: "B", Text: "B"},
				{Key: "C", Text: "C"},
			},
			rangetype: page.PAGE_NORMAL,
			pageNum:   5,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:      "key@text syntax",
			input:     `\indexentry{key@text}{2}`,
			levels:    []index.IndexEntryLevel{{Key: "key", Text: "text"}},
			rangetype: page.PAGE_NORMAL,
			pageNum:   2,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:      "encap syntax",
			input:     `\indexentry{foo|textbf}{3}`,
			levels:    []index.IndexEntryLevel{{Key: "foo", Text: "foo"}},
			encap:     "textbf",
			rangetype: page.PAGE_NORMAL,
			pageNum:   3,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:      "range open",
			input:     `\indexentry{foo|(}{10}`,
			levels:    []index.IndexEntryLevel{{Key: "foo", Text: "foo"}},
			rangetype: page.PAGE_OPEN,
			pageNum:   10,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:      "range close",
			input:     `\indexentry{foo|)}{15}`,
			levels:    []index.IndexEntryLevel{{Key: "foo", Text: "foo"}},
			rangetype: page.PAGE_CLOSE,
			pageNum:   15,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:      "range open with encap",
			input:     `\indexentry{foo|(textit}{7}`,
			levels:    []index.IndexEntryLevel{{Key: "foo", Text: "foo"}},
			encap:     "textit",
			rangetype: page.PAGE_OPEN,
			pageNum:   7,
			pageFmt:   page.NUM_ARABIC,
		},
		{
			name:      "roman numeral page",
			input:     `\indexentry{bar}{xiv}`,
			levels:    []index.IndexEntryLevel{{Key: "bar", Text: "bar"}},
			rangetype: page.PAGE_NORMAL,
			pageNum:   14,
			pageFmt:   page.NUM_ROMAN_LOWER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := reader.NewNumberdReader(strings.NewReader(tt.input))
			entry, err := ScanIndexEntry(reader, option, style)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entry.Level) != len(tt.levels) {
				t.Fatalf("expected %d levels, got %d", len(tt.levels), len(entry.Level))
			}
			for i, lv := range tt.levels {
				if entry.Level[i].Key != lv.Key {
					t.Errorf("level[%d].key = %q, want %q", i, entry.Level[i].Key, lv.Key)
				}
				if entry.Level[i].Text != lv.Text {
					t.Errorf("level[%d].text = %q, want %q", i, entry.Level[i].Text, lv.Text)
				}
			}
			if len(entry.Pagelist) != 1 {
				t.Fatalf("expected 1 page, got %d", len(entry.Pagelist))
			}
			pg := entry.Pagelist[0]
			if pg.Encap != tt.encap {
				t.Errorf("encap = %q, want %q", pg.Encap, tt.encap)
			}
			if pg.Rangetype != tt.rangetype {
				t.Errorf("rangetype = %v, want %v", pg.Rangetype, tt.rangetype)
			}
			if len(pg.Numbers) < 1 {
				t.Fatal("no page numbers")
			}
			if pg.Numbers[0].Num != tt.pageNum {
				t.Errorf("page num = %d, want %d", pg.Numbers[0].Num, tt.pageNum)
			}
			if pg.Numbers[0].Format != tt.pageFmt {
				t.Errorf("page format = %d, want %d", pg.Numbers[0].Format, tt.pageFmt)
			}
		})
	}
}

func TestScanIndexEntryEOF(t *testing.T) {
	style := style.NewInputStyle()
	option := &InputOptions{}
	reader := reader.NewNumberdReader(strings.NewReader(""))
	_, err := ScanIndexEntry(reader, option, style)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestScanIndexEntrySyntaxError(t *testing.T) {
	style := style.NewInputStyle()
	option := &InputOptions{}
	reader := reader.NewNumberdReader(strings.NewReader("not an index entry"))
	_, err := ScanIndexEntry(reader, option, style)
	if err != page.ScanSyntaxError {
		t.Errorf("expected page.ScanSyntaxError, got %v", err)
	}
}

func TestCompareIndexEntry(t *testing.T) {
	tests := []struct {
		name string
		a, b *index.IndexEntry
		want int
	}{
		{
			name: "equal entries",
			a:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "foo", Text: "foo"}}},
			b:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "foo", Text: "foo"}}},
			want: 0,
		},
		{
			name: "a less than b by key",
			a:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "abc", Text: "abc"}}},
			b:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "def", Text: "def"}}},
			want: -1,
		},
		{
			name: "parent less than child",
			a:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "A", Text: "A"}}},
			b:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "A", Text: "A"}, {Key: "B", Text: "B"}}},
			want: -1,
		},
		{
			name: "child greater than parent",
			a:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "A", Text: "A"}, {Key: "B", Text: "B"}}},
			b:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "A", Text: "A"}}},
			want: 1,
		},
		{
			name: "same key different text",
			a:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "foo", Text: "aaa"}}},
			b:    &index.IndexEntry{Level: []index.IndexEntryLevel{{Key: "foo", Text: "zzz"}}},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := index.CompareIndexEntry(rbtree.Item(tt.a), rbtree.Item(tt.b))
			if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
				t.Errorf("CompareIndexEntry() = %d, want sign of %d", got, tt.want)
			}
		})
	}
}

func TestSkipspaces(t *testing.T) {
	style := style.NewInputStyle()

	tests := []struct {
		name     string
		input    string
		wantRune rune
	}{
		{
			name:     "skip whitespace",
			input:    "  \n  foo",
			wantRune: 'f',
		},
		{
			name:     "skip comment",
			input:    "%comment\nfoo",
			wantRune: 'f',
		},
		{
			name:     "no whitespace",
			input:    "bar",
			wantRune: 'b',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := reader.NewNumberdReader(strings.NewReader(tt.input))
			err := skipspaces(reader, style)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			r, _, err := reader.ReadRune()
			if err != nil {
				t.Fatalf("ReadRune error: %v", err)
			}
			if r != tt.wantRune {
				t.Errorf("next rune = %c, want %c", r, tt.wantRune)
			}
		})
	}
}
