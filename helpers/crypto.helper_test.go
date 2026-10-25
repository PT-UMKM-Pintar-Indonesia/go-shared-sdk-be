package sdk_helper

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestAES256RoundTrip(t *testing.T) {
	c := &cryptoHelper{}

	const key = "this-is-a-secret-key-of-32-chars!"
	const plain = "hello, sdk!"

	enc, err := c.AES256Encrypt(key, plain)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	dec, err := c.AES256Decrypt(key, enc)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if dec != plain {
		t.Fatalf("round trip mismatch: got %q want %q", dec, plain)
	}

	// Random salt/nonce: two encryptions of the same text must differ.
	enc2, _ := c.AES256Encrypt(key, plain)
	if enc == enc2 {
		t.Fatal("expected randomized salt/nonce to change ciphertext")
	}
}

// A7 regression: aes.NewCipher / NewGCM errors used to be discarded,
// producing panics or garbage instead of errors.
func TestAES256EncryptShortKey(t *testing.T) {
	c := &cryptoHelper{}

	if _, err := c.AES256Encrypt("short", "data"); err == nil {
		t.Fatal("expected error for key shorter than 16 chars")
	}
}

func TestAES256DecryptMalformed(t *testing.T) {
	c := &cryptoHelper{}

	tests := []struct {
		name string
		in   string
	}{
		{"not hex", "zzzz"},
		{"too short", hex.EncodeToString([]byte("tiny"))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := c.AES256Decrypt("0123456789abcdef0123", tc.in); err == nil {
				t.Fatal("expected error for malformed ciphertext")
			}
		})
	}
}

func TestAES256DecryptTampered(t *testing.T) {
	c := &cryptoHelper{}

	const key = "0123456789abcdef0123456789abcdef"

	enc, err := c.AES256Encrypt(key, "payload")
	if err != nil {
		t.Fatal(err)
	}

	raw, _ := hex.DecodeString(enc)
	raw[len(raw)-1] ^= 0xFF // flip a ciphertext byte -> GCM tag must fail

	if _, err := c.AES256Decrypt(key, hex.EncodeToString(raw)); err == nil {
		t.Fatal("expected authentication failure for tampered ciphertext")
	}

	// Wrong key must also fail cleanly (no panic).
	if _, err := c.AES256Decrypt("ffffffffffffffffffffffffffffffff", enc); err == nil {
		t.Fatal("expected decryption failure with wrong key")
	}
}

func TestAES256EncryptWithSalt(t *testing.T) {
	c := &cryptoHelper{}

	if _, err := c.AES256EncryptWithSalt("0123456789abcdef", "x", "not-hex"); err == nil {
		t.Fatal("expected error for invalid salt hex")
	}

	salt := hex.EncodeToString([]byte("0123456789abcdef"))

	out, err := c.AES256EncryptWithSalt("secret-secret-secret", "value", salt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) < 2*(12+16) {
		t.Fatalf("output too short: %q", out)
	}
}

func TestHMACSHA512(t *testing.T) {
	c := &cryptoHelper{}

	sig, err := c.HMACSHA512Sign("key", "data")
	if err != nil {
		t.Fatal(err)
	}

	if !c.HMACSHA512Verify("key", "data", sig) {
		t.Fatal("expected verification to succeed")
	}

	if c.HMACSHA512Verify("key", "other", sig) {
		t.Fatal("expected verification to fail for different data")
	}

	if c.HMACSHA512Verify("key", "data", "not-hex!!") {
		t.Fatal("expected verification to fail for non hex hash")
	}
}

func TestSHAKnownVectors(t *testing.T) {
	c := &cryptoHelper{}

	sha256abc, _ := c.SHA256Sign("abc")
	if want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"; sha256abc != want {
		t.Fatalf("SHA256(abc) = %s", sha256abc)
	}

	sha512abc, _ := c.SHA512Sign("abc")
	if !strings.HasPrefix(sha512abc, "ddaf35a193617abacc417349ae204131") {
		t.Fatalf("unexpected SHA512(abc) = %s", sha512abc)
	}

	if len(sha512abc) != 128 {
		t.Fatalf("expected 128 hex chars, got %d", len(sha512abc))
	}
}

func TestGenerateRandomBytesAndHex(t *testing.T) {
	c := &cryptoHelper{}

	if _, err := c.GenerateRandomBytes(0); err == nil {
		t.Fatal("expected error for non positive length")
	}

	b, err := c.GenerateRandomBytes(32)
	if err != nil || len(b) != 32 {
		t.Fatalf("unexpected: %v %v", b, err)
	}

	b2, _ := c.GenerateRandomBytes(32)
	if hex.EncodeToString(b) == hex.EncodeToString(b2) {
		t.Fatal("expected unique random bytes")
	}

	h, err := c.GenerateRandomHex(16)
	if err != nil || len(h) != 32 {
		t.Fatalf("unexpected hex: %q %v", h, err)
	}
}
