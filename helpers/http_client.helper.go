package sdk_helper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type requestPolicy struct {
	perAttempt     time.Duration
	budget         time.Duration
	retryMax       int
	retryWaitMin   time.Duration
	retryWaitMax   time.Duration
	idempotent     bool
	retryOnTimeout bool
	logRequests    bool
}

func resolvePolicy(options *sdk_dto.HttpClientOptions) requestPolicy {
	tune := httpTuningConfig()

	policy := requestPolicy{
		perAttempt:     tune.timeout,
		retryMax:       defaultRetryMax,
		retryWaitMin:   defaultRetryWaitMin,
		retryWaitMax:   defaultRetryWaitMax,
		idempotent:     isIdempotentMethod(options.Method) || hasIdempotencyKey(options.Headers),
		retryOnTimeout: tune.retryOnTimeout,
		logRequests:    tune.logRequests,
	}

	if options.Configs.Timeout > 0 {
		policy.perAttempt = options.Configs.Timeout
	}

	switch {
	case options.Configs.RetryMax > 0:
		policy.retryMax = options.Configs.RetryMax
	case options.Configs.RetryMax < 0:
		policy.retryMax = 0
	}

	if options.Configs.RetryWaitMin > 0 {
		policy.retryWaitMin = options.Configs.RetryWaitMin
	}

	if options.Configs.RetryWaitMax > 0 {
		policy.retryWaitMax = options.Configs.RetryWaitMax
	}

	if options.Configs.RetryOnTimeout {
		policy.retryOnTimeout = true
	}

	if options.Configs.LogRequests {
		policy.logRequests = true
	}

	policy.budget = options.Configs.Budget
	if policy.budget <= 0 {
		policy.budget = tune.budget
	}

	if policy.budget <= 0 {
		policy.budget = policy.perAttempt
	}

	if policy.budget < policy.perAttempt {
		policy.budget = policy.perAttempt
	}

	if policy.budget > maxHTTPBudget {
		policy.budget = maxHTTPBudget
	}

	return policy
}

func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	}

	return false
}

func hasIdempotencyKey(headers map[string]string) bool {
	for key := range headers {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "idempotency-key", "x-idempotency-key":
			return true
		}
	}

	return false
}

var (
	sharedTransport *http.Transport
	transportOnce   sync.Once
)

func getSharedTransport() *http.Transport {
	transportOnce.Do(func() {
		tune := httpTuningConfig()

		sharedTransport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          tune.maxIdleConns,
			MaxIdleConnsPerHost:   tune.maxIdleConnsPerHost,
			MaxConnsPerHost:       tune.maxConnsPerHost,
			IdleConnTimeout:       tune.idleConnTimeout,
			TLSHandshakeTimeout:   tune.tlsHandshakeTimeout,
			ResponseHeaderTimeout: tune.responseHeaderTimeout,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   tune.connectTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
	})

	return sharedTransport
}

func httpClientFor(transport http.RoundTripper, perAttempt time.Duration) *http.Client {
	if transport == nil {
		transport = getSharedTransport()
	}

	return &http.Client{Transport: transport, Timeout: perAttempt}
}

func newRetryClient(options *sdk_dto.HttpClientOptions, policy requestPolicy) *retryablehttp.Client {
	client := &retryablehttp.Client{
		HTTPClient:      httpClientFor(options.Transport, policy.perAttempt),
		RetryMax:        policy.retryMax,
		RetryWaitMin:    policy.retryWaitMin,
		RetryWaitMax:    policy.retryWaitMax,
		Backoff:         retryablehttp.DefaultBackoff,
		CheckRetry:      newCheckRetry(policy),
		ErrorHandler:    retryablehttp.PassthroughErrorHandler,
		Logger:          nil,
		RequestLogHook:  newRequestLogHook(policy.logRequests),
		ResponseLogHook: newResponseLogHook(policy.logRequests),
	}

	if options.Configs.Backoff != nil {
		client.Backoff = options.Configs.Backoff
	}

	if options.Configs.CheckRetry != nil {
		client.CheckRetry = options.Configs.CheckRetry
	}

	if options.Configs.ErrorHandler != nil {
		client.ErrorHandler = options.Configs.ErrorHandler
	}

	if options.Configs.PrepareRetry != nil {
		client.PrepareRetry = options.Configs.PrepareRetry
	}

	return client
}

func newRequestLogHook(verbose bool) retryablehttp.RequestLogHook {
	return func(_ retryablehttp.Logger, r *http.Request, attempt int) {
		entry := logrus.WithFields(logrus.Fields{
			"namespace": "http_client",
			"url":       redactURL(r.URL),
			"method":    r.Method,
			"attempt":   attempt,
		})

		if verbose {
			entry.Info("LOG_WORKER_HTTP_REQUEST")
			return
		}

		entry.Debug("LOG_WORKER_HTTP_REQUEST")
	}
}

func newResponseLogHook(verbose bool) retryablehttp.ResponseLogHook {
	return func(_ retryablehttp.Logger, r *http.Response) {
		var (
			url    string
			method string
		)
		if r.Request != nil {
			url = redactURL(r.Request.URL)
			method = r.Request.Method
		}

		entry := logrus.WithFields(logrus.Fields{
			"namespace":   "http_client",
			"url":         url,
			"method":      method,
			"status_code": r.StatusCode,
		})

		if verbose || r.StatusCode >= http.StatusInternalServerError {
			entry.Info("LOG_WORKER_HTTP_RESPONSE")
			return
		}

		entry.Debug("LOG_WORKER_HTTP_RESPONSE")
	}
}

func redactURL(u *url.URL) string {
	if u == nil {
		return sdk_cons.EMPTY
	}

	clone := *u
	clone.User = nil
	clone.RawQuery = sdk_cons.EMPTY
	clone.Fragment = sdk_cons.EMPTY

	return clone.String()
}

func redactURLString(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	return redactURL(parsed)
}

func newCheckRetry(policy requestPolicy) retryablehttp.CheckRetry {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if err != nil {
			if isPreWriteNetworkError(err) {
				return true, nil
			}

			if policy.idempotent {
				if isTimeoutError(err) && !policy.retryOnTimeout {
					return false, nil
				}

				return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
			}

			return false, nil
		}

		if resp == nil {
			return false, nil
		}

		if policy.idempotent && resp.StatusCode >= http.StatusInternalServerError {
			return true, nil
		}

		return false, nil
	}
}

func isPreWriteNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return true
	}

	if strings.Contains(strings.ToLower(err.Error()), "tls handshake timeout") {
		return true
	}

	return false
}

func NewHttpClient(options *sdk_dto.HttpClientOptions) (res sdk_opt.Response) {
	if options == nil {
		return httpClientError(ErrOptionsRequired)
	}

	parser := NewParser()
	policy := resolvePolicy(options)
	client := newRetryClient(options, policy)

	var bodyReader io.Reader
	switch v := options.Body.(type) {

	case nil:
	case io.Reader:
		bodyReader = v

	case string:
		bodyReader = strings.NewReader(v)

	case []byte:
		bodyReader = bytes.NewReader(v)

	case url.Values:
		bodyReader = strings.NewReader(v.Encode())

	default:
		contentType := sdk_cons.EMPTY
		if options.Headers != nil {
			contentType = strings.ToLower(options.Headers["Content-Type"])
		}

		switch {

		case strings.Contains(contentType, "json"):
			bodyBytes, err := parser.Marshal(v)
			if err != nil {
				res.StatCode = http.StatusBadRequest
				res.ErrMsg = fmt.Errorf("failed to marshal body: %w", err).Error()
				return
			}

			bodyReader = bytes.NewReader(bodyBytes)
		case strings.Contains(contentType, "x-www-form-urlencoded"):
			if form, ok := toFormValues(v); ok {
				bodyReader = strings.NewReader(form.Encode())
			} else {
				logrus.Warnf("NewHttpClient: body %T is not form-encodable for %s; sending fmt %%v representation", v, redactURLString(options.Url))
				bodyReader = strings.NewReader(fmt.Sprintf("%v", v))
			}

		default:
			logrus.Warnf("NewHttpClient: body %T serialized via fmt %%v for %s (no JSON Content-Type set)", v, redactURLString(options.Url))
			bodyReader = strings.NewReader(fmt.Sprintf("%v", v))
		}
	}

	ctx := options.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, policy.budget)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, options.Method, options.Url, bodyReader)
	if err != nil {
		return httpClientError(err)
	}

	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return httpClientError(err)
	}

	httpRes, httpErr := client.Do(retryReq)

	if httpRes != nil && httpErr != nil {
		io.Copy(io.Discard, io.LimitReader(httpRes.Body, errorDrainLimit))
		httpRes.Body.Close()
		return httpClientError(httpErr)
	}

	if httpErr != nil {
		return httpClientError(httpErr)
	}
	defer httpRes.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(httpRes.Body, maxResponseSize+1))
	if err != nil {
		return httpClientError(err)
	}

	if len(bodyBytes) > maxResponseSize {
		logrus.Warnf("HTTP Response exceeded %d bytes limit for URL: %s", maxResponseSize, redactURLString(options.Url))
		return httpClientError(ErrResponseTooLarge)
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.WithFields(logrus.Fields{
			"namespace":        "http_client",
			"url":              redactURLString(options.Url),
			"method":           options.Method,
			"status_code":      httpRes.StatusCode,
			"response_preview": TrimLogStringWithMax(string(bodyBytes), logBodyPreviewLimit),
		}).Debug("LOG_WORKER_HTTP_RESPONSE_BODY")
	}

	res.StatCode = httpRes.StatusCode
	if httpRes.StatusCode < http.StatusBadRequest {
		res.Message = "Network Success"
	} else {
		res.Message = httpRes.Status
	}
	res.Data = io.NopCloser(bytes.NewReader(bodyBytes))

	return
}

func toFormValues(v any) (url.Values, bool) {
	switch m := v.(type) {

	case url.Values:
		return m, true

	case map[string]string:
		form := make(url.Values, len(m))
		for key, value := range m {
			form.Set(key, value)
		}
		return form, true

	case map[string][]string:
		return url.Values(m), true

	default:
		return nil, false
	}
}

func httpClientError(err error) (res sdk_opt.Response) {
	if err == nil {
		res.StatCode = http.StatusOK
		res.Message = "Network Success"
		return
	}

	switch {
	case errors.Is(err, ErrOptionsRequired):
		res.StatCode = http.StatusBadRequest
		res.ErrMsg = err.Error()

	case errors.Is(err, context.DeadlineExceeded) || isTimeoutError(err):
		res.StatCode = http.StatusGatewayTimeout
		res.ErrCode = sdk_cons.SNAP_REQUEST_TIMEOUT
		res.ErrMsg = fmt.Sprintf("%s: %v: %v", upstreamFailurePrefix, ErrUpstreamTimeout, err)

	case errors.Is(err, context.Canceled):
		res.StatCode = http.StatusRequestTimeout
		res.ErrMsg = fmt.Sprintf("%s: %v: %v", upstreamFailurePrefix, ErrUpstreamCanceled, err)

	case errors.Is(err, ErrResponseTooLarge):
		res.StatCode = http.StatusBadGateway
		res.ErrMsg = fmt.Sprintf("%s: %v", upstreamFailurePrefix, err)
		res.Message = res.ErrMsg

	default:
		var netErr *net.OpError
		if errors.As(err, &netErr) || errors.Is(err, net.ErrClosed) || errors.Is(err, net.ErrWriteToConnected) {
			res.StatCode = http.StatusBadGateway
			res.ErrMsg = fmt.Sprintf("%s: %v: %v", upstreamFailurePrefix, ErrUpstreamNetwork, err)

			return
		}

		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.ErrNoProgress) {
			res.StatCode = http.StatusBadGateway
			res.ErrMsg = fmt.Sprintf("%s: %v: %v", upstreamFailurePrefix, ErrUpstreamNetwork, err)
			res.Message = res.ErrMsg

			return
		}

		res.StatCode = http.StatusBadRequest
		res.ErrMsg = err.Error()
		res.Message = err.Error()
	}

	return
}

func isTimeoutError(err error) bool {
	var netErr net.Error

	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	return false
}
