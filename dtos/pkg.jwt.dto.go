package sdk_dto

import (
	"crypto/rsa"
	"time"

	sdk_dro "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type (
	JwtSignOption struct {
		PrivateKey *rsa.PrivateKey
		Claim      interface{}
		Kid        string
		SecretKey  string
		Iss        string
		Sub        string
		Aud        []string
		Exp        time.Time
		Nbf        float64
		Iat        time.Time
		Jti        string
	}

	SecretMetadata struct {
		PrivKeyRaw string `json:"privKeyRaw"`
		CipherKey  string `json:"cipherKey"`
	}

	SignatureMetadata struct {
		PrivKey    *rsa.PrivateKey            `json:"privKey"`
		PrivKeyRaw string                     `json:"privKeyRaw"`
		SigKey     string                     `json:"sigKey"`
		CipherKey  string                     `json:"cipherKey"`
		JweKey     sdk_dro.JweEncryptMetadata `json:"jweKey"`
	}

	SignMetadata struct {
		Token   string `json:"token"`
		Expired int    `json:"expired"`
	}
)
