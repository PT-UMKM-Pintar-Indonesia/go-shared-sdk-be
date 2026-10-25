package sdk_helper

import (
	"strings"
	"testing"
)

func TestTrimLogStringPassthrough(t *testing.T) {
	s := "short log line"

	if got := TrimLogString(s); got != s {
		t.Fatalf("short string must be untouched, got %q", got)
	}
}

func TestTrimLogStringTruncates(t *testing.T) {
	long := strings.Repeat("a", DefaultMaxLogLength+100)

	got := TrimLogString(long)

	if len(got) > DefaultMaxLogLength {
		t.Fatalf("expected at most %d bytes, got %d", DefaultMaxLogLength, len(got))
	}

	if !strings.HasSuffix(got, TruncatedSuffix) {
		t.Fatalf("expected truncation suffix, got %q", got[len(got)-20:])
	}

	if len(got) != DefaultMaxLogLength {
		t.Fatalf("expected exact max length, got %d", len(got))
	}
}

func TestTrimLogStringRuneSafe(t *testing.T) {
	// Multi-byte content: truncation must not split runes.
	s := strings.Repeat("é", 1000) // 2 bytes per rune

	got := TrimLogStringWithMax(s, 100)

	if !strings.HasSuffix(got, TruncatedSuffix) {
		t.Fatal("expected suffix")
	}

	runes := []rune(got)
	if len(runes) != 100 {
		t.Fatalf("expected 100 runes, got %d", len(runes))
	}
}

func TestTrimLogStringEdgeCases(t *testing.T) {
	if got := TrimLogStringWithMax("abc", 0); got != "abc" {
		t.Fatalf("maxLen<=0 must fall back to default, got %q", got)
	}

	if got := TrimLogStringWithMax("abc", -5); got != "abc" {
		t.Fatalf("negative maxLen must fall back, got %q", got)
	}

	// maxLen smaller than suffix length must not panic.
	got := TrimLogStringWithMax(strings.Repeat("x", 50), 5)
	if got != TruncatedSuffix {
		t.Fatalf("expected just the suffix for tiny maxLen, got %q", got)
	}
}

func TestTrimLogBytes(t *testing.T) {
	got := TrimLogBytes([]byte(strings.Repeat("b", 2000)), 10)

	// maxLen smaller than the suffix length collapses to just the suffix.
	if got != TruncatedSuffix {
		t.Fatalf("expected bare suffix, got %q", got)
	}

	got = TrimLogBytes([]byte(strings.Repeat("b", 2000)), 100)
	if len([]rune(got)) != 100 || !strings.HasSuffix(got, TruncatedSuffix) {
		t.Fatalf("unexpected %q", got)
	}
}

func TestSanitizeLogString(t *testing.T) {
	got := SanitizeLogString("line1\nline2\r\nline3   spaced", 1024)

	if strings.ContainsAny(got, "\n\r") {
		t.Fatalf("newlines must be removed, got %q", got)
	}

	if strings.Contains(got, "  ") {
		t.Fatalf("whitespace must be collapsed, got %q", got)
	}

	if got != "line1 line2 line3 spaced" {
		t.Fatalf("unexpected %q", got)
	}

	long := SanitizeLogString(strings.Repeat("word ", 500), 50)
	if len([]rune(long)) != 50 {
		t.Fatalf("expected truncation to 50 runes, got %d", len([]rune(long)))
	}
}
