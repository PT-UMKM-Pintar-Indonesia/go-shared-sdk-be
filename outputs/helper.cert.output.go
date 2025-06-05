package sdk_opt

import "crypto/rsa"

type (
	CertResponse struct {
		KeyType       string          `json:"key_type,omitempty"`
		KeyRawPrivate []byte          `json:"key_raw_private,omitempty"`
		KeyPrivate    *rsa.PrivateKey `json:"key_private,omitempty"`
		KeyRawPublic  []byte          `json:"key_raw_public,omitempty"`
		KeyPublic     *rsa.PublicKey  `json:"key_public,omitempty"`
		Error         error           `json:"error,omitempty"`
	}
)
