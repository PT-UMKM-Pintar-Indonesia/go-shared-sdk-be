package sdk_helper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.ErrorLevel)
	m.Run()
}

func fastRetryConfigs() sdk_dto.HttpClientConfigs {
	return sdk_dto.HttpClientConfigs{
		RetryWaitMin: 1 * time.Millisecond,
		RetryWaitMax: 2 * time.Millisecond,
	}
}

func readData(t *testing.T, data any) string {
	t.Helper()
	if data == nil {
		return ""
	}
	reader, ok := data.(io.Reader)
	if !ok {
		t.Fatalf("Data is not an io.Reader: %T", data)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read Data: %v", err)
	}
	return string(raw)
}

func TestNewHttpClient_PropagatesStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"code":"not_found"}`)
	}))
	defer server.Close()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:    server.URL,
		Method: http.MethodGet,
	})

	if res.StatCode != http.StatusNotFound {
		t.Fatalf("expected propagated 404, got %d", res.StatCode)
	}
	if body := readData(t, res.Data); body != `{"code":"not_found"}` {
		t.Fatalf("expected body preserved on 4xx, got %q", body)
	}
}

func TestNewHttpClient_PostNotRetriedOnServerError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodPost,
		Body:    `{"amount":1000}`,
		Headers: map[string]string{"Content-Type": "application/json"},
		Configs: fastRetryConfigs(),
	})

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("POST must not be retried on 5xx (duplicate-payment safety), got %d attempts", got)
	}
	if res.StatCode != http.StatusInternalServerError {
		t.Fatalf("expected propagated 500, got %d", res.StatCode)
	}
}

func TestNewHttpClient_GetRetriedOnServerError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: fastRetryConfigs(),
	})

	if got := atomic.LoadInt32(&attempts); got != defaultRetryMax+1 {
		t.Fatalf("expected %d attempts for GET, got %d", defaultRetryMax+1, got)
	}
	if res.StatCode != http.StatusInternalServerError {
		t.Fatalf("expected propagated 500 after retries, got %d", res.StatCode)
	}
}

func TestNewHttpClient_ConfigIsolationUnderConcurrency(t *testing.T) {
	// Caller A asks for RetryMax=1 (2 attempts); caller B uses the default
	// (4 attempts). Before, Configs mutated a shared singleton, so these
	// calls raced and leaked configuration into each other.
	var attemptsA, attemptsB int32
	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptsA, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer serverA.Close()

	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptsB, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer serverB.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		opts := sdk_dto.HttpClientOptions{
			Url:     serverA.URL,
			Method:  http.MethodGet,
			Configs: fastRetryConfigs(),
		}
		opts.Configs.RetryMax = 1
		NewHttpClient(&opts)
	}()
	go func() {
		defer wg.Done()
		NewHttpClient(&sdk_dto.HttpClientOptions{
			Url:     serverB.URL,
			Method:  http.MethodGet,
			Configs: fastRetryConfigs(),
		})
	}()
	wg.Wait()

	if got := atomic.LoadInt32(&attemptsA); got != 2 {
		t.Fatalf("caller A expected 2 attempts (RetryMax=1), got %d", got)
	}
	if got := atomic.LoadInt32(&attemptsB); got != defaultRetryMax+1 {
		t.Fatalf("caller B expected %d attempts (default), got %d — config leaked between callers", defaultRetryMax+1, got)
	}
}

func TestNewHttpClient_ErrorHandlerReturningBothDoesNotLeak(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "boom-body")
	}))
	defer server.Close()

	cfgs := fastRetryConfigs()
	cfgs.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		// Called when retries are exhausted. Hand back both a live response
		// and an error, like PassthroughErrorHandler would. The helper must
		// still drain+close the body instead of leaking the connection.
		return resp, errors.New("boom")
	}

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: cfgs,
	})

	if res.StatCode < http.StatusBadRequest {
		t.Fatalf("expected error status when ErrorHandler returns an error, got %d", res.StatCode)
	}

	// The connection must have been drained+closed, so the pool can reuse it:
	// a second request should still succeed over the same client.
	res2 := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: fastRetryConfigs(),
	})
	if res2.StatCode != http.StatusInternalServerError {
		t.Fatalf("expected follow-up request to work after error path, got stat=%d", res2.StatCode)
	}
	if body := readData(t, res2.Data); body != "boom-body" {
		t.Fatalf("expected body on exhausted-retry 5xx, got %q", body)
	}
}

func TestNewHttpClient_TruncatesLargeBody(t *testing.T) {
	payload := strings.Repeat("x", maxResponseSize+1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, payload)
	}))
	defer server.Close()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: fastRetryConfigs(),
	})

	body := readData(t, res.Data)
	if len(body) != maxResponseSize {
		t.Fatalf("expected body capped at %d bytes, got %d", maxResponseSize, len(body))
	}
}

func TestNewHttpClient_TimeoutMapsTo504(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		fmt.Fprint(w, "late")
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Ctx:     ctx,
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: fastRetryConfigs(),
	})

	if res.StatCode != http.StatusGatewayTimeout {
		t.Fatalf("expected 504 on context deadline, got %d (%v)", res.StatCode, res.ErrMsg)
	}
}

func TestNewHttpClient_UrlValuesFormBody(t *testing.T) {
	var gotCT, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		fmt.Fprint(w, "ok")
	}))
	defer server.Close()

	form := url.Values{"merchant": {"abc"}, "amount": {"1000"}}
	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodPost,
		Body:    form,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	})

	if res.StatCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatCode)
	}
	if gotCT != "application/x-www-form-urlencoded" {
		t.Fatalf("unexpected content type %q", gotCT)
	}
	if gotBody != form.Encode() {
		t.Fatalf("expected form-encoded body %q, got %q", form.Encode(), gotBody)
	}
}

func TestNewHttpClient_CustomCheckRetryOverridesDefault(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfgs := fastRetryConfigs()
	cfgs.CheckRetry = retryablehttp.DefaultRetryPolicy

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodPost,
		Body:    "payload",
		Configs: cfgs,
	})

	if got := atomic.LoadInt32(&attempts); got != defaultRetryMax+1 {
		t.Fatalf("explicit CheckRetry must be honored (it was silently ignored before), got %d attempts", got)
	}
	if res.StatCode != http.StatusServiceUnavailable {
		t.Fatalf("expected propagated 503, got %d", res.StatCode)
	}
}

func TestTrimLogString_FastPath(t *testing.T) {
	short := strings.Repeat("a", 100)
	if got := TrimLogString(short); got != short {
		t.Fatalf("short string must be returned as-is")
	}

	long := strings.Repeat("b", 5000)
	got := TrimLogString(long)
	if len([]rune(got)) > DefaultMaxLogLength {
		t.Fatalf("truncated output must respect max length, got %d runes", len([]rune(got)))
	}
	if !strings.HasSuffix(got, TruncatedSuffix) {
		t.Fatalf("truncated output must carry suffix, got %q", got[len(got)-20:])
	}
}
