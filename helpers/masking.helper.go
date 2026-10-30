package sdk_helper

import (
	"bytes"
	"regexp"
	"strings"
)

const (
	MaxLogBodyLength = 4096
	RedactedValue    = "[REDACTED]"
)

var (
	rxSensitiveKey = regexp.MustCompile(`(?i)"(pin|password|secret|card_cvv|access_token)"\s*:\s*("[^"]*"|[0-9]+)`)
	rxBase64Data   = regexp.MustCompile(`"data:image/[^;]+;base64,[^"]*"`)
)

// SanitizeHeaders returns a sanitized copy of HTTP headers with sensitive tokens redacted.
// It does not mutate the input map.
func SanitizeHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return headers
	}

	sanitized := make(map[string]string, len(headers))
	for k, v := range headers {
		lower := strings.ToLower(strings.TrimSpace(k))
		switch lower {
		case "authorization", "signature", "x-signature", "x-api-key", "client-secret", "x-client-secret":
			sanitized[k] = RedactedValue
		default:
			sanitized[k] = v
		}
	}

	return sanitized
}

// SanitizeBody truncates large payloads and masks sensitive fields with zero heap allocations on clean traffic.
func SanitizeBody(bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return ""
	}

	truncated := false
	if len(bodyBytes) > MaxLogBodyLength {
		bodyBytes = bodyBytes[:MaxLogBodyLength]
		truncated = true
	}

	// Zero-Alloc Short-Circuit: Skip regex if no sensitive keywords or image payloads exist
	hasPin := bytes.Contains(bodyBytes, []byte(`"pin"`)) || bytes.Contains(bodyBytes, []byte(`"PIN"`))
	hasPassword := bytes.Contains(bodyBytes, []byte(`"password"`)) || bytes.Contains(bodyBytes, []byte(`"PASSWORD"`))
	hasSecret := bytes.Contains(bodyBytes, []byte(`"secret"`)) || bytes.Contains(bodyBytes, []byte(`"SECRET"`))
	hasToken := bytes.Contains(bodyBytes, []byte(`"access_token"`)) || bytes.Contains(bodyBytes, []byte(`"token"`))
	hasCardCvv := bytes.Contains(bodyBytes, []byte(`"card_cvv"`))
	hasBase64Image := bytes.Contains(bodyBytes, []byte(`data:image`))

	if !hasPin && !hasPassword && !hasSecret && !hasToken && !hasCardCvv && !hasBase64Image {
		if truncated {
			return string(bodyBytes) + TruncatedSuffix
		}
		return string(bodyBytes)
	}

	res := bodyBytes
	if hasPin || hasPassword || hasSecret || hasToken || hasCardCvv {
		res = rxSensitiveKey.ReplaceAll(res, []byte(`"${1}":"`+RedactedValue+`"`))
	}

	if hasBase64Image {
		res = rxBase64Data.ReplaceAll(res, []byte(`"data:image/png;base64,`+TruncatedSuffix+`"`))
	}

	if truncated {
		return string(res) + TruncatedSuffix
	}

	return string(res)
}
