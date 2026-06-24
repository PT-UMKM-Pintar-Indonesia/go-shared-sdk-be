package sdk_helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type cert struct{}

func NewCert() sdk_inf.ICert {
	return &cert{}
}

func (h *cert) GenerateKey(req sdk_dto.GeneratePrivateKey) (res sdk_opt.CertResponse) {
	if req.KeySize < 2048 {
		res.Error = errors.New("security risk: key size must be at least 2048 bits")
		return
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, int(req.KeySize))
	if err != nil {
		res.Error = fmt.Errorf("failed to generate rsa key: %w", err)
		return
	}

	privRes := h.PrivateKeyToRaw(sdk_dto.PrivateKeyToRaw{
		KeyType:    req.PrivateKeyType,
		KeyPrivate: privateKey,
	})

	if privRes.Error != nil {
		res.Error = privRes.Error
		return
	}

	if req.Password != "" {
		block, _ := pem.Decode(privRes.KeyRawPrivate)

		encBlock, err := x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(req.Password), x509.PEMCipherAES256)
		if err != nil {
			res.Error = fmt.Errorf("failed to encrypt pem: %w", err)
			return
		}

		res.KeyRawPrivate = pem.EncodeToMemory(encBlock)
	} else {
		res.KeyRawPrivate = privRes.KeyRawPrivate
	}

	pubRes := h.PublicKeyToRaw(sdk_dto.PublicKeyToRaw{
		KeyType:   req.PublicKeyType,
		KeyPublic: &privateKey.PublicKey,
	})

	if pubRes.Error != nil {
		res.Error = pubRes.Error
		return
	}

	res.KeyRawPublic = pubRes.KeyRawPublic
	return
}

func (h *cert) PrivateKeyRawToKey(req sdk_dto.PrivateKeyRawToKey) (res sdk_opt.CertResponse) {
	block, _ := pem.Decode(req.KeyRawPrivate)
	if block == nil {
		res.Error = errors.New("failed to decode PEM block")
		return
	}

	der := block.Bytes
	var err error

	if x509.IsEncryptedPEMBlock(block) {
		if req.Password == "" {
			res.Error = errors.New("private key is encrypted but no password provided")
			return
		}
		der, err = x509.DecryptPEMBlock(block, []byte(req.Password))
		if err != nil {
			res.Error = fmt.Errorf("failed to decrypt PEM: %w", err)
			return
		}
	}

	var priv interface{}
	if req.KeyType == sdk_cons.PRIVPKCS1 {
		priv, err = x509.ParsePKCS1PrivateKey(der)
	} else {
		priv, err = x509.ParsePKCS8PrivateKey(der)
	}

	if err != nil {
		res.Error = fmt.Errorf("failed to parse private key: %w", err)
		return
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		res.Error = errors.New("not an RSA private key")
		return
	}

	res.KeyType = req.KeyType
	res.KeyPrivate = rsaPriv
	res.KeyPublic = &rsaPriv.PublicKey
	return
}

func (h *cert) PublicKeyRawToKey(req sdk_dto.PublicKeyRawToKey) (res sdk_opt.CertResponse) {
	block, _ := pem.Decode(req.KeyRawPublic)
	if block == nil {
		res.Error = errors.New("failed to decode PEM block")
		return
	}

	var pub interface{}
	var err error

	if req.KeyType == sdk_cons.PUBPKCS1 {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	} else {
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		res.Error = fmt.Errorf("failed to parse public key: %w", err)
		return
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		res.Error = errors.New("not an RSA public key")
		return
	}

	res.KeyType = req.KeyType
	res.KeyPublic = rsaPub
	return
}

func (h *cert) PrivateKeyToRaw(req sdk_dto.PrivateKeyToRaw) (res sdk_opt.CertResponse) {
	var der []byte
	var err error

	if req.KeyType == sdk_cons.PRIVPKCS1 {
		der = x509.MarshalPKCS1PrivateKey(req.KeyPrivate)
	} else if req.KeyType == sdk_cons.PRIVPKCS8 {
		der, err = x509.MarshalPKCS8PrivateKey(req.KeyPrivate)
	} else {
		res.Error = errors.New("unsupported private key type")
		return
	}

	if err != nil {
		res.Error = err
		return
	}

	res.KeyRawPrivate = pem.EncodeToMemory(&pem.Block{Type: req.KeyType, Bytes: der})
	res.KeyType = req.KeyType
	return
}

func (h *cert) PublicKeyToRaw(req sdk_dto.PublicKeyToRaw) (res sdk_opt.CertResponse) {
	var der []byte
	var err error

	if req.KeyType == sdk_cons.PUBPKCS1 {
		der = x509.MarshalPKCS1PublicKey(req.KeyPublic)
	} else if req.KeyType == sdk_cons.PUBPKCS8 {
		der, err = x509.MarshalPKIXPublicKey(req.KeyPublic)
	} else {
		res.Error = errors.New("unsupported public key type")
		return
	}

	if err != nil {
		res.Error = err
		return
	}

	res.KeyRawPublic = pem.EncodeToMemory(&pem.Block{Type: req.KeyType, Bytes: der})
	res.KeyType = req.KeyType
	return
}

func (h *cert) base64ToRaw(input string, expectedTypes ...string) ([]byte, string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input))
	if err != nil {
		return nil, "", fmt.Errorf("base64 decode failed: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, "", errors.New("invalid PEM after base64 decode")
	}

	for _, t := range expectedTypes {
		if block.Type == t {
			return pem.EncodeToMemory(block), block.Type, nil
		}
	}

	return nil, "", fmt.Errorf("unexpected PEM type: %s", block.Type)
}

func (h *cert) PrivateKeyBase64ToRaw(req sdk_dto.PrivateKeyBase64ToRaw) (res sdk_opt.CertResponse) {
	raw, keyType, err := h.base64ToRaw(req.KeyRawPrivate, sdk_cons.PRIVPKCS1, sdk_cons.PRIVPKCS8, sdk_cons.CERTIFICATE)
	if err != nil {
		res.Error = err
		return
	}

	res.KeyRawPrivate = raw
	res.KeyType = keyType
	return
}

func (h *cert) PublicKeyBase64ToRaw(req sdk_dto.PublicKeyBase64ToRaw) (res sdk_opt.CertResponse) {
	raw, keyType, err := h.base64ToRaw(req.KeyRawPublic, sdk_cons.PUBPKCS1, sdk_cons.PUBPKCS8, sdk_cons.CERTIFICATE)
	if err != nil {
		res.Error = err
		return
	}

	res.KeyRawPublic = raw
	res.KeyType = keyType
	return
}
