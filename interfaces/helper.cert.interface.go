package sdk_inf

import (
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type ICert interface {
	GenerateKey(req *sdk_dto.GeneratePrivateKey) (res opt.CertResponse)
	PrivateKeyRawToKey(req *sdk_dto.PrivateKeyRawToKey) (res opt.CertResponse)
	PublicKeyRawToKey(req *sdk_dto.PublicKeyRawToKey) (res opt.CertResponse)
	PrivateKeyToRaw(req *sdk_dto.PrivateKeyToRaw) (res opt.CertResponse)
	PublicKeyToRaw(req *sdk_dto.PublicKeyToRaw) (res opt.CertResponse)
	PrivateKeyBase64ToRaw(req *sdk_dto.PrivateKeyBase64ToRaw) (res opt.CertResponse)
	PublicKeyBase64ToRaw(req *sdk_dto.PublicKeyBase64ToRaw) (res opt.CertResponse)
}
