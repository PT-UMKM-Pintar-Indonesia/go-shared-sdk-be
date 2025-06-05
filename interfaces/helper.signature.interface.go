package sdk_inf

import (
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type ISignature interface {
	GenerateAsymmetric(req sdk_dto.Asymmetric) (res sdk_opt.SignatureResponse)
	GenerateSymmetric(req sdk_dto.Symetric) (res sdk_opt.SignatureResponse)
	VerifyAsymmetric(req sdk_dto.VerifyAsymmetric) (res sdk_opt.SignatureResponse)
	VerifySymmetric(req sdk_dto.VerifySymetric) (res sdk_opt.SignatureResponse)
}
