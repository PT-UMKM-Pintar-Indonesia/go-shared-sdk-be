package sdk_inf

import (
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/dtos"
	opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/outputs"
)

type ICert interface {
	GenerateKey(req sdk_dto.GeneratePrivateKey) (res opt.CertResponse)
	PrivateKeyRawToKey(sdk_dto sdk_dto.PrivateKeyRawToKey) (res opt.CertResponse)
	PublicKeyRawToKey(sdk_dto sdk_dto.PublicKeyRawToKey) (res opt.CertResponse)
	PrivateKeyToRaw(sdk_dto sdk_dto.PrivateKeyToRaw) (res opt.CertResponse)
	PublicKeyToRaw(sdk_dto sdk_dto.PublicKeyToRaw) (res opt.CertResponse)
	PrivateKeyBase64ToRaw(sdk_dto sdk_dto.PrivateKeyBase64ToRaw) (res opt.CertResponse)
	PublicKeyBase64ToRaw(sdk_dto sdk_dto.PublicKeyBase64ToRaw) (res opt.CertResponse)
}
