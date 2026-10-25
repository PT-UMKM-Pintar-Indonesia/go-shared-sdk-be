package sdk_helper

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBase64RoundTrip(t *testing.T) {
	c := NewCipher()

	enc := c.Base64Encode("hello world")
	if enc != base64.StdEncoding.EncodeToString([]byte("hello world")) {
		t.Fatalf("unexpected encoding %q", enc)
	}

	dec, err := c.Base64Decode(enc)
	if err != nil || dec != "hello world" {
		t.Fatalf("round trip: %q %v", dec, err)
	}

	if _, err := c.Base64Decode("!!!not-base64!!!"); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestCaesar(t *testing.T) {
	c := NewCipher()

	if got := c.CaesarEncrypt("abcXYZ", 3); got != "defABC" {
		t.Fatalf("encrypt got %q", got)
	}

	if got := c.CaesarDecrypt("defABC", 3); got != "abcXYZ" {
		t.Fatalf("decrypt got %q", got)
	}

	// Rotation of 26 is identity.
	if got := c.CaesarEncrypt("keep", 26); got != "keep" {
		t.Fatalf("identity rotation got %q", got)
	}

	// Non letters untouched.
	if got := c.CaesarEncrypt("a1b!", 1); got != "b1c!" {
		t.Fatalf("got %q", got)
	}

	// Negative rotation behaves like decryption.
	if got := c.CaesarEncrypt("bcd", -1); got != "abc" {
		t.Fatalf("negative rotation got %q", got)
	}
}

func TestRotateNumber(t *testing.T) {
	c := NewCipher()

	if got := c.RotateNumber("12345", 3); got != "45678" {
		t.Fatalf("rotate got %q", got)
	}

	if got := c.RotateNumber("999", 1); got != "000" {
		t.Fatalf("wrap got %q", got)
	}

	if got := c.DerotateNumber("45678", 3); got != "12345" {
		t.Fatalf("derotate got %q", got)
	}

	if got := c.RotateNumber("a1b2", 5); got != "a6b7" {
		t.Fatalf("non digits must pass through, got %q", got)
	}

	round := c.DerotateNumber(c.RotateNumber("0123456789", 7), 7)
	if round != "0123456789" {
		t.Fatalf("round trip got %q", round)
	}
}

func TestEncodeDecodeRotation(t *testing.T) {
	c := NewCipher()

	secret := "sensitive-payload"
	encoded := c.EncodeRotation(secret)

	if strings.Contains(encoded, "sensitive") {
		t.Fatal("plaintext must not be visible in encoded output")
	}

	decoded, err := c.DecodeRotation(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !strings.HasPrefix(decoded, secret+":") {
		t.Fatalf("decoded payload mismatch: %q", decoded)
	}

	if _, err := c.DecodeRotation(""); err == nil {
		t.Fatal("expected error for empty cipher text")
	}
}
