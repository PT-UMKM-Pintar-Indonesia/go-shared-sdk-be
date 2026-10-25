package sdk_helper

import (
	"strings"
	"sync"
	"testing"
)

// A2 regression: a customer ID with more than 18 digits used to write past
// the end of the pooled 18-byte buffer and panic.
func TestVAGenerateLongCustomerIDNoPanic(t *testing.T) {
	h := NewVaGenerator()

	longID := strings.Repeat("9", 500)

	va, err := h.Generate(BCA, longID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(va) != 15 { // BCA MaxLength
		t.Fatalf("expected length 15, got %d (%q)", len(va), va)
	}

	if !strings.HasPrefix(va, "11234") {
		t.Fatalf("expected BCA prefix, got %q", va)
	}

	// Only the last digits may be retained; the tail of the input is "9"s.
	if va != "11234"+"9999999999" {
		t.Fatalf("expected last-18-digits behaviour, got %q", va)
	}
}

func TestVAGenerateKeepsLastDigits(t *testing.T) {
	h := NewVaGenerator()

	// BCA: prefix 11234 (5) + 10 unique digits; input has 13 digits so the
	// last 10 must survive.
	va, err := h.Generate(BCA, "CUST-1234567890123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "112344567890123"; va != want {
		t.Fatalf("got %q want %q", va, want)
	}
}

func TestVAGeneratePadsShortCustomerID(t *testing.T) {
	h := NewVaGenerator()

	va, err := h.Generate(BCA, "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "11234" + "0000000" + "123"; va != want {
		t.Fatalf("got %q want %q", va, want)
	}
}

func TestVAGenerateErrors(t *testing.T) {
	h := NewVaGenerator()

	if _, err := h.Generate(BankType("UNKNOWN"), "123"); err == nil {
		t.Fatal("expected error for unsupported bank")
	}

	if _, err := h.Generate(BCA, "no-digits-here"); err == nil {
		t.Fatal("expected error for customer ID without digits")
	}
}

func TestVAGenerateExactlyBufferCapacity(t *testing.T) {
	h := NewVaGenerator()

	// 18 digits == buffer size (boundary), and 19 digits == overflow case.
	for _, id := range []string{strings.Repeat("7", 18), strings.Repeat("7", 19)} {
		va, err := h.Generate(BRI, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(va) != 18 { // BRI MaxLength
			t.Fatalf("expected length 18, got %d", len(va))
		}

		if !strings.HasPrefix(va, "12345") {
			t.Fatalf("expected BRI prefix, got %q", va)
		}
	}
}

func TestVAValidate(t *testing.T) {
	h := NewVaGenerator()

	tests := []struct {
		name    string
		bank    BankType
		va      string
		wantErr bool
	}{
		{"valid bca", BCA, "11234" + "0000000123", false},
		{"wrong length", BCA, "11234000000012", true},
		{"wrong prefix", BCA, "999990000000123", true},
		{"non numeric", BCA, "1123400000001AB", true},
		{"unknown bank", BankType("XX"), "123", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := h.Validate(tc.bank, tc.va)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate(%q) err = %v, wantErr %v", tc.va, err, tc.wantErr)
			}
		})
	}

	// Anything generated must validate.
	va, err := h.Generate(MANDIRI, "42")
	if err != nil {
		t.Fatal(err)
	}

	if err := h.Validate(MANDIRI, va); err != nil {
		t.Fatalf("generated VA failed validation: %v", err)
	}
}

func TestGenerateRandomVA(t *testing.T) {
	va := GenerateRandomVA("bca")

	if !strings.HasPrefix(va, "39012") {
		t.Fatalf("expected known BCA prefix, got %q", va)
	}

	body := va[:len(va)-1]
	last := int(va[len(va)-1] - '0')

	check, err := calculateCheckDigit(body)
	if err != nil {
		t.Fatalf("generated VA contains non digits: %v", err)
	}

	if check != last {
		t.Fatalf("Luhn check digit mismatch: expected %d got %d in %q", check, last, va)
	}

	// Unknown bank codes fall back to a random known prefix.
	if other := GenerateRandomVA("unknown-bank"); other == "" {
		t.Fatal("expected fallback prefix, got empty")
	}
}

func TestVAGenerateConcurrent(t *testing.T) {
	h := NewVaGenerator()

	var wg sync.WaitGroup

	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			va, err := h.Generate(BCA, strings.Repeat("8", 300))
			if err != nil || len(va) != 15 {
				t.Errorf("concurrent generate failed: %q %v", va, err)
			}
		}(i)
	}

	wg.Wait()
}
