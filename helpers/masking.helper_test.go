package sdk_helper

import (
	"strings"
	"testing"
)

func TestSanitizeHeaders(t *testing.T) {
	orig := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer secret-token-123",
		"X-SIGNATURE":   "signature-abc-999",
		"x-api-key":     "api-key-xyz",
		"User-Agent":    "UMKMS-Client/1.0",
		"Client-Secret": "client-secret-pass",
	}

	sanitized := SanitizeHeaders(orig)

	// Check original was not mutated
	if orig["Authorization"] != "Bearer secret-token-123" {
		t.Fatalf("SanitizeHeaders mutated original map: %v", orig["Authorization"])
	}

	// Check redacted keys
	if sanitized["Authorization"] != RedactedValue {
		t.Fatalf("expected Authorization redacted, got %s", sanitized["Authorization"])
	}
	if sanitized["X-SIGNATURE"] != RedactedValue {
		t.Fatalf("expected X-SIGNATURE redacted, got %s", sanitized["X-SIGNATURE"])
	}
	if sanitized["x-api-key"] != RedactedValue {
		t.Fatalf("expected x-api-key redacted, got %s", sanitized["x-api-key"])
	}
	if sanitized["Client-Secret"] != RedactedValue {
		t.Fatalf("expected Client-Secret redacted, got %s", sanitized["Client-Secret"])
	}

	// Check non-redacted keys preserved
	if sanitized["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type preserved, got %s", sanitized["Content-Type"])
	}
	if sanitized["User-Agent"] != "UMKMS-Client/1.0" {
		t.Fatalf("expected User-Agent preserved, got %s", sanitized["User-Agent"])
	}
}

func TestSanitizeBody_CleanTraffic(t *testing.T) {
	clean := []byte(`{"account":"123456","amount":50000,"bank":"bca"}`)
	out := SanitizeBody(clean)
	if out != string(clean) {
		t.Fatalf("expected %s, got %s", string(clean), out)
	}
}

func TestSanitizeBody_SensitiveFields(t *testing.T) {
	input := []byte(`{"pin":"123456","PASSWORD":"my-password","secret":8888,"access_token":"token-abc","card_cvv":"123","bank":"bca"}`)
	out := SanitizeBody(input)

	if strings.Contains(out, "123456") {
		t.Fatalf("PIN was not redacted: %s", out)
	}
	if strings.Contains(out, "my-password") {
		t.Fatalf("PASSWORD was not redacted: %s", out)
	}
	if strings.Contains(out, "8888") {
		t.Fatalf("secret was not redacted: %s", out)
	}
	if strings.Contains(out, "token-abc") {
		t.Fatalf("access_token was not redacted: %s", out)
	}
	if !strings.Contains(out, `"[REDACTED]"`) {
		t.Fatalf("expected [REDACTED] in output: %s", out)
	}
	if !strings.Contains(out, `"bank":"bca"`) {
		t.Fatalf("expected non-sensitive field bank preserved: %s", out)
	}
}

func TestSanitizeBody_Base64Image(t *testing.T) {
	input := []byte(`{"qr_image":"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==","amount":10000}`)
	out := SanitizeBody(input)

	if strings.Contains(out, "iVBORw0KGgo") {
		t.Fatalf("raw base64 payload should have been truncated: %s", out)
	}
	if !strings.Contains(out, `data:image/png;base64,...[truncated]`) {
		t.Fatalf("expected base64 truncation format, got: %s", out)
	}
	if !strings.Contains(out, `"amount":10000`) {
		t.Fatalf("expected amount preserved, got: %s", out)
	}
}

func TestSanitizeBody_ExceedsMaxLogLength(t *testing.T) {
	large := strings.Repeat("A", MaxLogBodyLength+500)
	out := SanitizeBody([]byte(large))

	if !strings.HasSuffix(out, TruncatedSuffix) {
		t.Fatalf("expected truncated suffix at end, got: %s", out[len(out)-30:])
	}
	if len(out) > MaxLogBodyLength+len(TruncatedSuffix) {
		t.Fatalf("length exceeded maximum allowed: %d", len(out))
	}
}
