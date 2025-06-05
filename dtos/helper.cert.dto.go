package sdk_dto

import "crypto/rsa"

type (
	CertRequest struct {
		KeyType       string          `json:"key_type,omitempty"`
		KeyRawPrivate []byte          `json:"key_raw_private,omitempty"`
		KeyPrivate    *rsa.PrivateKey `json:"key_private,omitempty"`
		KeyRawPublic  []byte          `json:"key_raw_public,omitempty"`
		KeyPublic     *rsa.PublicKey  `json:"key_public,omitempty"`
		Password      string          `json:"password,omitempty"`
	}

	GeneratePrivateKey struct {
		PrivateKeyType string `json:"private_key_type"`
		PublicKeyType  string `json:"public_key_type"`
		KeySize        uint   `json:"key_size"`
		Password       string `json:"password"`
	}

	PrivateKeyRawToKey struct {
		KeyType       string `json:"key_type,omitempty"`
		KeyRawPrivate []byte `json:"key_raw_private,omitempty"`
		Password      string `json:"password,omitempty"`
	}

	PublicKeyRawToKey struct {
		KeyType      string `json:"key_type,omitempty"`
		KeyRawPublic []byte `json:"key_raw_public,omitempty"`
	}

	PrivateKeyToRaw struct {
		KeyType    string          `json:"key_type,omitempty"`
		KeyPrivate *rsa.PrivateKey `json:"key_private,omitempty"`
	}

	PublicKeyToRaw struct {
		KeyType   string         `json:"key_type,omitempty"`
		KeyPublic *rsa.PublicKey `json:"key_public,omitempty"`
	}

	PrivateKeyBase64ToRaw struct {
		KeyRawPrivate string `json:"key_raw_private,omitempty"`
	}

	PublicKeyBase64ToRaw struct {
		KeyRawPublic string `json:"key_raw_public,omitempty"`
	}
)
