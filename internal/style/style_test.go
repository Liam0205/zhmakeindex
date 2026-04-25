package style

import (
	"strings"
	"testing"
)

func TestScanStyleTokens(t *testing.T) {
	data := []byte(" \t'hello'world")
	advance, token, err := ScanStyleTokens(data, false)
	if err != nil || advance != 9 || string(token) != "'hello'" {
		t.Error(err, string(token), advance)
	}
	advance0, token0, err0 := ScanStyleTokens(data[advance:], false)
	if err0 != nil || advance0 != 0 || token0 != nil {
		t.Error(err0, string(token0), advance0)
	}
	advance1, token1, err1 := ScanStyleTokens(data[advance:], true)
	if err1 != nil || advance1 != 5 || string(token1) != "world" {
		t.Error(err1, string(token1), advance1)
	}
}

func TestScanStyleTokens_comment(t *testing.T) {
	data := []byte("%comment\nfoo bar")
	adv, tok, err := ScanStyleTokens(data, false)
	if err != nil || string(tok) != "foo" || adv != 12 {
		t.Error(err, string(tok), adv)
	}

	data0 := []byte("\n%c\n\n%c\nfoo%c\nbar")
	adv0, tok0, err0 := ScanStyleTokens(data0, false)
	if err0 != nil || string(tok0) != "foo" || adv0 != 11 {
		t.Error(err0, string(tok0), adv0)
	}

	adv1, tok1, err1 := ScanStyleTokens(data0[adv0:], true)
	if err1 != nil || string(tok1) != "bar" || adv1 != 6 {
		t.Error(err1, string(tok1), adv1)
	}
}

func TestScanStyleTokensBacktick(t *testing.T) {
	data := []byte("`hello` world")
	advance, token, err := ScanStyleTokens(data, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(token) != "`hello`" {
		t.Fatalf("first token = %q, want %q", string(token), "`hello`")
	}
	if advance != len("`hello`") {
		t.Fatalf("advance = %d, want %d", advance, len("`hello`"))
	}
}

func TestUnquote(t *testing.T) {
	if got := Unquote(`"hello"`); got != "hello" {
		t.Fatalf("Unquote(\"\\\"hello\\\"\") = %q, want %q", got, "hello")
	}
	if got := Unquote("`hello`"); got != "hello" {
		t.Fatalf("Unquote(%q) = %q, want %q", "`hello`", got, "hello")
	}
}

func TestUnquoteEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unquote(\"\") panicked: %v", r)
		}
	}()
	got := Unquote("")
	if got != "" {
		t.Fatalf("Unquote(\"\") = %q, want %q", got, "")
	}
}

func TestParseInt(t *testing.T) {
	if got := ParseInt("42"); got != 42 {
		t.Fatalf("ParseInt(%q) = %d, want %d", "42", got, 42)
	}
	if got := ParseInt("0"); got != 0 {
		t.Fatalf("ParseInt(%q) = %d, want %d", "0", got, 0)
	}
}

func TestNewInputStyle(t *testing.T) {
	s := NewInputStyle()
	if s.Keyword != "\\indexentry" {
		t.Fatalf("Keyword = %q, want %q", s.Keyword, "\\indexentry")
	}
	if s.ArgOpen != '{' {
		t.Fatalf("ArgOpen = %q, want %q", s.ArgOpen, '{')
	}
	if s.Level != '!' {
		t.Fatalf("Level = %q, want %q", s.Level, '!')
	}
	if s.Comment != '%' {
		t.Fatalf("Comment = %q, want %q", s.Comment, '%')
	}
}

func TestNewOutputStyle(t *testing.T) {
	s := NewOutputStyle()
	if !strings.Contains(s.Preamble, "\\begin{theindex}") {
		t.Fatalf("Preamble = %q, want to contain %q", s.Preamble, "\\begin{theindex}")
	}
	if s.DelimR != "--" {
		t.Fatalf("DelimR = %q, want %q", s.DelimR, "--")
	}
	if s.PagePrecedence != "rnaRA" {
		t.Fatalf("PagePrecedence = %q, want %q", s.PagePrecedence, "rnaRA")
	}
	if s.HeadingsFlag != 0 {
		t.Fatalf("HeadingsFlag = %d, want %d", s.HeadingsFlag, 0)
	}
}
