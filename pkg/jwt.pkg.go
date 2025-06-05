package pkg

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/interfaces"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/outputs"
	goredis "github.com/redis/go-redis/v9"
)

type jsonWebToken struct {
	env       *sdk_dto.Environtment
	rds       sdk_inf.IRedis
	jose      sdk_inf.IJose
	cipher    sdk_inf.ICrypto
	cert      sdk_inf.ICert
	parser    sdk_inf.IParser
	transform sdk_inf.ITransform
}

func NewJsonWebToken(ctx context.Context, env *sdk_dto.Environtment, con *goredis.Client) sdk_inf.IJsonWebToken {
	jose := NewJose(ctx)

	rds, err := NewRedis(ctx, con)
	if err != nil {
		Logrus(sdk_cons.FATAL, err)
	}

	cipher := sdk_helper.NewCrypto()
	cert := sdk_helper.NewCert()
	parser := sdk_helper.NewParser()
	transform := sdk_helper.NewTransform()

	return jsonWebToken{
		env:       env,
		rds:       rds,
		jose:      jose,
		cipher:    cipher,
		cert:      cert,
		parser:    parser,
		transform: transform,
	}
}

func (p jsonWebToken) createSecret(prefix string, body []byte) (*sdk_opt.SecretMetadata, error) {
	secretMetadataReq := new(sdk_dto.SecretMetadata)
	secretMetadataRes := new(sdk_opt.SecretMetadata)

	timeNow := time.Now().Format(time.UnixDate)
	cipherTextRandom := fmt.Sprintf("%s:%s:%s:%d", prefix, string(body), timeNow, p.env.JWT.EXPIRED)
	cipherTextData := hex.EncodeToString([]byte(cipherTextRandom))

	cipherSecretKey, err := p.cipher.SHA512Sign(cipherTextData)
	if err != nil {
		return nil, err
	}

	cipherText, err := p.cipher.SHA512Sign(timeNow)
	if err != nil {
		return nil, err
	}

	cipherKey, err := p.cipher.AES256Encrypt(cipherSecretKey, cipherText)
	if err != nil {
		return nil, err
	}

	generatePrivateKey := sdk_dto.GeneratePrivateKey{}
	generatePrivateKey.Password = cipherKey

	privateKey := p.cert.GenerateKey(generatePrivateKey)
	if privateKey.Error != nil {
		return nil, err
	}

	secretMetadataReq.PrivKeyRaw = string(privateKey.KeyRawPrivate)
	secretMetadataReq.CipherKey = cipherKey

	if err := p.transform.ReqToRes(secretMetadataReq, secretMetadataRes); err != nil {
		return nil, err
	}

	return secretMetadataRes, nil
}

func (p jsonWebToken) createSignature(prefix string, body any) (*sdk_opt.SignatureMetadata, error) {
	var (
		signatureMetadataReq *sdk_dto.SignatureMetadata = new(sdk_dto.SignatureMetadata)
		signatureMetadataRes *sdk_opt.SignatureMetadata = new(sdk_opt.SignatureMetadata)
		signatureKey         string                     = fmt.Sprintf("CREDENTIAL:%s", prefix)
		signatureField       string                     = "signature_metadata"
	)

	bodyByte, err := p.parser.Marshal(body)
	if err != nil {
		return nil, err
	}

	secretKey, err := p.createSecret(prefix, bodyByte)
	if err != nil {
		return nil, err
	}

	privateKeyRawToKey := sdk_dto.PrivateKeyRawToKey{}
	privateKeyRawToKey.KeyRawPrivate = []byte(secretKey.PrivKeyRaw)
	privateKeyRawToKey.Password = secretKey.CipherKey

	rsaPrivateKey := p.cert.PrivateKeyRawToKey(privateKeyRawToKey)
	if rsaPrivateKey.Error != nil {
		return nil, err
	}

	cipherHash512 := sha512.New()
	cipherHash512.Write(bodyByte)
	cipherHash512Body := cipherHash512.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey.KeyPrivate, crypto.SHA512, cipherHash512Body)
	if err != nil {
		return nil, err
	}

	if err := rsa.VerifyPKCS1v15(rsaPrivateKey.KeyPublic, crypto.SHA512, cipherHash512Body, signature); err != nil {
		return nil, err
	}

	signatureOutput := hex.EncodeToString(signature)

	_, jweKey, err := p.jose.JweEncrypt(rsaPrivateKey.KeyPublic, signatureOutput)
	if err != nil {
		return nil, err
	}

	signatureMetadataReq.PrivKeyRaw = secretKey.PrivKeyRaw
	signatureMetadataReq.SigKey = signatureOutput
	signatureMetadataReq.CipherKey = secretKey.CipherKey
	signatureMetadataReq.JweKey = *jweKey
	signatureMetadataReq.PrivKey = rsaPrivateKey.KeyPrivate

	signatureMetadataByte, err := p.parser.Marshal(signatureMetadataReq)
	if err != nil {
		return nil, err
	}

	jwtClaim := string(signatureMetadataByte)
	jwtExpired := time.Duration(time.Minute * time.Duration(p.env.JWT.EXPIRED))

	if err := p.rds.HSetEx(signatureKey, jwtExpired, signatureField, jwtClaim); err != nil {
		return nil, err
	}

	if err := p.transform.ReqToRes(signatureMetadataReq, signatureMetadataRes); err != nil {
		return nil, err
	}

	return signatureMetadataRes, nil
}

func (p jsonWebToken) Sign(prefix string, body any) (*sdk_opt.SignMetadata, error) {
	tokenKey := fmt.Sprintf("TOKEN:%s", prefix)
	signMetadataRes := new(sdk_opt.SignMetadata)

	tokenExist, err := p.rds.Exists(tokenKey)
	if err != nil {
		return nil, err
	}

	if tokenExist < 1 {
		signature, err := p.createSignature(prefix, body)
		if err != nil {
			return nil, err
		}

		timestamp := time.Now().Format(sdk_cons.DATE_TIME_FORMAT)
		aud := signature.SigKey[10:20]
		iss := signature.SigKey[30:40]
		sub := signature.SigKey[50:60]
		suffix := int(math.Pow(float64(p.env.JWT.EXPIRED), float64(len(aud)+len(iss)+len(sub))))

		secretKey := fmt.Sprintf("%s:%s:%s:%s:%d", aud, iss, sub, timestamp, suffix)
		secretData := hex.EncodeToString([]byte(secretKey))

		jti, err := p.cipher.AES256Encrypt(secretData, prefix)
		if err != nil {
			return nil, err
		}

		duration := time.Duration(time.Minute * time.Duration(p.env.JWT.EXPIRED))
		jwtIat := time.Now().UTC().Add(-duration)
		jwtExp := time.Now().Add(duration)

		tokenPayload := new(sdk_dto.JwtSignOption)
		tokenPayload.SecretKey = signature.CipherKey
		tokenPayload.Kid = signature.JweKey.CipherText
		tokenPayload.PrivateKey = signature.PrivKey
		tokenPayload.Aud = []string{aud}
		tokenPayload.Iss = iss
		tokenPayload.Sub = sub
		tokenPayload.Jti = jti
		tokenPayload.Iat = jwtIat
		tokenPayload.Exp = jwtExp
		tokenPayload.Claim = timestamp

		tokenData, err := p.jose.JwtSign(tokenPayload)
		if err != nil {
			return nil, err
		}

		if err := p.rds.SetEx(tokenKey, duration, string(tokenData)); err != nil {
			return nil, err
		}

		signMetadataRes.Token = string(tokenData)
		signMetadataRes.Expired = p.env.JWT.EXPIRED

		return signMetadataRes, nil
	} else {
		tokenData, err := p.rds.Get(tokenKey)
		if err != nil {
			return nil, err
		}

		signMetadataRes.Token = string(tokenData)
		signMetadataRes.Expired = p.env.JWT.EXPIRED

		return signMetadataRes, nil
	}
}
