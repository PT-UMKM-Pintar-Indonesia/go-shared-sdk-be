package sdk_helper

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func x509IsEncryptedPEMBlockForTest(block *pem.Block) bool {
	return x509.IsEncryptedPEMBlock(block) //nolint:staticcheck // legacy PEM encryption is what the SDK produces
}

// A11 regression: nil request used to panic inside GenerateKey.
func TestCertGenerateKeyNilRequest(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GenerateKey(nil) panicked: %v", r)
		}
	}()

	res := NewCert().GenerateKey(nil)
	if res.Error == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestCertGenerateKeyRejectsSmallKeys(t *testing.T) {
	res := NewCert().GenerateKey(&sdk_dto.GeneratePrivateKey{
		KeySize:        1024,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})

	if res.Error == nil || !strings.Contains(res.Error.Error(), "at least 2048") {
		t.Fatalf("expected key size error, got %v", res.Error)
	}
}

func TestCertGenerateKeyAndReparse(t *testing.T) {
	c := NewCert()

	res := c.GenerateKey(&sdk_dto.GeneratePrivateKey{
		KeySize:        2048,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})

	if res.Error != nil {
		t.Fatalf("generation failed: %v", res.Error)
	}

	if len(res.KeyRawPrivate) == 0 || len(res.KeyRawPublic) == 0 {
		t.Fatal("expected raw key material")
	}

	back := c.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       sdk_cons.PRIVPKCS8,
		KeyRawPrivate: res.KeyRawPrivate,
	})

	if back.Error != nil {
		t.Fatalf("reparse failed: %v", back.Error)
	}

	if back.KeyPrivate == nil || back.KeyPrivate.N.BitLen() < 2048 {
		t.Fatal("parsed key is missing or too small")
	}
}

func TestCertGenerateKeyWithPassword(t *testing.T) {
	c := NewCert()

	res := c.GenerateKey(&sdk_dto.GeneratePrivateKey{
		KeySize:        2048,
		Password:       "hunter2-hunter2",
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})

	if res.Error != nil {
		t.Fatalf("generation failed: %v", res.Error)
	}

	block, _ := pem.Decode(res.KeyRawPrivate)
	if block == nil {
		t.Fatal("expected PEM output")
	}

	if !x509IsEncryptedPEMBlockForTest(block) {
		t.Fatal("expected an encrypted PEM block")
	}

	// Wrong password must fail with an error, not a panic.
	wrong := c.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       sdk_cons.PRIVPKCS8,
		KeyRawPrivate: res.KeyRawPrivate,
		Password:      "wrong-password",
	})

	if wrong.Error == nil {
		t.Fatal("expected failure with wrong password")
	}

	// Missing password for an encrypted key must error too.
	missing := c.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       sdk_cons.PRIVPKCS8,
		KeyRawPrivate: res.KeyRawPrivate,
	})

	if missing.Error == nil {
		t.Fatal("expected failure without password")
	}

	right := c.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       sdk_cons.PRIVPKCS8,
		KeyRawPrivate: res.KeyRawPrivate,
		Password:      "hunter2-hunter2",
	})

	if right.Error != nil {
		t.Fatalf("expected success with correct password, got %v", right.Error)
	}
}

func TestCertPrivateKeyRawToKeyGarbage(t *testing.T) {
	res := NewCert().PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       sdk_cons.PRIVPKCS8,
		KeyRawPrivate: []byte("not a pem block"),
	})

	if res.Error == nil {
		t.Fatal("expected PEM decode error")
	}
}
