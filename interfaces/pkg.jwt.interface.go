package sdk_inf

import (
	"github.com/lestrrat-go/jwx/v3/jwt"

	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type IJsonWebToken interface {
	Sign(prefix string, body any) (*sdk_opt.SignMetadata, error)
	Verify(prefix string, token string) (jwt.Token, error)
}
