package sdk_pkg

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_dro "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type verifyFixture struct {
	token    string
	prefix   string
	redis    *fakeIRedis
	metadata *sdk_dro.SignatureMetadata
}

// buildVerifyFixture mirrors the real Sign flow's metadata layout: an
// encrypted PKCS#8 key plus the hex signature that JwtVerify unpacks from
// the jwt:cert:<prefix> hash. RSA-2048 keeps the fixture fast enough for
// unit runs.
func buildVerifyFixture(t *testing.T) *verifyFixture {
	t.Helper()

	ctx := context.Background()
	cert := sdk_helper.NewCert()
	password := "fixture-cipher-key"

	gen := cert.GenerateKey(&sdk_dto.GeneratePrivateKey{
		Password:       password,
		KeySize:        sdk_cons.KEY_SIZE_2048,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})
	if gen.Error != nil {
		t.Fatalf("generate key: %v", gen.Error)
	}

	priv := cert.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       sdk_cons.PRIVPKCS8,
		KeyRawPrivate: gen.KeyRawPrivate,
		Password:      password,
	})
	if priv.Error != nil {
		t.Fatalf("import private: %v", priv.Error)
	}

	pub := cert.PublicKeyRawToKey(&sdk_dto.PublicKeyRawToKey{
		KeyType:      sdk_cons.PUBPKCS8,
		KeyRawPublic: gen.KeyRawPublic,
	})
	if pub.Error != nil {
		t.Fatalf("import public: %v", pub.Error)
	}

	sig := make([]byte, 64)
	if _, err := rand.Read(sig); err != nil {
		t.Fatal(err)
	}

	sigKey := hex.EncodeToString(sig)

	jose := NewJose(ctx)

	_, jweKey, err := jose.JweEncrypt(pub.KeyPublic, sigKey)
	if err != nil {
		t.Fatalf("jwe encrypt: %v", err)
	}

	token, err := jose.JwtSign(&sdk_dto.JwtSignOption{
		SecretKey:  password,
		Kid:        jweKey.CipherText,
		PrivateKey: priv.KeyPrivate,
		Aud:        []string{sigKey[10:20]},
		Iss:        sigKey[30:40],
		Sub:        sigKey[50:60],
		Jti:        "jti-fixture",
		Iat:        time.Now().UTC().Add(-time.Minute),
		Exp:        time.Now().Add(time.Minute),
		Claim:      time.Now().Format(sdk_cons.DATE_TIME_FORMAT),
	})
	if err != nil {
		t.Fatalf("jwt sign: %v", err)
	}

	metadata := &sdk_dro.SignatureMetadata{
		PrivKeyRaw: string(gen.KeyRawPrivate),
		SigKey:     sigKey,
		CipherKey:  password,
		JweKey:     *jweKey,
	}

	rds := newFakeIRedis()

	if err := rds.HSetEx(ctx, "jwt:cert:merchant", time.Minute, "signature_metadata", mustJSON(t, metadata)); err != nil {
		t.Fatalf("seed metadata: %v", err)
	}

	return &verifyFixture{
		token:    string(token),
		prefix:   "merchant",
		redis:    rds,
		metadata: metadata,
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestJwtVerifyHappyPath(t *testing.T) {
	fx := buildVerifyFixture(t)

	jose := NewJose(context.Background())

	token, err := jose.JwtVerify(fx.prefix, fx.token, fx.redis)
	if err != nil {
		t.Fatalf("JwtVerify: %v", err)
	}

	if jti, ok := token.JwtID(); !ok || jti != "jti-fixture" {
		t.Fatalf("jti mismatch %q ok=%v", jti, ok)
	}

	aud, ok := token.Audience()
	if !ok || len(aud) != 1 || aud[0] != fx.metadata.SigKey[10:20] {
		t.Fatalf("audience mismatch %v ok=%v", aud, ok)
	}

	iss, ok := token.Issuer()
	if !ok || iss != fx.metadata.SigKey[30:40] {
		t.Fatalf("issuer mismatch %q ok=%v", iss, ok)
	}

	sub, ok := token.Subject()
	if !ok || sub != fx.metadata.SigKey[50:60] {
		t.Fatalf("subject mismatch %q ok=%v", sub, ok)
	}

	var timestampClaim any
	if err := token.Get("timestamp", &timestampClaim); err != nil {
		t.Fatalf("timestamp claim missing: %v", err)
	}
}

func TestJwtVerifyNegativeCases(t *testing.T) {
	jose := NewJose(context.Background())

	t.Run("missing redis metadata", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		empty := newFakeIRedis()

		_, err := jose.JwtVerify(fx.prefix, fx.token, empty)
		if err == nil || !strings.Contains(err.Error(), "redis hget failed") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("corrupt metadata json", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		if err := fx.redis.HSet(context.Background(), "jwt:cert:merchant", "signature_metadata", []byte("{not json")); err != nil {
			t.Fatal(err)
		}

		_, err := jose.JwtVerify(fx.prefix, fx.token, fx.redis)
		if err == nil || !strings.Contains(err.Error(), "parser unmarshal failed") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("zero metadata", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		if err := fx.redis.HSet(context.Background(), "jwt:cert:merchant", "signature_metadata", []byte("{}")); err != nil {
			t.Fatal(err)
		}

		_, err := jose.JwtVerify(fx.prefix, fx.token, fx.redis)
		if err == nil || !strings.Contains(err.Error(), "invalid secretkey or signature") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("wrong key id", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		mutated := *fx.metadata
		mutated.JweKey.CipherText = "other-kid"

		if err := fx.redis.HSet(context.Background(), "jwt:cert:merchant", "signature_metadata", mustJSON(t, &mutated)); err != nil {
			t.Fatal(err)
		}

		_, err := jose.JwtVerify(fx.prefix, fx.token, fx.redis)
		if err == nil || !strings.Contains(err.Error(), "invalid keyid") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("short signature key", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		mutated := *fx.metadata
		mutated.SigKey = strings.Repeat("a", 40)

		if err := fx.redis.HSet(context.Background(), "jwt:cert:merchant", "signature_metadata", mustJSON(t, &mutated)); err != nil {
			t.Fatal(err)
		}

		_, err := jose.JwtVerify(fx.prefix, fx.token, fx.redis)
		if err == nil || !strings.Contains(err.Error(), "invalid signature key length") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("wrong cipher key", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		mutated := *fx.metadata
		mutated.CipherKey = "wrong-password"

		if err := fx.redis.HSet(context.Background(), "jwt:cert:merchant", "signature_metadata", mustJSON(t, &mutated)); err != nil {
			t.Fatal(err)
		}

		_, err := jose.JwtVerify(fx.prefix, fx.token, fx.redis)
		if err == nil {
			t.Fatal("expected decrypt failure")
		}
	})

	t.Run("garbage token", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		_, err := jose.JwtVerify(fx.prefix, "not-a-jws", fx.redis)
		if err == nil || !strings.Contains(err.Error(), "jws parse failed") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("tampered token", func(t *testing.T) {
		fx := buildVerifyFixture(t)

		parts := strings.Split(fx.token, ".")
		if len(parts) != 3 {
			t.Fatalf("unexpected token shape")
		}

		_, err := jose.JwtVerify(fx.prefix, parts[0]+"."+parts[1]+".AAAA", fx.redis)
		if err == nil || !strings.Contains(err.Error(), "jws verify failed") {
			t.Fatalf("got %v", err)
		}
	})
}
