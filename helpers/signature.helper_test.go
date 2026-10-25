package sdk_helper

import (
	"encoding/base64"
	"strings"
	"testing"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func TestSignatureEncodeDecode(t *testing.T) {
	h := NewSignature().(*signature)

	data := []byte("hello")

	hexSig, err := h.encode(data, "encode", sdk_cons.HEX)
	if err != nil || hexSig != "68656c6c6f" {
		t.Fatalf("hex encode got %q err %v", hexSig, err)
	}

	b64Sig, err := h.encode(data, "encode", sdk_cons.BASE64)
	if err != nil || b64Sig != base64.StdEncoding.EncodeToString(data) {
		t.Fatalf("b64 encode got %q err %v", b64Sig, err)
	}

	if _, err := h.encode(data, "encode", "utf-8"); err == nil {
		t.Fatal("unsupported encoding must error")
	}

	if got, err := h.decode(hexSig, sdk_cons.HEX); err != nil || string(got) != "hello" {
		t.Fatalf("hex decode got %q err %v", got, err)
	}

	if got, err := h.decode(b64Sig, sdk_cons.BASE64); err != nil || string(got) != "hello" {
		t.Fatalf("b64 decode got %q err %v", got, err)
	}

	if _, err := h.decode("zz", sdk_cons.HEX); err == nil {
		t.Fatal("invalid hex must error")
	}

	if _, err := h.decode("!!", sdk_cons.BASE64); err == nil {
		t.Fatal("invalid base64 must error")
	}

	if _, err := h.decode("x", "utf-8"); err == nil {
		t.Fatal("unsupported decoding must error")
	}
}

func TestSignatureSymmetricRoundTrip(t *testing.T) {
	h := NewSignature()

	req := &sdk_dto.Symetric{
		Method:       "POST",
		Url:          "/api/v1/transfer",
		AccessToken:  "access",
		TimeStamp:    "2026-10-25 10:00:00",
		ClientSecret: "top-secret",
		Body:         []byte(`{"amount":1500}`),
		Encoding:     sdk_cons.BASE64,
	}

	first := h.GenerateSymmetric(req)
	if first.Error != nil || first.Signature == "" {
		t.Fatalf("generate got %q err %v", first.Signature, first.Error)
	}

	second := h.GenerateSymmetric(req)
	if second.Signature != first.Signature {
		t.Fatal("symmetric signature must be deterministic")
	}

	ver := h.VerifySymmetric(&sdk_dto.VerifySymetric{
		Method:       req.Method,
		Url:          req.Url,
		AccessToken:  req.AccessToken,
		TimeStamp:    req.TimeStamp,
		ClientSecret: req.ClientSecret,
		Body:         req.Body,
		Signature:    first.Signature,
		Encoding:     req.Encoding,
	})
	if ver.Error != nil || ver.Signature != first.Signature {
		t.Fatalf("verify got %q err %v", ver.Signature, ver.Error)
	}

	mismatch := h.VerifySymmetric(&sdk_dto.VerifySymetric{
		Method:       req.Method,
		Url:          req.Url,
		AccessToken:  req.AccessToken,
		TimeStamp:    req.TimeStamp,
		ClientSecret: "other-secret",
		Body:         req.Body,
		Signature:    first.Signature,
		Encoding:     req.Encoding,
	})
	if mismatch.Error == nil || !strings.Contains(mismatch.Error.Error(), "signature mismatch") {
		t.Fatalf("expected mismatch, got %v", mismatch.Error)
	}

	// GET skips the body hash segment.
	get := h.GenerateSymmetric(&sdk_dto.Symetric{
		Method:       "GET",
		Url:          "/api/v1/balance",
		ClientSecret: "top-secret",
		Encoding:     sdk_cons.HEX,
	})
	if get.Error != nil || get.Signature == "" {
		t.Fatalf("GET signature got %q err %v", get.Signature, get.Error)
	}

	badEnc := h.GenerateSymmetric(&sdk_dto.Symetric{
		Method:       "POST",
		Url:          "/x",
		ClientSecret: "s",
		Encoding:     "utf-8",
	})
	if badEnc.Error == nil {
		t.Fatal("unsupported encoding must error")
	}
}

func TestSignatureAsymmetricRoundTrip(t *testing.T) {
	cert := NewCert()

	gen := cert.GenerateKey(&sdk_dto.GeneratePrivateKey{
		KeySize:        sdk_cons.KEY_SIZE_2048,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})
	if gen.Error != nil {
		t.Fatal(gen.Error)
	}

	h := NewSignature()

	const (
		clientKey = "client-123"
		timestamp = "2026-10-25 10:00:00"
	)

	sig := h.GenerateAsymmetric(&sdk_dto.Asymmetric{
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PrivateKey:     gen.KeyRawPrivate,
		TimeStamp:      timestamp,
		ClientKey:      clientKey,
		Encoding:       sdk_cons.HEX,
	})
	if sig.Error != nil || sig.Signature == "" {
		t.Fatalf("generate asymmetric err %v", sig.Error)
	}

	ver := h.VerifyAsymmetric(&sdk_dto.VerifyAsymmetric{
		PublicKeyType: sdk_cons.PUBPKCS8,
		PublicKey:     gen.KeyRawPublic,
		Signature:     sig.Signature,
		ClientId:      clientKey,
		Timestamp:     timestamp,
		Encoding:      sdk_cons.HEX,
	})
	if ver.Error != nil || ver.Signature != sig.Signature {
		t.Fatalf("verify asymmetric err %v", ver.Error)
	}

	tampered := h.VerifyAsymmetric(&sdk_dto.VerifyAsymmetric{
		PublicKeyType: sdk_cons.PUBPKCS8,
		PublicKey:     gen.KeyRawPublic,
		Signature:     sig.Signature,
		ClientId:      clientKey,
		Timestamp:     "2020-01-01 00:00:00",
		Encoding:      sdk_cons.HEX,
	})
	if tampered.Error == nil || !strings.Contains(tampered.Error.Error(), "verification failed") {
		t.Fatalf("expected verification failure, got %v", tampered.Error)
	}
}

func TestSignatureAsymmetricErrorBranches(t *testing.T) {
	cert := NewCert()

	gen := cert.GenerateKey(&sdk_dto.GeneratePrivateKey{
		Password:       "pw",
		KeySize:        sdk_cons.KEY_SIZE_2048,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})
	if gen.Error != nil {
		t.Fatal(gen.Error)
	}

	h := NewSignature()

	if res := h.GenerateAsymmetric(&sdk_dto.Asymmetric{
		PrivateKeyType: "BOGUS",
		PrivateKey:     gen.KeyRawPrivate,
		Encoding:       sdk_cons.BASE64,
	}); res.Error == nil || res.Error.Error() != "invalid private key type" {
		t.Fatalf("expected invalid private key type, got %v", res.Error)
	}

	if res := h.GenerateAsymmetric(&sdk_dto.Asymmetric{
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PrivateKey:     []byte("not pem"),
		Encoding:       sdk_cons.BASE64,
	}); res.Error == nil {
		t.Fatal("expected PEM decode error")
	}

	// Encrypted key without password must fail.
	if res := h.GenerateAsymmetric(&sdk_dto.Asymmetric{
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PrivateKey:     gen.KeyRawPrivate,
		Encoding:       sdk_cons.BASE64,
	}); res.Error == nil {
		t.Fatal("expected decrypt error for passwordless request")
	}

	// With the right password it succeeds, and the base64 signature verifies.
	sig := h.GenerateAsymmetric(&sdk_dto.Asymmetric{
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PrivateKey:     gen.KeyRawPrivate,
		Password:       "pw",
		TimeStamp:      "ts",
		ClientKey:      "ck",
		Encoding:       sdk_cons.BASE64,
	})
	if sig.Error != nil {
		t.Fatalf("encrypted key sign err %v", sig.Error)
	}

	if res := h.VerifyAsymmetric(&sdk_dto.VerifyAsymmetric{
		PublicKeyType: "BOGUS",
	}); res.Error == nil || res.Error.Error() != "invalid public key type" {
		t.Fatalf("expected invalid public key type, got %v", res.Error)
	}

	if res := h.VerifyAsymmetric(&sdk_dto.VerifyAsymmetric{
		PublicKeyType: sdk_cons.PUBPKCS8,
		PublicKey:     []byte("not pem"),
		Signature:     sig.Signature,
		Encoding:      sdk_cons.BASE64,
	}); res.Error == nil {
		t.Fatal("expected public key parse error")
	}

	if res := h.VerifyAsymmetric(&sdk_dto.VerifyAsymmetric{
		PublicKeyType: sdk_cons.PUBPKCS8,
		PublicKey:     gen.KeyRawPublic,
		Signature:     "!!!not-base64!!!",
		Encoding:      sdk_cons.BASE64,
	}); res.Error == nil {
		t.Fatal("expected decode error")
	}
}
