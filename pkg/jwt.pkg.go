package sdk_pkg

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	sdk_dro "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type jsonWebToken struct {
	ctx       context.Context
	env       *sdk_dto.Environment
	rds       sdk_inf.IRedis
	jose      sdk_inf.IJose
	cipher    sdk_inf.ICrypto
	cert      sdk_inf.ICert
	parser    sdk_inf.IParser
	transform sdk_inf.ITransform
	mu        sync.RWMutex
}

func NewJsonWebToken(ctx context.Context, env *sdk_dto.Environment, rds sdk_inf.IRedis) sdk_inf.IJsonWebToken {
	return &jsonWebToken{
		ctx:       ctx,
		env:       env,
		rds:       rds,
		jose:      NewJose(ctx),
		cipher:    sdk_helper.NewCrypto(),
		cert:      sdk_helper.NewCert(),
		parser:    sdk_helper.NewParser(),
		transform: sdk_helper.NewTransform(),
	}
}

func (p *jsonWebToken) IsRedisAvailable() bool {
	return p.rds != nil
}

func (p *jsonWebToken) createSecret(prefix string, body []byte) (*sdk_dro.SecretMetadata, error) {
	timeNow := time.Now().Format(time.UnixDate)
	cipherTextRandom := fmt.Sprintf("%s:%s:%s:%d", prefix, string(body), timeNow, p.env.JWT.EXPIRED)
	cipherTextData := hex.EncodeToString([]byte(cipherTextRandom))

	cipherSecretKey, err := p.cipher.SHA512Sign(cipherTextData)
	if err != nil {
		return nil, fmt.Errorf("sha512 sign failed: %w", err)
	}

	cipherText, err := p.cipher.SHA512Sign(timeNow)
	if err != nil {
		return nil, fmt.Errorf("sha512 sign time failed: %w", err)
	}

	cipherKey, err := p.cipher.AES256Encrypt(cipherSecretKey, cipherText)
	if err != nil {
		return nil, fmt.Errorf("aes encrypt failed: %w", err)
	}

	privateKey := p.cert.GenerateKey(&sdk_dto.GeneratePrivateKey{
		Password:       cipherKey,
		KeySize:        sdk_cons.KEY_SIZE_4096,
		PrivateKeyType: sdk_cons.PRIVPKCS8,
		PublicKeyType:  sdk_cons.PUBPKCS8,
	})
	if privateKey.Error != nil {
		return nil, privateKey.Error
	}

	return &sdk_dro.SecretMetadata{
		PrivKeyRaw: string(privateKey.KeyRawPrivate),
		CipherKey:  cipherKey,
	}, nil
}

func (p *jsonWebToken) createSignature(prefix string, body any) (*sdk_dro.SignatureMetadata, error) {
	bodyByte, err := p.parser.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("parser marshal failed: %w", err)
	}

	secretKey, err := p.createSecret(prefix, bodyByte)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey := p.cert.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyRawPrivate: []byte(secretKey.PrivKeyRaw),
		Password:      secretKey.CipherKey,
	})
	if rsaPrivateKey.Error != nil {
		return nil, rsaPrivateKey.Error
	}

	cipherHash512 := sha512.New()
	cipherHash512.Write(bodyByte)
	cipherHash512Body := cipherHash512.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey.KeyPrivate, crypto.SHA512, cipherHash512Body)
	if err != nil {
		return nil, fmt.Errorf("rsa sign failed: %w", err)
	}

	if err := rsa.VerifyPKCS1v15(rsaPrivateKey.KeyPublic, crypto.SHA512, cipherHash512Body, signature); err != nil {
		return nil, fmt.Errorf("rsa verify failed: %w", err)
	}

	signatureOutput := hex.EncodeToString(signature)
	_, jweKey, err := p.jose.JweEncrypt(rsaPrivateKey.KeyPublic, signatureOutput)
	if err != nil {
		return nil, err
	}

	signatureMetadata := &sdk_dro.SignatureMetadata{
		PrivKeyRaw: secretKey.PrivKeyRaw,
		SigKey:     signatureOutput,
		CipherKey:  secretKey.CipherKey,
		JweKey:     *jweKey,
		PrivKey:    rsaPrivateKey.KeyPrivate,
	}

	if p.rds != nil {
		jwtMetadataByte, err := p.parser.Marshal(signatureMetadata)
		if err != nil {
			return nil, fmt.Errorf("parser marshal failed: %w", err)
		}

		jwtExpired := time.Duration(time.Minute * time.Duration(p.env.JWT.EXPIRED))
		if err := p.rds.HSetEx(p.ctx, fmt.Sprintf("jwt:cert:%s", prefix), jwtExpired, "signature_metadata", jwtMetadataByte); err != nil {
			return nil, fmt.Errorf("redis hsetex failed: %w", err)
		}
	}

	return signatureMetadata, nil
}

func (p *jsonWebToken) Sign(prefix string, body any) (*sdk_dro.SignMetadata, error) {
	tokenKey := fmt.Sprintf("jwt:token:%s", prefix)

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.rds != nil {
		tokenExist, err := p.rds.Exists(p.ctx, tokenKey)
		if err != nil {
			return nil, fmt.Errorf("redis exists failed: %w", err)
		}

		if tokenExist > 0 {
			tokenData, err := p.rds.Get(p.ctx, tokenKey)
			if err != nil {
				return nil, fmt.Errorf("redis get failed: %w", err)
			}
			return &sdk_dro.SignMetadata{Token: string(tokenData), Expired: p.env.JWT.EXPIRED}, nil
		}
	}

	signature, err := p.createSignature(prefix, body)
	if err != nil {
		return nil, err
	}

	if len(signature.SigKey) < 60 {
		return nil, fmt.Errorf("invalid signature key length: expected >= 60, got %d", len(signature.SigKey))
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
		return nil, fmt.Errorf("aes encrypt failed: %w", err)
	}

	duration := time.Duration(time.Minute * time.Duration(p.env.JWT.EXPIRED))

	tokenData, err := p.jose.JwtSign(&sdk_dto.JwtSignOption{
		SecretKey:  signature.CipherKey,
		Kid:        signature.JweKey.CipherText,
		PrivateKey: signature.PrivKey,
		Aud:        []string{aud},
		Iss:        iss,
		Sub:        sub,
		Jti:        jti,
		Iat:        time.Now().UTC().Add(-duration),
		Exp:        time.Now().Add(duration),
		Claim:      timestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("jwt sign failed: %w", err)
	}

	if p.rds != nil {
		if err := p.rds.SetEx(p.ctx, tokenKey, duration, string(tokenData)); err != nil {
			return nil, fmt.Errorf("redis setex token failed: %w", err)
		}

		bodyMarshal, err := p.parser.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("redis setex body failed: %w", err)
		}

		if err := p.rds.SetEx(p.ctx, fmt.Sprintf("jwt:body:%s", prefix), duration, string(bodyMarshal)); err != nil {
			return nil, fmt.Errorf("redis setex body failed: %w", err)
		}
	}

	return &sdk_dro.SignMetadata{Token: string(tokenData), Expired: p.env.JWT.EXPIRED}, nil
}

func (p *jsonWebToken) Verify(prefix string, token string) (jwt.Token, error) {
	return p.jose.JwtVerify(prefix, token, p.rds)
}
