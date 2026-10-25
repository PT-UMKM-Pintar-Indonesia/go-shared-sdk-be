package sdk_helper

import (
	"math"
	"testing"
)

func TestNumberToStringAndValueOf(t *testing.T) {
	if NumberToString(42) != "42" {
		t.Fatal("NumberToString")
	}

	if ValueOf(7) != 7 {
		t.Fatal("ValueOf")
	}
}

func TestNumericFormatters(t *testing.T) {
	if got := ToFixed(3.14159, 2); got != "3.14" {
		t.Fatalf("ToFixed got %q", got)
	}

	if got := ToExponential(1500, 1); got != "1.5e+03" {
		t.Fatalf("ToExponential got %q", got)
	}

	if got := ToPrecision(123.456, 4); got != "123.5" {
		t.Fatalf("ToPrecision got %q", got)
	}
}

func TestNumberPredicates(t *testing.T) {
	if !IsFinite(1.0) || IsFinite(math.Inf(1)) || IsFinite(math.NaN()) {
		t.Fatal("IsFinite")
	}

	if !IsInteger(4.0) || IsInteger(4.5) {
		t.Fatal("IsInteger")
	}

	if !IsNaN(math.NaN()) || IsNaN(0) {
		t.Fatal("IsNaN")
	}

	if !IsSafeInteger(9007199254740991) || IsSafeInteger(9007199254740992) {
		t.Fatal("IsSafeInteger bounds")
	}

	if IsSafeInteger(math.Inf(1)) {
		t.Fatal("IsSafeInteger must reject infinity")
	}
}

func TestParseIntFloat(t *testing.T) {
	n, err := ParseInt("  42 ")
	if err != nil || n != 42 {
		t.Fatalf("ParseInt: %d %v", n, err)
	}

	if _, err := ParseInt("abc"); err == nil {
		t.Fatal("expected ParseInt error")
	}

	f, err := ParseFloat(" 1.25")
	if err != nil || f != 1.25 {
		t.Fatalf("ParseFloat: %f %v", f, err)
	}

	if _, err := ParseFloat("x"); err == nil {
		t.Fatal("expected ParseFloat error")
	}
}

func TestMinMaxClamp(t *testing.T) {
	if Min(3, 5) != 3 || Max(3, 5) != 5 {
		t.Fatal("Min/Max")
	}

	if Clamp(10, 0, 5) != 5 || Clamp(-1, 0, 5) != 0 || Clamp(3, 0, 5) != 3 {
		t.Fatal("Clamp")
	}
}
