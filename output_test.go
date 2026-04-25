package main

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/text/encoding"

	"github.com/leo-liu/zhmakeindex/internal/index"
	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

func TestOutputFirstItemSubentry(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "output_test_*.ind")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	tmpfile.Close()
	defer os.Remove(name)

	outstyle := style.NewOutputStyle()
	option := &OutputOptions{
		output:  name,
		encoder: encoding.Nop.NewEncoder(),
	}

	out := &OutputIndex{
		groups: []index.IndexGroup{
			{
				Name: "T",
				Items: []index.IndexItem{
					{Level: 1, Text: "orphan-sub", Page: []index.PageRange{
						{Begin: &page.Page{
							Numbers:   []page.PageNumber{{Format: page.NUM_ARABIC, Num: 1}},
							Rangetype: page.PAGE_NORMAL,
						}, End: &page.Page{
							Numbers:   []page.PageNumber{{Format: page.NUM_ARABIC, Num: 1}},
							Rangetype: page.PAGE_NORMAL,
						}},
					}},
				},
			},
		},
		style:  outstyle,
		option: option,
	}

	out.Output(option)

	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "orphan-sub") {
		t.Fatalf("output missing subentry text, got: %s", data)
	}
}

func TestOutputFirstItemSubsubentry(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "output_test_*.ind")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	tmpfile.Close()
	defer os.Remove(name)

	outstyle := style.NewOutputStyle()
	option := &OutputOptions{
		output:  name,
		encoder: encoding.Nop.NewEncoder(),
	}

	out := &OutputIndex{
		groups: []index.IndexGroup{
			{
				Name: "T",
				Items: []index.IndexItem{
					{Level: 2, Text: "orphan-subsub", Page: []index.PageRange{
						{Begin: &page.Page{
							Numbers:   []page.PageNumber{{Format: page.NUM_ARABIC, Num: 5}},
							Rangetype: page.PAGE_NORMAL,
						}, End: &page.Page{
							Numbers:   []page.PageNumber{{Format: page.NUM_ARABIC, Num: 5}},
							Rangetype: page.PAGE_NORMAL,
						}},
					}},
				},
			},
		},
		style:  outstyle,
		option: option,
	}

	out.Output(option)

	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "orphan-subsub") {
		t.Fatalf("output missing subsubentry text, got: %s", data)
	}
}
