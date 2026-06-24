package sdk_dro

type (
	SignatureResponse struct {
		Signature string `json:"signature"`
		Error     error  `json:"error,omitempty"`
	}

	TokenResponse struct {
		Token   string `json:"token"`
		Expired int    `json:"expired"`
		Error   error  `json:"error,omitempty"`
	}
)
