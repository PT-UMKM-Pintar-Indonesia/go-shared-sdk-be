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

const (
	defaultHTTPTimeout = 15 * time.Second
	maxResponseSize    = 5 * 1024 * 1024
	errorDrainLimit    = 64 * 1024

	defaultRetryMax     = 3
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 3 * time.Second

	logBodyPreviewLimit = 1024
)

var (
	sharedHTTPClient *http.Client
	initOnce         sync.Once
)

func getSharedHTTPClient() *http.Client {
	initOnce.Do(func() {
		transport := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          400,
			MaxIdleConnsPerHost:   200,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}

		sharedHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   defaultHTTPTimeout,
		}
	})

	return sharedHTTPClient
}

func newRetryClient(options *sdk_dto.HttpClientOptions) *retryablehttp.Client {
	httpClient := getSharedHTTPClient()
	if options.Transport != nil {
		httpClient = &http.Client{
			Transport: options.Transport,
			Timeout:   defaultHTTPTimeout,
		}
	}

	client := &retryablehttp.Client{
		HTTPClient:      httpClient,
		RetryMax:        defaultRetryMax,
		RetryWaitMin:    defaultRetryWaitMin,
		RetryWaitMax:    defaultRetryWaitMax,
		Backoff:         retryablehttp.DefaultBackoff,
		CheckRetry:      defaultCheckRetry(options.Method),
		ErrorHandler:    retryablehttp.PassthroughErrorHandler,
		Logger:          nil,
		RequestLogHook:  logRequestHook,
		ResponseLogHook: logResponseHook,
	}

	if options.Configs.RetryWaitMin != 0 {
		client.RetryWaitMin = options.Configs.RetryWaitMin
	}

	if options.Configs.RetryWaitMax != 0 {
		client.RetryWaitMax = options.Configs.RetryWaitMax
	}

	if options.Configs.RetryMax != 0 {
		client.RetryMax = options.Configs.RetryMax
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

func logRequestHook(_ retryablehttp.Logger, r *http.Request, attempt int) {
	logrus.WithFields(logrus.Fields{
		"namespace": "http_client",
		"url":       r.URL.String(),
		"method":    r.Method,
		"attempt":   attempt,
	}).Info("LOG_WORKER_HTTP_REQUEST")
}

func logResponseHook(_ retryablehttp.Logger, r *http.Response) {
	var (
		url    string
		method string
	)
	if r.Request != nil {
		url = r.Request.URL.String()
		method = r.Request.Method
	}

	logrus.WithFields(logrus.Fields{
		"namespace":   "http_client",
		"url":         url,
		"method":      method,
		"status_code": r.StatusCode,
	}).Info("LOG_WORKER_HTTP_RESPONSE")
}

func defaultCheckRetry(method string) retryablehttp.CheckRetry {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		switch method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}

		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if err == nil {
			return false, nil
		}

		if isPreWriteNetworkError(err) {
			return true, nil
		}

		return false, nil
	}
}

func isPreWriteNetworkError(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return true
	}

	return false
}

func NewHttpClient(options *sdk_dto.HttpClientOptions) (res sdk_opt.Response) {
	parser := NewParser()
	client := newRetryClient(options)

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
				logrus.Warnf("NewHttpClient: body %T is not form-encodable for %s; sending fmt %%v representation", v, options.Url)
				bodyReader = strings.NewReader(fmt.Sprintf("%v", v))
			}

		default:
			logrus.Warnf("NewHttpClient: body %T serialized via fmt %%v for %s (no JSON Content-Type set)", v, options.Url)
			bodyReader = strings.NewReader(fmt.Sprintf("%v", v))
		}
	}

	ctx := options.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultHTTPTimeout)
		defer cancel()
	}

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
		bodyBytes = bodyBytes[:maxResponseSize]
		logrus.Warnf("HTTP Response truncated: exceeded %d bytes limit for URL: %s", maxResponseSize, options.Url)
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.WithFields(logrus.Fields{
			"namespace":        "http_client",
			"url":              options.Url,
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

	if errors.Is(err, context.DeadlineExceeded) || isTimeoutError(err) {
		res.StatCode = http.StatusGatewayTimeout
		res.ErrMsg = err.Error()

		return
	}

	if errors.Is(err, context.Canceled) {
		res.StatCode = http.StatusRequestTimeout
		res.ErrMsg = err.Error()

		return
	}

	var netErr *net.OpError
	if errors.As(err, &netErr) || errors.Is(err, net.ErrClosed) || errors.Is(err, net.ErrWriteToConnected) {
		res.StatCode = http.StatusBadGateway
		res.ErrMsg = err.Error()

		return
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.ErrNoProgress) {
		res.StatCode = http.StatusBadGateway
		res.ErrMsg = err.Error()
		res.Message = err.Error()

		return
	}

	res.StatCode = http.StatusBadRequest
	res.ErrMsg = err.Error()
	res.Message = err.Error()
	return
}

func isTimeoutError(err error) bool {
	var netErr net.Error

	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	return false
}
