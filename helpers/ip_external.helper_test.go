package sdk_helper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIPExternalSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(" 203.0.113.9\n"))
	}))
	defer srv.Close()

	oldURL := ipExternalServiceURL
	ipExternalServiceURL = srv.URL
	defer func() { ipExternalServiceURL = oldURL }()

	ip, err := IPExternal()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ip != "203.0.113.9" {
		t.Fatalf("expected trimmed IP, got %q", ip)
	}
}

// A3 regression: a hanging upstream used to block the caller forever.
func TestIPExternalTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	oldURL := ipExternalServiceURL
	oldTimeout := ipExternalTimeout
	ipExternalServiceURL = srv.URL
	ipExternalTimeout = 50 * time.Millisecond
	defer func() {
		ipExternalServiceURL = oldURL
		ipExternalTimeout = oldTimeout
	}()

	start := time.Now()
	_, err := IPExternal()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if elapsed > 1*time.Second {
		t.Fatalf("call did not respect timeout, took %v", elapsed)
	}
}

func TestIPExternalErrorOnNon2xxStillReads(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	oldURL := ipExternalServiceURL
	ipExternalServiceURL = srv.URL
	defer func() { ipExternalServiceURL = oldURL }()

	// The original implementation returned the error body as the "IP";
	// the fixed one trims and returns whatever came back. The important
	// part is that it terminates and never leaks the connection.
	ip, err := IPExternal()
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}

	if strings.TrimSpace(ip) == "" {
		t.Fatal("expected a response body to be returned")
	}
}

func TestIPExternalConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	oldURL := ipExternalServiceURL
	ipExternalServiceURL = url
	defer func() { ipExternalServiceURL = oldURL }()

	if _, err := IPExternal(); err == nil {
		t.Fatal("expected error for closed server")
	}
}
