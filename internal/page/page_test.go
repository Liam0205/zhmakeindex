package page

import (
	"testing"
)

func TestScanNumber(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    PageNumber
		wantErr error
	}{
		{name: "arabic one", token: "1", want: PageNumber{Format: NUM_ARABIC, Num: 1}},
		{name: "arabic forty two", token: "42", want: PageNumber{Format: NUM_ARABIC, Num: 42}},
		{name: "arabic zero", token: "0", want: PageNumber{Format: NUM_ARABIC, Num: 0}},
		{name: "roman lower one", token: "i", want: PageNumber{Format: NUM_ROMAN_LOWER, Num: 1}},
		{name: "roman lower four", token: "iv", want: PageNumber{Format: NUM_ROMAN_LOWER, Num: 4}},
		{name: "roman lower fourteen", token: "xiv", want: PageNumber{Format: NUM_ROMAN_LOWER, Num: 14}},
		{name: "roman upper three", token: "III", want: PageNumber{Format: NUM_ROMAN_UPPER, Num: 3}},
		{name: "roman upper fourteen", token: "XIV", want: PageNumber{Format: NUM_ROMAN_UPPER, Num: 14}},
		{name: "alpha lower a", token: "a", want: PageNumber{Format: NUM_ALPH_LOWER, Num: 1}},
		{name: "alpha lower z", token: "z", want: PageNumber{Format: NUM_ALPH_LOWER, Num: 26}},
		{name: "alpha upper A", token: "A", want: PageNumber{Format: NUM_ALPH_UPPER, Num: 1}},
		{name: "empty token", token: "", wantErr: ScanSyntaxError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScanNumber([]rune(tt.token))
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ScanNumber(%q) error = %v, want %v", tt.token, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ScanNumber(%q) unexpected error: %v", tt.token, err)
			}
			if got != tt.want {
				t.Fatalf("ScanNumber(%q) = %+v, want %+v", tt.token, got, tt.want)
			}
		})
	}
}

func TestScanPage(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		compositor string
		want       []PageNumber
		wantErr    error
	}{
		{
			name:       "composite page",
			token:      "5-ii-1",
			compositor: "-",
			want: []PageNumber{
				{Format: NUM_ARABIC, Num: 5},
				{Format: NUM_ROMAN_LOWER, Num: 2},
				{Format: NUM_ARABIC, Num: 1},
			},
		},
		{
			name:       "single page",
			token:      "42",
			compositor: "-",
			want: []PageNumber{{Format: NUM_ARABIC, Num: 42}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScanPage([]rune(tt.token), tt.compositor)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ScanPage(%q, %q) error = %v, want %v", tt.token, tt.compositor, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ScanPage(%q, %q) unexpected error: %v", tt.token, tt.compositor, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ScanPage(%q, %q) len = %d, want %d", tt.token, tt.compositor, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("ScanPage(%q, %q)[%d] = %+v, want %+v", tt.token, tt.compositor, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRomanNumString(t *testing.T) {
	tests := []struct {
		name  string
		num   int
		upper bool
		want  string
	}{
		{name: "lower 1", num: 1, want: "i"},
		{name: "lower 4", num: 4, want: "iv"},
		{name: "lower 9", num: 9, want: "ix"},
		{name: "lower 14", num: 14, want: "xiv"},
		{name: "lower 1999", num: 1999, want: "mcmxcix"},
		{name: "upper 1", num: 1, upper: true, want: "I"},
		{name: "upper 14", num: 14, upper: true, want: "XIV"},
		{name: "zero", num: 0, want: ""},
		{name: "negative", num: -1, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RomanNumString(tt.num, tt.upper)
			if got != tt.want {
				t.Fatalf("RomanNumString(%d, %v) = %q, want %q", tt.num, tt.upper, got, tt.want)
			}
		})
	}
}

func TestNumFormatFormatNum(t *testing.T) {
	tests := []struct {
		name   string
		format NumFormat
		num    int
		want   string
	}{
		{name: "arabic", format: NUM_ARABIC, num: 42, want: "42"},
		{name: "roman lower", format: NUM_ROMAN_LOWER, num: 4, want: "iv"},
		{name: "alpha lower", format: NUM_ALPH_LOWER, num: 1, want: "b"},
		{name: "alpha upper", format: NUM_ALPH_UPPER, num: 1, want: "B"},
		{name: "unknown", format: NUM_UNKNOWN, num: 0, want: "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.FormatNum(tt.num)
			if got != tt.want {
				t.Fatalf("%v.FormatNum(%d) = %q, want %q", tt.format, tt.num, got, tt.want)
			}
		})
	}
}

func TestPageString(t *testing.T) {
	p := &Page{
		Numbers: []PageNumber{
			{Format: NUM_ARABIC, Num: 5},
			{Format: NUM_ROMAN_LOWER, Num: 2},
			{Format: NUM_ARABIC, Num: 1},
		},
		Compositor: "-",
	}

	got := p.String()
	want := "5-ii-1"
	if got != want {
		t.Fatalf("Page.String() = %q, want %q", got, want)
	}
}

func TestPageCompatible(t *testing.T) {
	tests := []struct {
		name  string
		page  *Page
		other *Page
		want  bool
	}{
		{
			name:  "same formats are compatible",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ROMAN_LOWER, Num: 2}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 8}, {Format: NUM_ROMAN_LOWER, Num: 4}}},
			want:  true,
		},
		{
			name:  "different formats are incompatible",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ROMAN_LOWER, Num: 2}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ARABIC, Num: 2}}},
			want:  false,
		},
		{
			name:  "different lengths are incompatible",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 42}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 42}, {Format: NUM_ROMAN_LOWER, Num: 1}}},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Compatible(tt.other)
			if got != tt.want {
				t.Fatalf("Compatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageDiff(t *testing.T) {
	tests := []struct {
		name  string
		page  *Page
		other *Page
		want  int
	}{
		{
			name:  "compatible last segment differs",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ROMAN_LOWER, Num: 2}, {Format: NUM_ARABIC, Num: 1}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ROMAN_LOWER, Num: 2}, {Format: NUM_ARABIC, Num: 4}}},
			want:  3,
		},
		{
			name:  "incompatible returns minus one",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ROMAN_LOWER, Num: 5}}},
			want:  -1,
		},
		{
			name:  "non last segment differs returns maxint",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ROMAN_LOWER, Num: 2}, {Format: NUM_ARABIC, Num: 1}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 6}, {Format: NUM_ROMAN_LOWER, Num: 2}, {Format: NUM_ARABIC, Num: 1}}},
			want:  MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Diff(tt.other)
			if got != tt.want {
				t.Fatalf("Diff() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPageCmp(t *testing.T) {
	precedence := map[NumFormat]int{
		NUM_ROMAN_LOWER: 0,
		NUM_ARABIC:      1,
		NUM_ALPH_LOWER:  2,
		NUM_ROMAN_UPPER: 3,
		NUM_ALPH_UPPER:  4,
	}

	tests := []struct {
		name  string
		page  *Page
		other *Page
		want  int
	}{
		{
			name:  "equal pages compare equal",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 42}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 42}}},
			want:  0,
		},
		{
			name:  "precedence decides format order",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ROMAN_LOWER, Num: 1}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 1}}},
			want:  -1,
		},
		{
			name:  "number decides within same format",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 7}}},
			want:  -2,
		},
		{
			name:  "prefix equal shorter page is smaller",
			page:  &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}}},
			other: &Page{Numbers: []PageNumber{{Format: NUM_ARABIC, Num: 5}, {Format: NUM_ROMAN_LOWER, Num: 1}}},
			want:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Cmp(tt.other, precedence)
			if got != tt.want {
				t.Fatalf("Cmp() = %d, want %d", got, tt.want)
			}
		})
	}
}
