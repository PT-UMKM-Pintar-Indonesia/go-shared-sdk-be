package sdk_helper

import (
	"regexp"
	"testing"
)

var codeRe = regexp.MustCompile(`^(UMKMS|OTC)\d{4}[A-Z2-9]{2}$`)

// B5 regression: codes used to be date-derived and identical for the whole
// day; generated codes must now vary.
func TestOTCGenerateFormatAndValidation(t *testing.T) {
	seen := map[string]bool{}

	for i := 0; i < 100; i++ {
		code := Generate("UMKMS")

		if !codeRe.MatchString(code) {
			t.Fatalf("generated code %q does not match expected format", code)
		}

		if !Validate(code) {
			t.Fatalf("Validate rejected freshly generated code %q", code)
		}

		seen[code] = true
	}

	if len(seen) < 2 {
		t.Fatalf("expected unique codes across generations, got only %d distinct", len(seen))
	}
}

func TestOTCGenerateUnknownPrefixFallsBack(t *testing.T) {
	code := Generate("BOGUS")

	if !regexp.MustCompile(`^OTC`).MatchString(code) {
		t.Fatalf("expected OTC fallback prefix, got %q", code)
	}

	if !Validate(code) {
		t.Fatalf("fallback code failed validation: %q", code)
	}
}

func TestOTCValidateRejectsBadCodes(t *testing.T) {
	tests := []string{
		"",
		"OTC",
		"XXXX1234AB",
		"OTC123",
		"OTC1234ab", // lowercase not allowed
		"OTC1234!!",
	}

	for _, tc := range tests {
		if Validate(tc) {
			t.Fatalf("expected %q to be rejected", tc)
		}
	}

	// A well-formed code with a wrong checksum must be rejected.
	base := "OTC1234"
	good := generateChecksum(base)
	bad := "ZZ"
	if good == bad {
		bad = "YY"
	}

	if Validate(base + bad) {
		t.Fatal("expected wrong checksum to be rejected")
	}

	if !Validate(base + good) {
		t.Fatalf("expected deterministic checksum %q to validate", good)
	}
}

func TestOTCChecksumDeterministic(t *testing.T) {
	inputs := []string{"UMKMS0000", "OTC9999", "hello", ""}

	for _, in := range inputs {
		if generateChecksum(in) != generateChecksum(in) {
			t.Fatalf("checksum for %q is not deterministic", in)
		}
	}
}
