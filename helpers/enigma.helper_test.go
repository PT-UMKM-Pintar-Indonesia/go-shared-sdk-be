package sdk_helper

import (
	"strings"
	"sync"
	"testing"
)

// A1 regression: the rotor III wiring used to contain a space, which made
// createRotor index backward[-29] and panic on the first NewEnigma() call.
func TestNewEnigmaDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewEnigma panicked: %v", r)
		}
	}()

	e := NewEnigma()
	if e == nil {
		t.Fatal("NewEnigma returned nil")
	}
}

func TestEnigmaEncodeDecodeRoundTrip(t *testing.T) {
	e := NewEnigma()

	const plain = "HELLOWORLD"
	cipher := e.Encode(plain)

	if cipher == plain {
		t.Fatal("expected ciphertext to differ from plaintext")
	}

	if len(cipher) != len(plain) {
		t.Fatalf("expected same length, got %d vs %d", len(cipher), len(plain))
	}

	for _, r := range cipher {
		if r < 'A' || r > 'Z' {
			t.Fatalf("ciphertext contains non-uppercase letter: %q", cipher)
		}
	}

	// Decode resets the rotor positions, so it must invert a fresh Encode.
	if got := e.Decode(cipher); got != plain {
		t.Fatalf("round trip mismatch: got %q want %q", got, plain)
	}
}

func TestEnigmaEncodeFromFreshMachine(t *testing.T) {
	// Two independent machines with identical settings must produce the
	// same ciphertext for the same input.
	a := NewEnigma().Encode("ATTACKATDAWN")
	b := NewEnigma().Encode("ATTACKATDAWN")

	if a != b {
		t.Fatalf("deterministic encoding mismatch: %q vs %q", a, b)
	}
}

func TestEnigmaStripsSpacesAndUppercases(t *testing.T) {
	e := NewEnigma()

	got := e.Encode("hello world")
	if strings.ContainsAny(got, " ") {
		t.Fatalf("spaces must be stripped, got %q", got)
	}

	if len(got) != len("helloworld") {
		t.Fatalf("expected %d chars, got %d (%q)", len("helloworld"), len(got), got)
	}

	// Non-letters are dropped by process(), digits included.
	if withDigits := e.Decode("ABC123"); len(withDigits) != 3 {
		t.Fatalf("expected non-letters to be skipped, got %q", withDigits)
	}
}

func TestEnigmaEmptyInput(t *testing.T) {
	e := NewEnigma()

	if got := e.Encode(""); got != "" {
		t.Fatalf("expected empty output for empty input, got %q", got)
	}
}

func TestEnigmaRotorsAreIndependentPerInstance(t *testing.T) {
	e := NewEnigma()

	first := e.Encode("AAAAA")
	second := e.Encode("AAAAA")

	// The rotors advance between Encode calls on the same machine, so the
	// outputs must differ (classic Enigma property).
	if first == second {
		t.Fatalf("expected rotor stepping to change output, both were %q", first)
	}
}

func TestEnigmaConcurrentInstances(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := NewEnigma()
			c := e.Encode("RACETEST")
			if p := e.Decode(c); p != "RACETEST" {
				t.Errorf("round trip failed under concurrency: %q", p)
			}
		}()
	}

	wg.Wait()
}
