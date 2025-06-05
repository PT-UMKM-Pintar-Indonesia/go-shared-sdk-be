package sdk_helper

import (
	cpt "crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type signature struct{}

func NewSignature() sdk_inf.ISignature {
	return signature{}
}

func (p signature) GenerateAsymmetric(req sdk_dto.Asymmetric) (res sdk_opt.SignatureResponse) {
	cert := NewCert()
	salt := rand.Reader

	cipherBody := []byte(req.ClientKey + "|" + req.TimeStamp)
	cipherBodyHash256 := sha256.New()
	cipherBodyHash256.Write(cipherBody)
	cipherBodyHash := cipherBodyHash256.Sum(nil)

	privateKeyRawToKeyReq := sdk_dto.PrivateKeyRawToKey{}
	privateKeyRawToKeyRes := sdk_opt.CertResponse{}

	switch req.PrivateKeyType {

	case sdk_cons.PRIVPKCS1:
		privateKeyRawToKeyReq.KeyType = req.PrivateKeyType
		privateKeyRawToKeyReq.KeyRawPrivate = req.PrivateKey
		privateKeyRawToKeyReq.Password = req.Password

		privateKeyRawToKeyRes = cert.PrivateKeyRawToKey(privateKeyRawToKeyReq)
		if privateKeyRawToKeyRes.Error != nil {
			res.Error = privateKeyRawToKeyRes.Error
			return
		}

		break

	case sdk_cons.PRIVPKCS8:
		privateKeyRawToKeyReq.KeyType = req.PrivateKeyType
		privateKeyRawToKeyReq.KeyRawPrivate = req.PrivateKey
		privateKeyRawToKeyReq.Password = req.Password

		privateKeyRawToKeyRes = cert.PrivateKeyRawToKey(privateKeyRawToKeyReq)
		if privateKeyRawToKeyRes.Error != nil {
			res.Error = privateKeyRawToKeyRes.Error
			return
		}

		break

	default:
		res.Error = errors.New("Invalid GenerateAsymmetric PEM PrivateKey certificate unsupported")
		return
	}

	if err := privateKeyRawToKeyRes.KeyPrivate.Validate(); err != nil {
		res.Error = err
		return
	}

	signature, err := rsa.SignPKCS1v15(salt, privateKeyRawToKeyRes.KeyPrivate, cpt.SHA256, cipherBodyHash)
	if err != nil {
		res.Error = err
		return
	}

	if err := rsa.VerifyPKCS1v15(&privateKeyRawToKeyRes.KeyPrivate.PublicKey, cpt.SHA256, cipherBodyHash, signature); err != nil {
		res.Error = err
		return
	}

	res.Signature = base64.StdEncoding.EncodeToString(signature)
	return
}

func (p signature) GenerateSymmetric(req sdk_dto.Symetric) (res sdk_opt.SignatureResponse) {
	sha256 := cpt.SHA256.New()

	if _, err := sha256.Write(req.Body); err != nil {
		res.Error = err
		return
	}

	sha256SecretKey := strings.ToLower(hex.EncodeToString(sha256.Sum(nil)))

	hmac512Body := req.Method + ":" + req.Url + ":" + req.AccessToken + ":" + sha256SecretKey + ":" + req.TimeStamp
	hmac512 := hmac.New(cpt.SHA512.New, []byte(req.ClientSecret))

	if _, err := hmac512.Write([]byte(strings.TrimSpace(hmac512Body))); err != nil {
		res.Error = err
		return
	}

	res.Signature = base64.StdEncoding.EncodeToString(hmac512.Sum(nil))
	return
}

func (p signature) VerifyAsymmetric(req sdk_dto.VerifyAsymmetric) (res sdk_opt.SignatureResponse) {
	cert := NewCert()

	cipherBody := []byte(req.ClientId + "|" + req.Timestamp)
	cipherBodyHash256 := sha256.New()
	cipherBodyHash256.Write(cipherBody)
	cipherBodyHash := cipherBodyHash256.Sum(nil)

	publicKeyRawToKeyReq := sdk_dto.PublicKeyRawToKey{}
	publicKeyRawToKeyRes := sdk_opt.CertResponse{}

	switch req.PublicKeyType {

	case sdk_cons.PRIVPKCS1:
		publicKeyRawToKeyReq.KeyType = req.PublicKeyType
		publicKeyRawToKeyReq.KeyRawPublic = req.PublicKey

		publicKeyRawToKeyRes = cert.PublicKeyRawToKey(publicKeyRawToKeyReq)
		if publicKeyRawToKeyRes.Error != nil {
			res.Error = publicKeyRawToKeyRes.Error
			return
		}

		break

	case sdk_cons.PRIVPKCS8:
		publicKeyRawToKeyReq.KeyType = req.PublicKeyType
		publicKeyRawToKeyReq.KeyRawPublic = req.PublicKey

		publicKeyRawToKeyRes = cert.PublicKeyRawToKey(publicKeyRawToKeyReq)
		if publicKeyRawToKeyRes.Error != nil {
			res.Error = publicKeyRawToKeyRes.Error
			return
		}

		break

	default:
		res.Error = errors.New("Invalid VerifyAsymmetric PEM PublicKey certificate unsupported")
		return
	}

	decodeSignature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		res.Error = errors.New("Invalid signature")
		return
	}

	err = rsa.VerifyPKCS1v15(publicKeyRawToKeyRes.KeyPublic, cpt.SHA256, cipherBodyHash, decodeSignature)
	if err != nil {
		res.Error = errors.New("Invalid signature")
		return
	}

	res.Signature = req.Signature
	return
}

func (p signature) VerifySymmetric(req sdk_dto.VerifySymetric) (res sdk_opt.SignatureResponse) {
	cert := NewCert()
	cipherBodyHash256 := sha256.New()

	if _, err := cipherBodyHash256.Write(req.Body); err != nil {
		res.Error = err
		return
	}

	cipherBodyHash := cipherBodyHash256.Sum(nil)
	publicKeyRawToKeyReq := sdk_dto.PublicKeyRawToKey{}
	publicKeyRawToKeyRes := sdk_opt.CertResponse{}

	switch req.PublicKeyType {

	case sdk_cons.PRIVPKCS1:
		publicKeyRawToKeyReq.KeyType = req.PublicKeyType
		publicKeyRawToKeyReq.KeyRawPublic = req.PublicKey

		publicKeyRawToKeyRes = cert.PublicKeyRawToKey(publicKeyRawToKeyReq)
		if publicKeyRawToKeyRes.Error != nil {
			res.Error = publicKeyRawToKeyRes.Error
			return
		}

		break

	case sdk_cons.PRIVPKCS8:
		publicKeyRawToKeyReq.KeyType = req.PublicKeyType
		publicKeyRawToKeyReq.KeyRawPublic = req.PublicKey

		publicKeyRawToKeyRes = cert.PublicKeyRawToKey(publicKeyRawToKeyReq)
		if publicKeyRawToKeyRes.Error != nil {
			res.Error = publicKeyRawToKeyRes.Error
			return
		}

		break

	default:
		res.Error = errors.New("Invalid VerifySymmetric PEM PublicKey certificate unsupported")
		return
	}

	decodeSignature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		res.Error = errors.New("Invalid signature")
		return
	}

	err = rsa.VerifyPKCS1v15(publicKeyRawToKeyRes.KeyPublic, cpt.SHA256, cipherBodyHash, decodeSignature)
	if err != nil {
		res.Error = errors.New("Invalid signature")
		return
	}

	sha256SecretKey := strings.ToLower(hex.EncodeToString(cipherBodyHash))

	hmac512Body := req.Method + ":" + req.Url + ":" + req.AccessToken + ":" + sha256SecretKey + ":" + req.TimeStamp
	hmac512 := hmac.New(cpt.SHA512.New, []byte(req.ClientSecret))

	if _, err := hmac512.Write([]byte(strings.TrimSpace(hmac512Body))); err != nil {
		res.Error = err
		return
	}

	res.Signature = base64.StdEncoding.EncodeToString(hmac512.Sum(nil))

	if ok := reflect.DeepEqual(req.Signature, res.Signature); !ok {
		res.Error = errors.New("Invalid signature")
		return
	}

	res.Signature = req.Signature
	return
}
