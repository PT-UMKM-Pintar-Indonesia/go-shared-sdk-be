package sdk_helper

import (
	"strings"
	"testing"
)

func TestStringCasing(t *testing.T) {
	s := NewString()

	if got := s.Capitalize("hELLO wORLD"); got != "Hello world" {
		t.Fatalf("Capitalize got %q", got)
	}

	if got := s.Capitalize(""); got != "" {
		t.Fatalf("Capitalize empty got %q", got)
	}

	if got := s.CamelCase("hello world-foo_bar baz"); got != "helloWorldFooBarBaz" {
		t.Fatalf("CamelCase got %q", got)
	}

	if got := s.CamelCase(""); got != "" {
		t.Fatalf("CamelCase empty got %q", got)
	}

	if got := s.SnakeCase("HelloWorld"); got != "hello_world" {
		t.Fatalf("SnakeCase got %q", got)
	}

	if got := s.SnakeCase("hello world"); got != "hello_world" {
		t.Fatalf("SnakeCase spaces got %q", got)
	}

	if got := s.SnakeCase("HTTPServer2Go"); got != "http_server2_go" {
		t.Fatalf("SnakeCase acronym got %q", got)
	}
}

func TestStringBasics(t *testing.T) {
	s := NewString()

	if s.Len("héllo") != 5 {
		t.Fatal("Len must count runes")
	}

	if s.CharAt("abc", 1) != "b" || s.CharAt("abc", 5) != "" || s.CharAt("abc", -1) != "" {
		t.Fatal("CharAt")
	}

	if r, ok := s.CharCodeAt("abc", 1); !ok || r != 'b' {
		t.Fatalf("CharCodeAt %v %v", r, ok)
	}

	if _, ok := s.CharCodeAt("abc", 9); ok {
		t.Fatal("CharCodeAt out of range")
	}

	if r, ok := s.CodePointAt("ab", 0); !ok || r != 'a' {
		t.Fatal("CodePointAt")
	}

	if got := s.Concat("a", "b", "c"); got != "abc" {
		t.Fatalf("Concat got %q", got)
	}

	if s.Concat() != "" || s.Concat("only") != "only" {
		t.Fatal("Concat edge cases")
	}

	if got := s.ToUpperCase("aB"); got != "AB" || s.ToLowerCase("Ab") != "ab" {
		t.Fatal("case conversion")
	}

	if !s.IsWellFormed("ok") || s.IsWellFormed("\xff\xfe") {
		t.Fatal("IsWellFormed")
	}

	if got := s.ToWellFormed("\xffok"); !s.IsWellFormed(got) {
		t.Fatal("ToWellFormed must produce valid utf8")
	}
}

func TestStringIndexSlice(t *testing.T) {
	s := NewString()

	if got := s.IndexOf("abcdef", 2); got != "c" {
		t.Fatalf("IndexOf got %q", got)
	}

	if got := s.IndexOf("abcdef", -1); got != "f" {
		t.Fatalf("negative IndexOf got %q", got)
	}

	if s.Slice("abcdef", 1, 3) != "bc" {
		t.Fatal("Slice")
	}

	if s.Slice("abcdef", -2, -1) != "e" {
		t.Fatal("negative Slice")
	}

	if s.Slice("abcdef", 4, 2) != "" {
		t.Fatal("inverted Slice must be empty")
	}

	if s.Substr("abcdef", 2, 100) != "cdef" {
		t.Fatal("Substr clamps length")
	}

	if s.Substr("abcdef", -4, 2) != "cd" {
		t.Fatalf("negative start Substr got %q", s.Substr("abcdef", -4, 2))
	}

	if s.Substr("abcdef", 1, -1) != "" {
		t.Fatal("negative length Substr must be empty")
	}

	if s.Substring("abcdef", 2, 4) != "cd" {
		t.Fatal("Substring")
	}

	if s.Substring("abcdef", 5, 2) != "" {
		t.Fatal("inverted Substring")
	}

	if idx := s.IndexOfStr("hello", "ll"); idx != 2 {
		t.Fatalf("IndexOfStr got %d", idx)
	}

	if pos, err := s.Search("hello world", "w.rld"); err != nil || pos != 6 {
		t.Fatalf("Search got %d %v", pos, err)
	}

	if _, err := s.Search("x", "[unclosed"); err == nil {
		t.Fatal("expected regex compile error")
	}
}

func TestStringPadTrimSplit(t *testing.T) {
	s := NewString()

	if got := s.PadStart("7", 3, "0"); got != "007" {
		t.Fatalf("PadStart got %q", got)
	}

	if got := s.PadEnd("7", 3, "0"); got != "700" {
		t.Fatalf("PadEnd got %q", got)
	}

	if got := s.PadStart("already-long", 3, "x"); got != "already-long" {
		t.Fatalf("PadStart no-op got %q", got)
	}

	if s.TrimStart("  x  ") != "x  " || s.TrimEnd("  x  ") != "  x" {
		t.Fatal("TrimStart/End")
	}

	if s.Trim(" x ") != "x" {
		t.Fatal("Trim")
	}

	if got := s.Split("a,b,c", ","); len(got) != 3 || got[1] != "b" {
		t.Fatalf("Split got %v", got)
	}

	if s.Reverse("abc") != "cba" {
		t.Fatal("Reverse")
	}

	if s.Repeat("ab", 3) != "ababab" {
		t.Fatal("Repeat")
	}

	if s.Replace("aaa", "a", "b") != "baa" {
		t.Fatal("Replace must replace first only")
	}

	if s.ReplaceAll("aaa", "a", "b") != "bbb" {
		t.Fatal("ReplaceAll")
	}

	if !s.StartsWith("abc", "ab") || !s.EndsWith("abc", "bc") {
		t.Fatal("StartsWith/EndsWith")
	}
}

func TestStringCollectionOps(t *testing.T) {
	s := NewString()

	if got := s.Truncate("hello world", 5); got != "hello" {
		t.Fatalf("Truncate got %q", got)
	}

	if got := s.Truncate("hi", 5); got != "hi" {
		t.Fatal("Truncate no-op")
	}

	if got := s.Compact(" a b\tc "); got != "abc" {
		t.Fatalf("Compact got %q", got)
	}

	if got := s.Uniq("aabbcc"); got != "abc" {
		t.Fatalf("Uniq got %q", got)
	}

	if got := s.Unshift("core", "pre1-", "pre2-"); got != "pre1-pre2-core" {
		t.Fatalf("Unshift got %q", got)
	}

	if s.Unshift("core") != "core" {
		t.Fatal("Unshift no prefixes")
	}

	last, rest, ok := s.Pop("abc")
	if !ok || last != "c" || rest != "ab" {
		t.Fatalf("Pop got %q %q %v", last, rest, ok)
	}

	if _, _, ok := s.Pop(""); ok {
		t.Fatal("Pop empty must report false")
	}

	if !s.Included("hello", "ell") || s.Included("hello", "zz") {
		t.Fatal("Included")
	}

	if !s.IncludedIgnoreCase("Hello", "hEL") {
		t.Fatal("IncludedIgnoreCase")
	}

	if got := s.Union("aab", "bbc", "cd"); got != "abcd" {
		t.Fatalf("Union got %q", got)
	}

	if s.Union() != "" {
		t.Fatal("Union empty")
	}

	// Multi-byte safety on truncation paths.
	if got := s.Truncate("ééé", 2); got != "éé" || !strings.Contains(s.Uniq("éé"), "é") {
		t.Fatal("rune handling")
	}
}
