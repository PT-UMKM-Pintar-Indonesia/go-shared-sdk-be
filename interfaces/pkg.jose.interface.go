package sdk_inf

import (
	"crypto/rsa"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/outputs"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type IJose interface {
	JweEncrypt(publicKey *rsa.PublicKey, plainText string) ([]byte, *sdk_opt.JweEncryptMetadata, error)
	JweDecrypt(privateKey *rsa.PrivateKey, cipherText []byte) (string, error)
	ImportJsonWebKey(jwkKey jwk.Key) (*sdk_opt.JwkMetadata, error)
	ExportJsonWebKey(privateKey *rsa.PrivateKey) (*sdk_opt.JwkMetadata, error)
	JwtSign(options *sdk_dto.JwtSignOption) ([]byte, error)
	JwtVerify(prefix string, token string, redis IRedis) (*jwt.Token, error)
}
