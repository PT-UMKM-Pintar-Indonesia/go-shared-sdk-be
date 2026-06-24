package sdk_helper

import (
	"io"
	"net/http"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
)

func IPExternal() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return sdk_cons.EMPTY, err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	return string(ip), nil
}
