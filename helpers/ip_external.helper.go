package sdk_helper

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
)

var (
	ipExternalTimeout    = 5 * time.Second
	ipExternalMaxResp    = int64(64 * 1024)
	ipExternalServiceURL = "https://api.ipify.org"
)

func IPExternal() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ipExternalTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipExternalServiceURL, nil)
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return sdk_cons.EMPTY, err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(io.LimitReader(resp.Body, ipExternalMaxResp))
	if err != nil {
		return sdk_cons.EMPTY, err
	}

	return strings.TrimSpace(string(ip)), nil
}
