package sdk_dto

type (
	Asymmetric struct {
		PrivateKeyType string `json:"private_key_type"`
		PrivateKey     []byte `json:"private_key"`
		PublicKey      string `json:"public_key"`
		TimeStamp      string `json:"time_stamp"`
		ClientKey      string `json:"client_key"`
		Password       string `json:"password,omitempty"`
		Encoding       string `json:"encoding,omitempty"`
	}

	Symetric struct {
		Url          string `json:"url"`
		Method       string `json:"method"`
		AccessToken  string `json:"access_token,omitempty"`
		TimeStamp    string `json:"time_stamp,omitempty"`
		ClientSecret string `json:"client_secret"`
		Body         []byte `json:"body"`
		Encoding     string `json:"encoding,omitempty"`
	}

	VerifyAsymmetric struct {
		PublicKeyType string `json:"public_key_type,omitempty"`
		PublicKey     []byte `json:"public_key,omitempty"`
		Signature     string `json:"signature,omitempty"`
		ClientId      string `json:"client_id,omitempty"`
		Timestamp     string `json:"timestamp,omitempty"`
		Encoding      string `json:"encoding,omitempty"`
	}

	VerifySymetric struct {
		PublicKeyType string `json:"public_key_type,omitempty"`
		PublicKey     []byte `json:"public_key,omitempty"`
		Signature     string `json:"signature"`
		Url           string `json:"url"`
		Method        string `json:"method"`
		AccessToken   string `json:"access_token,omitempty"`
		TimeStamp     string `json:"time_stamp,omitempty"`
		ClientSecret  string `json:"client_secret"`
		Body          []byte `json:"body"`
		Encoding      string `json:"encoding,omitempty"`
	}

	Sign struct {
		Claim     []byte `json:"claim"`
		SecretKey string `json:"secret_key"`
		Expired   int    `json:"expired"`
		ClientID  string `json:"client_id"`
		Signature string `json:"signature"`
	}
)
