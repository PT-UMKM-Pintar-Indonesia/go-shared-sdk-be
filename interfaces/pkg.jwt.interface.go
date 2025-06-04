package sdk_inf

import sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"

type IJsonWebToken interface {
	Sign(prefix string, body any) (*sdk_opt.SignMetadata, error)
}
