package sdk_helper

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type signature struct{}

func NewSignature() sdk_inf.ISignature {
	return &signature{}
}

func (h *signature) encode(data []byte, mode string, encoding string) (string, error) {
	switch encoding {
	case sdk_cons.HEX:
		return hex.EncodeToString(data), nil
	case sdk_cons.BASE64:
		return base64.StdEncoding.EncodeToString(data), nil
	default:
		return "", errors.New("unsupported encoding")
	}
}

func (h *signature) decode(data string, encoding string) ([]byte, error) {
	switch encoding {
	case sdk_cons.HEX:
		return hex.DecodeString(data)
	case sdk_cons.BASE64:
		return base64.StdEncoding.DecodeString(data)
	default:
		return nil, errors.New("unsupported decoding")
	}
}

func (h *signature) GenerateAsymmetric(req *sdk_dto.Asymmetric) (res sdk_opt.SignatureResponse) {
	cert := NewCert()

	var sb strings.Builder
	sb.WriteString(req.ClientKey)
	sb.WriteString("|")
	sb.WriteString(req.TimeStamp)

	hashed := sha256.Sum256([]byte(sb.String()))

	if req.PrivateKeyType != sdk_cons.PRIVPKCS1 && req.PrivateKeyType != sdk_cons.PRIVPKCS8 {
		res.Error = errors.New("invalid private key type")
		return
	}

	certRes := cert.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyType:       req.PrivateKeyType,
		KeyRawPrivate: req.PrivateKey,
		Password:      req.Password,
	})

	if certRes.Error != nil {
		res.Error = certRes.Error
		return
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, certRes.KeyPrivate, crypto.SHA256, hashed[:])
	if err != nil {
		res.Error = err
		return
	}

	res.Signature, res.Error = h.encode(sig, "encode", req.Encoding)
	return
}

func (h *signature) GenerateSymmetric(req *sdk_dto.Symetric) (res sdk_opt.SignatureResponse) {
	bodyHash := sha256.Sum256(req.Body)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))

	var sb strings.Builder
	sb.WriteString(req.Method)
	sb.WriteString(":")
	sb.WriteString(req.Url)
	sb.WriteString(":")
	sb.WriteString(req.AccessToken)
	sb.WriteString(":")
	sb.WriteString(bodyHashHex)
	sb.WriteString(":")
	sb.WriteString(req.TimeStamp)

	mac := hmac.New(sha512.New, []byte(req.ClientSecret))
	mac.Write([]byte(strings.TrimSpace(sb.String())))

	res.Signature, res.Error = h.encode(mac.Sum(nil), "encode", req.Encoding)
	return
}

func (h *signature) VerifyAsymmetric(req *sdk_dto.VerifyAsymmetric) (res sdk_opt.SignatureResponse) {
	cert := NewCert()

	var sb strings.Builder
	sb.WriteString(req.ClientId)
	sb.WriteString("|")
	sb.WriteString(req.Timestamp)
	hashed := sha256.Sum256([]byte(sb.String()))

	if req.PublicKeyType != sdk_cons.PUBPKCS1 && req.PublicKeyType != sdk_cons.PUBPKCS8 {
		res.Error = errors.New("invalid public key type")
		return
	}

	certRes := cert.PublicKeyRawToKey(&sdk_dto.PublicKeyRawToKey{
		KeyType:      req.PublicKeyType,
		KeyRawPublic: req.PublicKey,
	})
	if certRes.Error != nil {
		res.Error = certRes.Error
		return
	}

	sigBytes, err := h.decode(req.Signature, req.Encoding)
	if err != nil {
		res.Error = err
		return
	}

	err = rsa.VerifyPKCS1v15(certRes.KeyPublic, crypto.SHA256, hashed[:], sigBytes)
	if err != nil {
		res.Error = errors.New("signature verification failed")
		return
	}

	res.Signature = req.Signature
	return
}

func (h *signature) VerifySymmetric(req *sdk_dto.VerifySymetric) (res sdk_opt.SignatureResponse) {
	internalSigRes := h.GenerateSymmetric(&sdk_dto.Symetric{
		Method:       req.Method,
		Url:          req.Url,
		AccessToken:  req.AccessToken,
		Body:         req.Body,
		TimeStamp:    req.TimeStamp,
		ClientSecret: req.ClientSecret,
		Encoding:     req.Encoding,
	})

	if internalSigRes.Error != nil {
		res.Error = internalSigRes.Error
		return
	}

	if !hmac.Equal([]byte(req.Signature), []byte(internalSigRes.Signature)) {
		res.Error = fmt.Errorf("signature mismatch")
		return
	}

	res.Signature = req.Signature
	return
}
