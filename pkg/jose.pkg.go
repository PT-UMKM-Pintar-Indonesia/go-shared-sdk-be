package sdk_pkg

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	sdk_dro "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type jose struct {
	ctx       context.Context
	cert      sdk_inf.ICert
	parser    sdk_inf.IParser
	transform sdk_inf.ITransform
}

func NewJose(ctx context.Context) sdk_inf.IJose {
	jwk.Configure(jwk.WithStrictKeyUsage(true))
	return &jose{
		ctx:       ctx,
		cert:      sdk_helper.NewCert(),
		parser:    sdk_helper.NewParser(),
		transform: sdk_helper.NewTransform(),
	}
}

func (p *jose) JweEncrypt(publicKey *rsa.PublicKey, plainText string) ([]byte, *sdk_dro.JweEncryptMetadata, error) {
	jweEncryptMetadataReq := new(sdk_dto.JweEncryptMetadata)
	jweEncryptMetadataRes := new(sdk_dro.JweEncryptMetadata)

	headers := jwe.NewHeaders()
	headers.Set("sig", plainText)
	headers.Set("alg", jwa.RSA_OAEP_512().String())
	headers.Set("enc", jwa.A256GCM().String())

	cipherText, err := jwe.Encrypt([]byte(plainText),
		jwe.WithKey(jwa.RSA_OAEP_512(), publicKey),
		jwe.WithContentEncryption(jwa.A256GCM()),
		jwe.WithJSON(),
		jwe.WithProtectedHeaders(headers),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("jwe encrypt failed: %w", err)
	}

	if err := p.parser.Unmarshal(cipherText, jweEncryptMetadataReq); err != nil {
		return nil, nil, fmt.Errorf("parser unmarshal failed: %w", err)
	}

	if err := p.transform.SrcToDest(jweEncryptMetadataReq, jweEncryptMetadataRes); err != nil {
		return nil, nil, fmt.Errorf("transform failed: %w", err)
	}

	return cipherText, jweEncryptMetadataRes, nil
}

func (p *jose) JweDecrypt(privateKey *rsa.PrivateKey, cipherText []byte) (string, error) {
	jwtKey, err := jwk.Import(privateKey)
	if err != nil {
		return "", fmt.Errorf("jwk import failed: %w", err)
	}

	jwkSet := jwk.NewSet()
	if err := jwkSet.AddKey(jwtKey); err != nil {
		return "", fmt.Errorf("jwk add key failed: %w", err)
	}

	plainText, err := jwe.Decrypt(cipherText,
		jwe.WithKey(jwa.RSA_OAEP_512(), jwtKey),
		jwe.WithKeySet(jwkSet, jwe.WithRequireKid(false)),
	)
	if err != nil {
		return "", fmt.Errorf("jwe decrypt failed: %w", err)
	}

	return string(plainText), nil
}

func (p *jose) ImportJsonWebKey(jwkKey jwk.Key) (*sdk_dro.JwkMetadata, error) {
	jwkRawMetadataReq := sdk_dto.JwkMetadata{}
	jwkRawMetadataRes := sdk_dro.JwkMetadata{}

	if _, err := jwk.IsPrivateKey(jwkKey); err != nil {
		return nil, err
	}

	if err := jwk.AssignKeyID(jwkKey); err != nil {
		return nil, err
	}

	jwkKeyByte, err := p.parser.Marshal(&jwkKey)
	if err != nil {
		return nil, fmt.Errorf("parser marshal failed: %w", err)
	}

	jwkRaw, err := jwk.ParseKey(jwkKeyByte)
	if err != nil {
		return nil, fmt.Errorf("jwk parse failed: %w", err)
	}

	if err := p.parser.Unmarshal(jwkKeyByte, &jwkRawMetadataReq.KeyRaw); err != nil {
		return nil, fmt.Errorf("parser unmarshal failed: %w", err)
	}

	if err := p.transform.SrcToDest(&jwkRawMetadataReq, &jwkRawMetadataRes); err != nil {
		return nil, fmt.Errorf("transform failed: %w", err)
	}

	jwkRawMetadataRes.Key = jwkRaw

	return &jwkRawMetadataRes, nil
}

func (p *jose) ExportJsonWebKey(privateKey *rsa.PrivateKey) (*sdk_dro.JwkMetadata, error) {
	jwkRawMetadataReq := sdk_dto.JwkMetadata{}
	jwkRawMetadataRes := sdk_dro.JwkMetadata{}

	privateKeyRawToKey := &sdk_dto.PrivateKeyToRaw{KeyPrivate: privateKey}
	rsaPrivateKey := p.cert.PrivateKeyToRaw(privateKeyRawToKey)
	if rsaPrivateKey.Error != nil {
		return nil, rsaPrivateKey.Error
	}

	jwkRaw, err := jwk.ParseKey([]byte(rsaPrivateKey.KeyRawPrivate), jwk.WithPEM(true))
	if err != nil {
		return nil, fmt.Errorf("jwk parse failed: %w", err)
	}

	jwkRawByte, err := p.parser.Marshal(&jwkRaw)
	if err != nil {
		return nil, fmt.Errorf("parser marshal failed: %w", err)
	}

	if err := p.parser.Unmarshal(jwkRawByte, &jwkRawMetadataReq.KeyRaw); err != nil {
		return nil, fmt.Errorf("parser unmarshal failed: %w", err)
	}

	if err := p.transform.SrcToDest(&jwkRawMetadataReq, &jwkRawMetadataRes); err != nil {
		return nil, fmt.Errorf("transform failed: %w", err)
	}

	jwkRawMetadataRes.Key = jwkRaw.(jwk.Key)

	return &jwkRawMetadataRes, nil
}

func (p *jose) JwtSign(opt *sdk_dto.JwtSignOption) ([]byte, error) {
	jwsHeader := jws.NewHeaders()
	jwsHeader.Set("alg", jwa.RS512)
	jwsHeader.Set("typ", "JWT")
	jwsHeader.Set("cty", "JWT")
	jwsHeader.Set("kid", opt.Kid)
	jwsHeader.Set("b64", true)

	jwtBuilder := jwt.NewBuilder()
	jwtBuilder.Audience(opt.Aud)
	jwtBuilder.Issuer(opt.Iss)
	jwtBuilder.Subject(opt.Sub)
	jwtBuilder.IssuedAt(opt.Iat)
	jwtBuilder.Expiration(opt.Exp)
	jwtBuilder.JwtID(opt.Jti)
	jwtBuilder.Claim("timestamp", opt.Claim)

	jwtToken, err := jwtBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("jwt build failed: %w", err)
	}

	token, err := jwt.Sign(jwtToken, jwt.WithKey(jwa.RS512(), opt.PrivateKey, jws.WithProtectedHeaders(jwsHeader)))
	if err != nil {
		return nil, fmt.Errorf("jwt sign failed: %w", err)
	}

	return token, nil
}

func (p *jose) JwtVerify(prefix string, token string, redis sdk_inf.IRedis) (jwt.Token, error) {
	signatureKey := fmt.Sprintf("jwt:cert:%s", prefix)
	signatureMetadataField := "signature_metadata"

	signatureMetadataBytes, err := redis.HGet(p.ctx, signatureKey, signatureMetadataField)
	if err != nil {
		return nil, fmt.Errorf("redis hget failed: %w", err)
	}

	var signatureMetadata sdk_dto.SignatureMetadata
	if err := p.parser.Unmarshal(signatureMetadataBytes, &signatureMetadata); err != nil {
		return nil, fmt.Errorf("parser unmarshal failed: %w", err)
	}

	if reflect.DeepEqual(signatureMetadata, sdk_dto.SignatureMetadata{}) {
		return nil, errors.New("invalid secretkey or signature")
	}

	privateKey := p.cert.PrivateKeyRawToKey(&sdk_dto.PrivateKeyRawToKey{
		KeyRawPrivate: []byte(signatureMetadata.PrivKeyRaw),
		Password:      signatureMetadata.CipherKey,
	})
	if privateKey.Error != nil {
		return nil, privateKey.Error
	}

	exportJws, err := jws.ParseString(token)
	if err != nil {
		return nil, fmt.Errorf("jws parse failed: %w", err)
	}

	signatures := exportJws.Signatures()
	if len(signatures) < 1 {
		return nil, errors.New("invalid signature")
	}

	jwsHeaders := signatures[0].ProtectedHeaders()
	algorithm, ok := jwsHeaders.Algorithm()
	if !ok || algorithm != jwa.RS512() {
		return nil, errors.New("invalid algorithm")
	}

	kid, ok := jwsHeaders.KeyID()
	if !ok || kid != signatureMetadata.JweKey.CipherText {
		return nil, errors.New("invalid keyid")
	}

	if len(signatureMetadata.SigKey) < 60 {
		return nil, fmt.Errorf("invalid signature key length: expected >= 60, got %d", len(signatureMetadata.SigKey))
	}

	jwkKey, err := jwk.Import(privateKey.KeyPrivate)
	if err != nil {
		return nil, fmt.Errorf("jwk import failed: %w", err)
	}

	_, err = jws.Verify([]byte(token), jws.WithKey(algorithm, jwkKey), jws.WithMessage(exportJws))
	if err != nil {
		return nil, fmt.Errorf("jws verify failed: %w", err)
	}

	return jwt.Parse([]byte(token),
		jwt.WithKey(algorithm, privateKey.KeyPrivate),
		jwt.WithAudience(signatureMetadata.SigKey[10:20]),
		jwt.WithIssuer(signatureMetadata.SigKey[30:40]),
		jwt.WithSubject(signatureMetadata.SigKey[50:60]),
		jwt.WithRequiredClaim("timestamp"),
	)
}
