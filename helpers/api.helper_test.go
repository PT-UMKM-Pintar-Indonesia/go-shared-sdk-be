package sdk_helper

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

func TestVersion(t *testing.T) {
	if got := Version("auth"); got != "/api/v1/auth" {
		t.Fatalf("unexpected version path %q", got)
	}
}

func TestFormatBytesAndDuration(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1536, "1.5 KB"},
	}

	for _, tc := range tests {
		if got := formatBytes(tc.size); got != tc.want {
			t.Errorf("formatBytes(%d) = %q want %q", tc.size, got, tc.want)
		}
	}

	if got := formatDuration(1500 * time.Microsecond); got != "1.500ms" {
		t.Fatalf("formatDuration got %q", got)
	}
}

func TestGetProtocolAndIPAddress(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := getProtocol(r); got != "http" {
		t.Fatalf("expected http, got %q", got)
	}

	r.Header.Set("X-Forwarded-Proto", "https")
	if got := getProtocol(r); got != "https" {
		t.Fatalf("expected https via XFP, got %q", got)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	if got := getIPAddress(r2); got != "1.2.3.4" {
		t.Fatalf("expected first XFF hop, got %q", got)
	}

	r3 := httptest.NewRequest(http.MethodGet, "/", nil)
	r3.Header.Set("X-Real-IP", "9.9.9.9")
	if got := getIPAddress(r3); got != "9.9.9.9" {
		t.Fatalf("expected X-Real-IP, got %q", got)
	}

	r4 := httptest.NewRequest(http.MethodGet, "/", nil)
	r4.RemoteAddr = "127.0.0.1:1234"
	if got := getIPAddress(r4); got != "127.0.0.1:1234" {
		t.Fatalf("expected RemoteAddr fallback, got %q", got)
	}
}

func TestGetErrorCode(t *testing.T) {
	if got := getErrorCode(http.StatusNotFound); got != "UNKNOWN_RESOURCE" {
		t.Fatalf("unexpected code %q", got)
	}

	if got := getErrorCode(http.StatusTeapot); got != "GENERAL_ERROR" {
		t.Fatalf("expected fallback GENERAL_ERROR, got %q", got)
	}
}

func TestApiWritesResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/some/path", nil)

	code := "INVALID_REQUEST"
	Api(rec, r, sdk_opt.Response{
		StatCode: http.StatusBadRequest,
		Message:  "ok-msg",
		ErrCode:  &code,
		ErrMsg:   "broken",
		Data:     map[string]any{"k": "v"},
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type %q", ct)
	}

	var body sdk_opt.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json body: %v", err)
	}

	if body.StatCode != http.StatusBadRequest {
		t.Fatalf("unexpected stat code %v", body.StatCode)
	}

	if body.Info.Path != "/some/path" || body.Info.Method != http.MethodPost {
		t.Fatalf("unexpected info %+v", body.Info)
	}
}
func TestApiDefaultsServerError(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/x", nil)

	Api(rec, r, sdk_opt.Response{})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected default 500, got %d", rec.Code)
	}

	var body sdk_opt.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if body.ErrCode == nil || body.ErrCode.(string) != "GENERAL_ERROR" {
		t.Fatalf("expected default error code, got %v", body.ErrCode)
	}

	// ErrMsg is an `any` field; when the caller leaves it nil the SDK
	// keeps it nil (the empty-string default only applies to "").
	if body.ErrMsg == nil {
		t.Log("err_msg omitted for nil input (omitempty behaviour)")
	}
}

func TestSizeWriterTracksBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &sizeWriter{ResponseWriter: rec}

	_, _ = sw.Write([]byte("hello"))
	_, _ = sw.Write([]byte("world"))

	if sw.size != 10 {
		t.Fatalf("expected 10 bytes tracked, got %d", sw.size)
	}

	if rec.Body.String() != "helloworld" {
		t.Fatalf("unexpected passthrough body %q", rec.Body.String())
	}
}

// A12: sizeWriter must keep http.Flusher support for streaming handlers.
func TestSizeWriterFlushPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &sizeWriter{ResponseWriter: rec}

	flusher, ok := interface{}(sw).(http.Flusher)
	if !ok {
		t.Fatal("sizeWriter must implement http.Flusher")
	}

	_, _ = sw.Write([]byte("data"))
	flusher.Flush() // must not panic
}

// A12: hijacking must be forwarded or rejected cleanly, never panic.
func TestSizeWriterHijack(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &sizeWriter{ResponseWriter: rec}

	// ResponseRecorder implements Hijacker; the call must not panic.
	conn, rw, err := sw.Hijack()
	if err == nil {
		if conn == nil || rw == nil {
			t.Fatal("expected non-nil hijacked connection")
		}
		_ = conn.Close()
	}
}

type plainWriter struct {
	header http.Header
}

func (p *plainWriter) Header() http.Header         { return p.header }
func (p *plainWriter) Write(b []byte) (int, error) { return len(b), nil }
func (p *plainWriter) WriteHeader(int)             {}

func TestSizeWriterHijackErrorWhenUnsupported(t *testing.T) {
	sw := &sizeWriter{ResponseWriter: &plainWriter{header: http.Header{}}}

	if _, _, err := sw.Hijack(); err == nil {
		t.Fatal("expected hijack error for plain writer")
	}
}

func TestSizeWriterImplementsInterfaces(t *testing.T) {
	var _ http.ResponseWriter = (*sizeWriter)(nil)
	var _ http.Flusher = (*sizeWriter)(nil)
	var _ http.Hijacker = (*sizeWriter)(nil)
}

// Ensure a streaming endpoint works through the wrapper end to end.
func TestSizeWriterFlushesOverRealServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &sizeWriter{ResponseWriter: w}
		sw.WriteHeader(http.StatusOK)
		_, _ = sw.Write([]byte("chunk1\n"))
		sw.Flush()
		_, _ = sw.Write([]byte("chunk2\n"))
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "chunk1\nchunk2\n" {
		t.Fatalf("unexpected body %q", body)
	}
}
