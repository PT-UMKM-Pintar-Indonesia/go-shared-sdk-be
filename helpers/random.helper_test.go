package sdk_helper

import (
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestRandomAlphaCharacters(t *testing.T) {
	r := &random{}

	got := r.AlphaCharacters(20)
	if len(got) != 20 {
		t.Fatalf("expected 20 chars, got %d", len(got))
	}

	for _, c := range got {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			t.Fatalf("non alpha char %q in %q", c, got)
		}
	}

	if r.AlphaCharacters(0) != "" {
		t.Fatal("expected empty for length 0")
	}
}

func TestRandomNumeric(t *testing.T) {
	r := &random{}

	for i := 0; i < 50; i++ {
		got := r.Numeric(8)
		if len(got) != 8 {
			t.Fatalf("expected 8 chars, got %q", got)
		}

		if got[0] == '0' {
			t.Fatalf("first digit must be non-zero, got %q", got)
		}

		if _, err := strconv.Atoi(got); err != nil {
			t.Fatalf("non numeric output %q: %v", got, err)
		}
	}

	if r.Numeric(-1) != "" {
		t.Fatal("expected empty for negative length")
	}

	if len(r.Numeric(1)) != 1 {
		t.Fatal("expected single digit")
	}
}

func TestRandomAlphanumeric(t *testing.T) {
	r := &random{}

	got, err := r.Alphanumeric(12)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 12 {
		t.Fatalf("expected 12 chars got %q", got)
	}

	if _, err := r.Alphanumeric(0); err == nil {
		t.Fatal("expected error for zero length")
	}
}

// A10 regression: charsets longer than 128 entries were silently truncated
// by the old 6/7-bit mask; Secure must now honour the full charset.
func TestRandomSecureLargeCharset(t *testing.T) {
	r := &random{}

	charset := strings.Join([]string{
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"0123456789",
		"!@#$%^&*()-_=+[]{};:,.<>?/~`",
		"ÄÖÜßåæøπΣΩ",
		"0123456789abcdefABCDEF",
	}, "")

	if len(charset) <= 128 {
		t.Fatalf("test charset too small: %d", len(charset))
	}

	allowed := map[byte]bool{}
	for i := 0; i < len(charset); i++ {
		allowed[charset[i]] = true
	}

	for i := 0; i < 20; i++ {
		got, err := r.Secure(64, charset)
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 64 {
			t.Fatalf("expected 64 bytes, got %d", len(got))
		}

		for j := 0; j < len(got); j++ {
			if !allowed[got[j]] {
				t.Fatalf("byte %d not in charset", got[j])
			}
		}
	}

	if _, err := r.Secure(4, ""); err == nil {
		t.Fatal("expected error for empty charset")
	}
}

func TestRandomHex(t *testing.T) {
	r := &random{}

	for _, n := range []int{1, 2, 5, 16, 33} {
		got, err := r.Hex(n)
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != n {
			t.Fatalf("Hex(%d) length = %d", n, len(got))
		}

		for _, c := range got {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Fatalf("non hex char %q", c)
			}
		}
	}

	if _, err := r.Hex(0); err == nil {
		t.Fatal("expected error for zero length")
	}
}

func TestRandomNumericStrAndItem(t *testing.T) {
	got := RandomNumericStr(10)
	if len(got) != 10 {
		t.Fatalf("expected 10 digits got %q", got)
	}

	if _, err := strconv.Atoi(got); err != nil {
		t.Fatalf("non numeric %q", got)
	}

	if RandomItem([]string{}) != "" {
		t.Fatal("expected zero value for empty slice")
	}

	items := []int{7}
	if RandomItem(items) != 7 {
		t.Fatal("expected the only element")
	}

	pool := []string{"a", "b", "c"}
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[RandomItem(pool)] = true
	}

	if len(seen) != 3 {
		t.Fatalf("expected all items to be reachable, saw %v", seen)
	}
}

func TestRandomConcurrent(t *testing.T) {
	r := &random{}

	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if len(r.AlphaCharacters(16)) != 16 {
				t.Error("unexpected length under concurrency")
			}
			if _, err := r.Alphanumeric(8); err != nil {
				t.Error(err)
			}
		}()
	}

	wg.Wait()
}
