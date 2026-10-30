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

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
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

func TestNewHttpClient_RejectsOversizeBody(t *testing.T) {
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

	// A truncated body used to surface downstream as a confusing JSON parse
	// error; now it is an explicit upstream error.
	if res.StatCode != http.StatusBadGateway {
		t.Fatalf("expected 502 for oversize body, got %d (%v)", res.StatCode, res.ErrMsg)
	}
	if !strings.Contains(fmt.Sprint(res.ErrMsg), ErrResponseTooLarge.Error()) {
		t.Fatalf("expected ErrResponseTooLarge in message, got %v", res.ErrMsg)
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

func TestNewHttpClient_TimeoutCarriesSNAPCodeAndObscurePrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprint(w, "late")
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Ctx:     ctx,
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: fastRetryConfigs(),
	})

	if res.StatCode != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d (%v)", res.StatCode, res.ErrMsg)
	}
	if code, _ := res.ErrCode.(string); code != sdk_cons.SNAP_REQUEST_TIMEOUT {
		t.Fatalf("expected SNAP code %s, got %v", sdk_cons.SNAP_REQUEST_TIMEOUT, res.ErrCode)
	}
	if !strings.Contains(fmt.Sprint(res.ErrMsg), "Request to external network failed") {
		t.Fatalf("timeout message must trip the api-be OBSCURE guard, got %v", res.ErrMsg)
	}
	if !IsTimeoutResponse(res) {
		t.Fatal("IsTimeoutResponse must report true for a 50400 response")
	}
}

func TestNewHttpClient_TimeoutNotRetriedByDefaultOnPost(t *testing.T) {
	// The request was written; retrying a disbursement here could move money
	// twice. Default policy must surface the timeout after a single attempt.
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		time.Sleep(150 * time.Millisecond)
		fmt.Fprint(w, "late")
	}))
	defer server.Close()

	cfgs := fastRetryConfigs()
	cfgs.Timeout = 30 * time.Millisecond
	cfgs.Budget = 500 * time.Millisecond

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodPost,
		Body:    `{"amount":1000}`,
		Headers: map[string]string{"Content-Type": "application/json"},
		Configs: cfgs,
	})

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("POST timeout must not be retried after write, got %d attempts", got)
	}
	if res.StatCode != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", res.StatCode)
	}
}

func TestNewHttpClient_TimeoutRetriedWhenEnabledForIdempotentCall(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		time.Sleep(120 * time.Millisecond)
		fmt.Fprint(w, "late")
	}))
	defer server.Close()

	cfgs := fastRetryConfigs()
	cfgs.Timeout = 30 * time.Millisecond
	cfgs.Budget = 400 * time.Millisecond
	cfgs.RetryOnTimeout = true

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: cfgs,
	})

	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Fatalf("expected retries once RetryOnTimeout is enabled, got %d attempts", got)
	}
	if res.StatCode != http.StatusGatewayTimeout {
		t.Fatalf("expected 504 after exhausted budget, got %d", res.StatCode)
	}
}

func TestNewHttpClient_PostWithIdempotencyKeyIsRetried(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:    server.URL,
		Method: http.MethodPost,
		Body:   `{"amount":1000}`,
		Headers: map[string]string{
			"Content-Type":    "application/json",
			"Idempotency-Key": "trx-123",
		},
		Configs: fastRetryConfigs(),
	})

	if got := atomic.LoadInt32(&attempts); got != defaultRetryMax+1 {
		t.Fatalf("request carrying Idempotency-Key is replayable, expected %d attempts, got %d", defaultRetryMax+1, got)
	}
	if res.StatCode != http.StatusInternalServerError {
		t.Fatalf("expected propagated 500, got %d", res.StatCode)
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

func TestNewHttpClient_BudgetIsBounded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(120 * time.Millisecond)
		fmt.Fprint(w, "late")
	}))
	defer server.Close()

	cfgs := fastRetryConfigs()
	cfgs.Timeout = 60 * time.Millisecond
	cfgs.Budget = 200 * time.Millisecond
	cfgs.RetryOnTimeout = true

	start := time.Now()
	res := NewHttpClient(&sdk_dto.HttpClientOptions{
		Url:     server.URL,
		Method:  http.MethodGet,
		Configs: cfgs,
	})
	elapsed := time.Since(start)

	if res.StatCode != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", res.StatCode)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("budget must bound the total call, took %s", elapsed)
	}
}

func TestNewHttpClient_NilOptionsDoesNotPanic(t *testing.T) {
	res := NewHttpClient(nil)
	if res.StatCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for nil options, got %d", res.StatCode)
	}
}

func TestRedactURL(t *testing.T) {
	parsed := &url.URL{
		Scheme:   "https",
		Host:     "vendor.example",
		Path:     "/v1/payment",
		RawQuery: "api_key=secret&token=secret",
		User:     url.UserPassword("user", "pass"),
	}

	got := redactURL(parsed)
	if strings.Contains(got, "secret") || strings.Contains(got, "pass") {
		t.Fatalf("credentials leaked into log url: %s", got)
	}
	if got != "https://vendor.example/v1/payment" {
		t.Fatalf("unexpected redacted url: %s", got)
	}
}

func TestIsTimeoutResponse(t *testing.T) {
	if !IsTimeoutResponse(sdk_opt.Response{ErrCode: "50400"}) {
		t.Fatal("expected true for 50400")
	}
	if IsTimeoutResponse(sdk_opt.Response{StatCode: http.StatusGatewayTimeout}) {
		t.Fatal("plain 504 without SNAP code must not be treated as vendor timeout")
	}
}

func BenchmarkNewHttpClient_Parallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res := NewHttpClient(&sdk_dto.HttpClientOptions{
				Url:     server.URL,
				Method:  http.MethodPost,
				Body:    `{"amount":1000}`,
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			if res.StatCode != http.StatusOK {
				b.Fatalf("unexpected status %d", res.StatCode)
			}
		}
	})
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

func TestNewHttpClient_SelectiveAuditLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"invalid_account","pin":"123456"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","name":"Budi"}`)
	}))
	defer server.Close()

	ctx := context.WithValue(context.Background(), sdk_cons.CtxRequestID, "req-test-12345")

	// 1. Success call (200 OK): caller gets complete response data
	resSuccess := NewHttpClient(&sdk_dto.HttpClientOptions{
		Ctx:     ctx,
		Url:     server.URL + "/success",
		Method:  http.MethodPost,
		Body:    `{"amount":10000}`,
		Headers: map[string]string{"Content-Type": "application/json", "Authorization": "secret-auth"},
	})
	if resSuccess.StatCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resSuccess.StatCode)
	}
	data := readData(t, resSuccess.Data)
	if !strings.Contains(string(data), `"name":"Budi"`) {
		t.Fatalf("body stream was not restored properly, got: %s", string(data))
	}

	// 2. Error call (400 Bad Request): caller gets complete error body
	resError := NewHttpClient(&sdk_dto.HttpClientOptions{
		Ctx:     ctx,
		Url:     server.URL + "/error",
		Method:  http.MethodPost,
		Body:    `{"pin":"999999","amount":10000}`,
		Headers: map[string]string{"Content-Type": "application/json"},
	})
	if resError.StatCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resError.StatCode)
	}
	errData := readData(t, resError.Data)
	if !strings.Contains(string(errData), `"invalid_account"`) {
		t.Fatalf("error body stream was not restored properly, got: %s", string(errData))
	}
}
