package sdk_pkg

import (
	"testing"

	"github.com/go-oauth2/oauth2/v4/manage"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

// A11 companion: the oauth2 wrappers used to dereference req fields
// without any nil checks, panicking on malformed input.
func TestOauth2NilGuards(t *testing.T) {
	p := NewOauth2()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("oauth2 call panicked: %v", r)
		}
	}()

	if _, err := p.GenerateClientCredentials(nil); err == nil {
		t.Fatal("expected error for nil req")
	}

	if _, err := p.GenerateClientCredentials(&sdk_dto.GenerateClientCredentialsOptions{}); err == nil {
		t.Fatal("expected error for zero-value req")
	}

	if _, err := p.GenerateClientCredentials(&sdk_dto.GenerateClientCredentialsOptions{
		Manager: manage.NewDefaultManager(),
	}); err == nil {
		t.Fatal("expected error for missing client store")
	}

	if _, _, err := p.GenerateServerClientCredentials(nil); err == nil {
		t.Fatal("expected error for nil server req")
	}

	if _, err := p.LoadAccessToken(nil); err == nil {
		t.Fatal("expected error for nil load req")
	}

	if _, err := p.LoadAccessToken(&sdk_dto.LoadAccessTokenOptions{}); err == nil {
		t.Fatal("expected error for missing server")
	}
}

func TestOauth2Constructors(t *testing.T) {
	p := NewOauth2()

	if p.Manager() == nil {
		t.Fatal("Manager must not be nil")
	}

	if p.Client() == nil {
		t.Fatal("Client store must not be nil")
	}

	if p.Server(p.Manager()) == nil {
		t.Fatal("Server must not be nil")
	}
}
