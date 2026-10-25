package sdk_pkg

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	return key
}

func TestJweRoundTrip(t *testing.T) {
	key := testRSAKey(t)
	p := NewJose(context.Background())

	const secret = "super-sensitive-123"

	cipherText, meta, err := p.JweEncrypt(&key.PublicKey, secret)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if len(cipherText) == 0 || meta == nil {
		t.Fatal("empty cipher material")
	}

	plain, err := p.JweDecrypt(key, cipherText)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if plain != secret {
		t.Fatalf("round trip mismatch: %q", plain)
	}
}

// B3: the JWE protected header used to carry the raw plaintext in "sig".
// It must now contain only a SHA-256 digest and never leak the secret.
func TestJweHeaderSigIsHashed(t *testing.T) {
	key := testRSAKey(t)
	p := NewJose(context.Background())

	const secret = "leak-me-not"

	_, meta, err := p.JweEncrypt(&key.PublicKey, secret)
	if err != nil {
		t.Fatal(err)
	}

	rawHeader, err := base64URLDecode(meta.Protected)
	if err != nil {
		t.Fatalf("protected header is not base64url: %v", err)
	}

	var header map[string]any
	if err := json.Unmarshal(rawHeader, &header); err != nil {
		t.Fatalf("protected header is not json: %v", err)
	}

	sig, ok := header["sig"].(string)
	if !ok {
		t.Fatalf("missing sig header, got %v", header)
	}

	if sig == secret {
		t.Fatal("plaintext leaked into protected header")
	}

	sum := sha256.Sum256([]byte(secret))
	if want := hex.EncodeToString(sum[:]); sig != want {
		t.Fatalf("sig header is not the SHA-256 digest: got %q want %q", sig, want)
	}

	if len(sig) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(sig))
	}

	if !strings.Contains(string(rawHeader), "RSA-OAEP-512") {
		t.Fatalf("alg header missing: %s", rawHeader)
	}
}

func TestJweDecryptWrongKey(t *testing.T) {
	key := testRSAKey(t)
	other := testRSAKey(t)
	p := NewJose(context.Background())

	cipherText, _, err := p.JweEncrypt(&key.PublicKey, "data")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.JweDecrypt(other, cipherText); err == nil {
		t.Fatal("expected decryption failure with wrong key")
	}
}

func TestJwtSignProducesVerifiableToken(t *testing.T) {
	key := testRSAKey(t)
	p := NewJose(context.Background())

	now := time.Now().UTC()

	token, err := p.JwtSign(&sdk_dto.JwtSignOption{
		PrivateKey: key,
		SecretKey:  "sec",
		Kid:        "kid-42",
		Aud:        []string{"aud-1"},
		Iss:        "iss-1",
		Sub:        "sub-1",
		Jti:        "jti-1",
		Iat:        now,
		Exp:        now.Add(time.Hour),
		Claim:      "timestamp-value",
	})
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	parts := strings.Split(strings.TrimSpace(string(token)), ".")
	if len(parts) != 3 {
		t.Fatalf("expected compact JWS, got %d parts", len(parts))
	}

	payload, err := jws.Verify(token, jws.WithKey(jwa.RS512(), &key.PublicKey))
	if err != nil {
		t.Fatalf("signature verification failed: %v", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("claims not json: %v", err)
	}

	if claims["iss"] != "iss-1" || claims["sub"] != "sub-1" || claims["jti"] != "jti-1" {
		t.Fatalf("unexpected claims %+v", claims)
	}

	if claims["timestamp"] != "timestamp-value" {
		t.Fatalf("custom claim missing: %+v", claims)
	}

	parsed, _ := jws.Parse(token)
	hdrs := parsed.Signatures()[0].ProtectedHeaders()

	if kid, ok := hdrs.KeyID(); !ok || kid != "kid-42" {
		t.Fatalf("kid header missing or wrong: %v", kid)
	}
}

func TestJwtSignInvalidOptions(t *testing.T) {
	p := NewJose(context.Background())

	// Expired window (exp before iat) must surface an error, not panic.
	if _, err := p.JwtSign(&sdk_dto.JwtSignOption{
		PrivateKey: nil,
		Iat:        time.Now(),
		Exp:        time.Now().Add(time.Hour),
	}); err == nil {
		t.Fatal("expected error for nil private key")
	}
}

func TestJwkImportExport(t *testing.T) {
	key := testRSAKey(t)
	p := NewJose(context.Background())

	jwkKey, err := jwk.Import(key)
	if err != nil {
		t.Fatal(err)
	}

	meta, err := p.ImportJsonWebKey(jwkKey)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if meta == nil || meta.Key == nil {
		t.Fatal("empty import metadata")
	}

	if meta.KeyRaw.Kty != "RSA" {
		t.Fatalf("unexpected kty %q", meta.KeyRaw.Kty)
	}

	if expMeta, err := p.ExportJsonWebKey(key); err == nil && expMeta != nil {
		if expMeta.KeyRaw.Kty != "RSA" {
			t.Fatalf("export kty %q", expMeta.KeyRaw.Kty)
		}
	}
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
