package sdk_helper

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestRecaptcha(t *testing.T, handler http.HandlerFunc, cfgMut func(*Config)) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := &Config{
		Secret:    "test-secret",
		VerifyURL: srv.URL,
	}
	if cfgMut != nil {
		cfgMut(cfg)
	}

	client, err := NewGoogleRecaptcha(cfg)
	if err != nil {
		t.Fatalf("client creation failed: %v", err)
	}

	t.Cleanup(client.Close)
	return client
}

// A9 regression: nil config used to dereference and panic.
func TestRecaptchaNewNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewGoogleRecaptcha(nil) panicked: %v", r)
		}
	}()

	if _, err := NewGoogleRecaptcha(nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestRecaptchaNewRequiresSecret(t *testing.T) {
	if _, err := NewGoogleRecaptcha(&Config{}); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestRecaptchaVerifySuccess(t *testing.T) {
	var gotSecret, gotResponse string

	client := newTestRecaptcha(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotSecret = r.FormValue("secret")
		gotResponse = r.FormValue("response")

		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":  true,
			"hostname": "example.com",
			"score":    0.9,
			"action":   "login",
		})
	}, nil)

	res, err := client.Verify(context.Background(), VerifyOptions{Response: "token-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotSecret != "test-secret" || gotResponse != "token-123" {
		t.Fatalf("server saw secret=%q response=%q", gotSecret, gotResponse)
	}

	if !res.Success || res.Score != 0.9 || res.Action != "login" {
		t.Fatalf("unexpected response %+v", res)
	}
}

func TestRecaptchaVerifyRequiresToken(t *testing.T) {
	client := newTestRecaptcha(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for empty token")
	}, nil)

	if _, err := client.Verify(context.Background(), VerifyOptions{}); err == nil {
		t.Fatal("expected error for empty response token")
	}
}

func TestRecaptchaVerifyErrorCodes(t *testing.T) {
	client := newTestRecaptcha(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":     false,
			"error-codes": []string{"invalid-input-response", "weird-code"},
		})
	}, nil)

	_, err := client.Verify(context.Background(), VerifyOptions{Response: "tok"})
	if err == nil {
		t.Fatal("expected aggregated error")
	}

	if !strings.Contains(err.Error(), "invalid or malformed token") {
		t.Fatalf("expected mapped error code, got %v", err)
	}

	if !strings.Contains(err.Error(), "unknown error: weird-code") {
		t.Fatalf("expected unknown code passthrough, got %v", err)
	}
}

func TestRecaptchaRateLimit(t *testing.T) {
	client := newTestRecaptcha(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	}, func(cfg *Config) {
		cfg.RateLimit = 2
		cfg.RateWindow = time.Minute
	})

	opts := VerifyOptions{Response: "tok", RemoteIP: "10.0.0.1"}

	for i := 0; i < 2; i++ {
		if _, err := client.Verify(context.Background(), opts); err != nil {
			t.Fatalf("request %d unexpectedly failed: %v", i, err)
		}
	}

	_, err := client.Verify(context.Background(), opts)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}

	// A different IP is unaffected.
	if _, err := client.Verify(context.Background(), VerifyOptions{Response: "tok", RemoteIP: "10.0.0.2"}); err != nil {
		t.Fatalf("other IP should not be limited: %v", err)
	}
}

// A8 regression: Close used to close the cleanup channel unconditionally,
// so a second Close (or a deferred double close) panicked.
func TestRecaptchaDoubleCloseNoPanic(t *testing.T) {
	client := newTestRecaptcha(t, func(w http.ResponseWriter, r *http.Request) {}, nil)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("second Close panicked: %v", r)
		}
	}()

	client.Close()
	client.Close()
}

func TestRecaptchaValidateV3(t *testing.T) {
	res := &Response{Success: true, Score: 0.8, Action: "login", Hostname: "example.com"}

	if err := res.ValidateV3(0.5, []string{"login"}, []string{"example.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := res.ValidateV3(0.9, nil, nil); err == nil {
		t.Fatal("expected low score rejection")
	}

	if err := res.ValidateV3(0.5, []string{"purchase"}, nil); err == nil {
		t.Fatal("expected action rejection")
	}

	if err := res.ValidateV3(0.5, nil, []string{"other.com"}); err == nil {
		t.Fatal("expected hostname rejection")
	}

	failed := &Response{Success: false}
	if err := failed.ValidateV3(0, nil, nil); err == nil {
		t.Fatal("expected failure for unsuccessful verification")
	}
}

func TestRecaptchaOversizedResponseIsLimited(t *testing.T) {
	client := newTestRecaptcha(t, func(w http.ResponseWriter, r *http.Request) {
		// Valid JSON prefix followed by a huge padding blob.
		_, _ = w.Write([]byte(`{"success":true,"score":0.7,"junk":"`))
		_, _ = w.Write([]byte(strings.Repeat("A", 64*1024)))
		_, _ = w.Write([]byte(`"}`))
	}, func(cfg *Config) {
		cfg.MaxRespSize = 1024
	})

	// Decoding truncated JSON must error, not exhaust memory.
	if _, err := client.Verify(context.Background(), VerifyOptions{Response: "tok"}); err == nil {
		t.Fatal("expected decode error for truncated response")
	}
}
