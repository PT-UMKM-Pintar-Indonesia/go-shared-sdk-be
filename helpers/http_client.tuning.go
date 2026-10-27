package sdk_helper

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

const (
	defaultHTTPTimeout = 15 * time.Second

	defaultConnectTimeout        = 2 * time.Second
	defaultTLSHandshakeTimeout   = 3 * time.Second
	defaultResponseHeaderTimeout = 8 * time.Second
	defaultIdleConnTimeout       = 30 * time.Second

	defaultMaxConnsPerHost     = 64
	defaultMaxIdleConns        = 400
	defaultMaxIdleConnsPerHost = 32

	maxHTTPBudget   = 60 * time.Second
	maxResponseSize = 5 * 1024 * 1024
	errorDrainLimit = 64 * 1024

	defaultRetryMax     = 3
	defaultRetryWaitMin = 250 * time.Millisecond
	defaultRetryWaitMax = 1 * time.Second

	logBodyPreviewLimit   = 1024
	upstreamFailurePrefix = "Request to external network failed"
)

var (
	ErrUpstreamTimeout  = errors.New("upstream timeout")
	ErrUpstreamCanceled = errors.New("upstream request canceled")
	ErrUpstreamNetwork  = errors.New("upstream network error")
	ErrResponseTooLarge = errors.New("upstream response exceeds size limit")
	ErrOptionsRequired  = errors.New("http client options is required")
)

type httpTuning struct {
	timeout               time.Duration
	budget                time.Duration
	connectTimeout        time.Duration
	tlsHandshakeTimeout   time.Duration
	responseHeaderTimeout time.Duration
	idleConnTimeout       time.Duration
	maxConnsPerHost       int
	maxIdleConns          int
	maxIdleConnsPerHost   int
	retryOnTimeout        bool
	logRequests           bool
}

var (
	tuningOnce sync.Once
	tuning     httpTuning
)

func httpTuningConfig() httpTuning {
	tuningOnce.Do(func() {
		tuning = httpTuning{
			timeout:               envDuration(defaultHTTPTimeout, "SDK_HTTP_TIMEOUT_MS", "HTTP_TIMEOUT_MS"),
			budget:                envDuration(0, "SDK_HTTP_BUDGET_MS", "HTTP_BUDGET_MS"),
			connectTimeout:        envDuration(defaultConnectTimeout, "SDK_HTTP_CONNECT_TIMEOUT_MS", "HTTP_CONNECT_TIMEOUT_MS"),
			tlsHandshakeTimeout:   envDuration(defaultTLSHandshakeTimeout, "SDK_HTTP_TLS_TIMEOUT_MS", "HTTP_TLS_TIMEOUT_MS"),
			responseHeaderTimeout: envDuration(defaultResponseHeaderTimeout, "SDK_HTTP_RESPONSE_HEADER_TIMEOUT_MS", "HTTP_RESPONSE_HEADER_TIMEOUT_MS"),
			idleConnTimeout:       envDuration(defaultIdleConnTimeout, "SDK_HTTP_IDLE_TIMEOUT_MS", "HTTP_IDLE_TIMEOUT_MS"),
			maxConnsPerHost:       envInt(defaultMaxConnsPerHost, "SDK_HTTP_MAX_CONNS_PER_HOST", "HTTP_MAX_CONNS_PER_HOST"),
			maxIdleConns:          envInt(defaultMaxIdleConns, "SDK_HTTP_MAX_IDLE_CONNS", "HTTP_MAX_IDLE_CONNS"),
			maxIdleConnsPerHost:   envInt(defaultMaxIdleConnsPerHost, "SDK_HTTP_MAX_IDLE_CONNS_PER_HOST", "HTTP_MAX_IDLE_CONNS_PER_HOST"),
			retryOnTimeout:        envBool(false, "SDK_HTTP_RETRY_ON_TIMEOUT", "HTTP_RETRY_ON_TIMEOUT"),
			logRequests:           envBool(false, "SDK_HTTP_LOG_REQUESTS", "HTTP_LOG_REQUESTS"),
		}
	})

	return tuning
}

func IsTimeoutResponse(res sdk_opt.Response) bool {
	code, ok := res.ErrCode.(string)

	return ok && code == sdk_cons.SNAP_REQUEST_TIMEOUT
}

func envDuration(def time.Duration, keys ...string) time.Duration {
	for _, key := range keys {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}

		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			return parsed
		}

		if millis, err := strconv.Atoi(raw); err == nil && millis > 0 {
			return time.Duration(millis) * time.Millisecond
		}
	}

	return def
}

func envInt(def int, keys ...string) int {
	for _, key := range keys {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}

		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			return parsed
		}
	}

	return def
}

func envBool(def bool, keys ...string) bool {
	for _, key := range keys {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}

		if parsed, err := strconv.ParseBool(raw); err == nil {
			return parsed
		}
	}

	return def
}
