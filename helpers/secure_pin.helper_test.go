package sdk_helper

import (
	"strings"
	"testing"
)

func TestSecurePinWeaknessDetection(t *testing.T) {
	weak := []string{"123456", "000000", "111111", "12345", "821234"}

	for _, pin := range weak {
		ok, err := SecurePin(pin)
		if err != nil {
			t.Fatalf("unexpected validation error for %q: %v", pin, err)
		}

		if ok {
			t.Errorf("expected %q to be flagged as weak", pin)
		}
	}

	strong, err := SecurePin("827361")
	if err != nil {
		t.Fatal(err)
	}

	if !strong {
		t.Fatal("expected 827361 to pass all checks")
	}
}

func TestValidatePin(t *testing.T) {
	if err := ValidatePin("1234"); err != nil {
		t.Fatalf("4 digit pin should pass: %v", err)
	}

	if err := ValidatePin("12"); err == nil {
		t.Fatal("too short")
	}

	if err := ValidatePin("12345678"); err == nil {
		t.Fatal("too long")
	}

	if err := ValidatePin("abcd"); err == nil {
		t.Fatal("must require digits")
	}

	if err := ValidatePinWithOptions("1234", 4, 6, true); err == nil {
		t.Fatal("requireAlpha should reject digits-only")
	}

	if err := ValidatePinWithOptions("12a4", 4, 6, true); err != nil {
		t.Fatalf("mixed should pass: %v", err)
	}
}

func TestHasSequentialOrRepeated(t *testing.T) {
	yes := []string{"123", "321", "111", "9876", "a1234zz"}
	no := []string{"12", "11", "1357", "1243"}

	for _, s := range yes {
		if !hasSequentialOrRepeated(s) {
			t.Errorf("expected %q flagged", s)
		}
	}

	for _, s := range no {
		if hasSequentialOrRepeated(s) {
			t.Errorf("expected %q clean", s)
		}
	}
}

func TestGenerateSecurePIN(t *testing.T) {
	if _, err := GenerateSecurePIN(3); err == nil {
		t.Fatal("expected error below min length")
	}

	if _, err := GenerateSecurePIN(13); err == nil {
		t.Fatal("expected error above max length")
	}

	seen := map[string]bool{}

	for i := 0; i < 20; i++ {
		pin, err := GenerateSecurePIN(8)
		if err != nil {
			t.Fatal(err)
		}

		if len(pin) != 8 {
			t.Fatalf("expected 8 digits, got %q", pin)
		}

		for _, c := range pin {
			if c < '0' || c > '9' {
				t.Fatalf("non digit in generated pin %q", pin)
			}
		}

		seen[pin] = true
	}

	if len(seen) < 15 {
		t.Fatalf("expected randomness, only %d distinct pins", len(seen))
	}
}

func TestCleanPin(t *testing.T) {
	if got := CleanPin("a1-b2\tc3"); got != "123" {
		t.Fatalf("got %q", got)
	}

	if got := CleanPin(strings.Repeat("x", 100)); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
