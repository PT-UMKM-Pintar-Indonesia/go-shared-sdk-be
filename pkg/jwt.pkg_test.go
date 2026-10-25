package sdk_pkg

import (
	"context"
	"strings"
	"testing"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_dro "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

func TestJwtSignNilRedisRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: generates a 4096-bit RSA key")
	}

	// B4 regression: the expiry suffix used to be computed with
	// int(math.Pow(EXPIRED, 30)) which overflows to garbage (and on some
	// platforms panics). Sign must now succeed deterministically.
	env := &sdk_dto.Environment{}
	env.JWT = sdk_dro.Jwt{SECRET: "unused", EXPIRED: 60}

	p := NewJsonWebToken(context.Background(), env, nil)

	res, err := p.Sign("client-a", map[string]any{"user": "42"})
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	if res == nil || res.Expired != 60 {
		t.Fatalf("unexpected metadata %+v", res)
	}

	if len(strings.Split(strings.TrimSpace(res.Token), ".")) != 3 {
		t.Fatalf("token is not a compact JWT: %q", res.Token)
	}
}

func TestJwtSignOverflowSuffixBoundary(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: generates a 4096-bit RSA key")
	}

	// EXPIRED large enough that the old float path definitely overflowed;
	// the big.Int modular path must still produce a valid token.
	env := &sdk_dto.Environment{}
	env.JWT = sdk_dro.Jwt{EXPIRED: 1_000_000}

	p := NewJsonWebToken(context.Background(), env, nil)

	res, err := p.Sign("client-b", "body")
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	if res.Expired != 1_000_000 {
		t.Fatalf("expired mismatch %d", res.Expired)
	}
}

func TestJwtIsRedisAvailable(t *testing.T) {
	env := &sdk_dto.Environment{}
	env.JWT = sdk_dro.Jwt{EXPIRED: 60}

	// IsRedisAvailable lives on the concrete type, not the interface.
	withNil := NewJsonWebToken(context.Background(), env, nil).(*jsonWebToken)
	if withNil.IsRedisAvailable() {
		t.Fatal("nil redis must report unavailable")
	}

	withFake := NewJsonWebToken(context.Background(), env, newFakeIRedis()).(*jsonWebToken)
	if !withFake.IsRedisAvailable() {
		t.Fatal("injected redis must report available")
	}
}

func TestJwtVerifyDelegatesToJose(t *testing.T) {
	fx := buildVerifyFixture(t)

	env := &sdk_dto.Environment{}
	env.JWT = sdk_dro.Jwt{EXPIRED: 60}

	p := NewJsonWebToken(context.Background(), env, fx.redis)

	token, err := p.Verify(fx.prefix, fx.token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if sub, ok := token.Subject(); !ok || sub != fx.metadata.SigKey[50:60] {
		t.Fatalf("subject mismatch %q ok=%v", sub, ok)
	}

	if _, err := p.Verify("unknown-prefix", fx.token); err == nil {
		t.Fatal("Verify with missing cert metadata must fail")
	}
}

func TestJwtSignVerifyRoundTripWithRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: generates a 4096-bit RSA key")
	}

	env := &sdk_dto.Environment{}
	env.JWT = sdk_dro.Jwt{SECRET: "unused", EXPIRED: 60}

	rds := newFakeIRedis()
	p := NewJsonWebToken(context.Background(), env, rds)

	sign, err := p.Sign("acquirer", map[string]any{"amount": 1500})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	if sign.Token == "" {
		t.Fatal("empty token")
	}

	// The token and body must be cached in redis under the documented keys.
	if n, err := rds.Exists(context.Background(), "jwt:token:acquirer"); err != nil || n != 1 {
		t.Fatalf("token cache missing: %d %v", n, err)
	}

	if _, err := rds.HGet(context.Background(), "jwt:cert:acquirer", "signature_metadata"); err != nil {
		t.Fatalf("cert metadata missing: %v", err)
	}

	// Second Sign must return the cached token unchanged.
	cached, err := p.Sign("acquirer", map[string]any{"amount": 9999})
	if err != nil {
		t.Fatalf("Sign cached: %v", err)
	}

	if cached.Token != sign.Token {
		t.Fatal("cached token differs from original")
	}

	if _, err := p.Verify("acquirer", sign.Token); err != nil {
		t.Fatalf("Verify round trip: %v", err)
	}
}
