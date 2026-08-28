package sdk_helper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"

	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type httpClient struct {
	standard *http.Client
}

func NewHttpClient(options *sdk_dto.HttpClientOptions) (res sdk_opt.Response) {
	client := retryablehttp.NewClient()
	parser := NewParser()

	standardClient := client.StandardClient()

	if options.Transport != nil {
		client.HTTPClient.Transport = options.Transport
	}

	hc := httpClient{standard: standardClient}

	client.RetryWaitMin = 3 * time.Second
	client.RetryWaitMax = 5 * time.Second
	client.RetryMax = 10
	client.Backoff = retryablehttp.DefaultBackoff

	hc.httpClientConfig(client, &options.Configs)

	client.RequestLogHook = func(l retryablehttp.Logger, r *http.Request, attempt int) {
		lgr, err := httputil.DumpRequestOut(r, sdk_cons.TRUE)
		if err != nil {
			logrus.WithError(err).Error("Failed to dump request")
			return
		}

		logrus.WithFields(logrus.Fields{
			"namespace":   "http_client",
			"url":         r.URL.String(),
			"method":      r.Method,
			"headers":     r.Header,
			"raw_request": TrimLogString(string(lgr)),
		}).Info("LOG_WORKER_HTTP_REQUEST")
	}

	client.ResponseLogHook = func(l retryablehttp.Logger, r *http.Response) {
		lgr, err := httputil.DumpResponse(r, sdk_cons.TRUE)
		if err != nil {
			logrus.WithError(err).Error("Failed to dump response")
			return
		}

		logrus.WithFields(logrus.Fields{
			"namespace":    "http_client",
			"url":          r.Request.URL.String(),
			"method":       r.Request.Method,
			"status_code":  r.StatusCode,
			"raw_response": TrimLogString(string(lgr)),
		}).Info("LOG_WORKER_HTTP_RESPONSE")
	}

	var bodyReader io.Reader

	switch v := options.Body.(type) {
	case nil:
		bodyReader = bytes.NewReader(nil)
	case io.Reader:
		bodyReader = v
	case string:
		bodyReader = strings.NewReader(v)
	case []byte:
		bodyReader = bytes.NewReader(v)
	default:
		if options.Headers != nil && strings.Contains(strings.ToLower(options.Headers["Content-Type"]), "json") {
			bodyBytes, err := parser.Marshal(v)
			if err != nil {
				res.StatCode = http.StatusBadRequest
				res.ErrMsg = fmt.Errorf("failed to marshal body: %w", err).Error()

				return
			}

			bodyReader = bytes.NewReader(bodyBytes)
		} else {
			bodyReader = strings.NewReader(fmt.Sprintf("%v", v))
		}
	}

	req, err := http.NewRequestWithContext(options.Ctx, options.Method, options.Url, bodyReader)
	if res = hc.httpClientError(err); res.StatCode >= http.StatusBadRequest {
		return
	}

	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	httpRes, httpErr := hc.standard.Do(req)
	if res = hc.httpClientError(httpErr); res.StatCode >= http.StatusBadRequest {
		return
	}
	defer httpRes.Body.Close()

	bodyBytes, err := io.ReadAll(httpRes.Body)
	if res = hc.httpClientError(err); res.StatCode >= http.StatusBadRequest {
		return
	}

	res.StatCode = http.StatusOK
	res.Message = "Network Success"
	res.Data = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return
}

func (h *httpClient) httpClientConfig(client *retryablehttp.Client, options *sdk_dto.HttpClientConfigs) {
	if options.RetryWaitMin != 0 {
		client.RetryWaitMin = options.RetryWaitMin
	}

	if options.RetryWaitMax != 0 {
		client.RetryWaitMax = options.RetryWaitMax
	}

	if options.RetryMax != 0 {
		client.RetryMax = options.RetryMax
	}

	if options.Backoff != nil {
		client.Backoff = options.Backoff
	}

	if options.ErrorHandler != nil {
		client.ErrorHandler = options.ErrorHandler
	}

	if options.PrepareRetry != nil {
		client.PrepareRetry = options.PrepareRetry
	}
}

func (h *httpClient) httpClientError(err error) (res sdk_opt.Response) {
	if err == nil {
		res.StatCode = http.StatusOK
		res.Message = "Network Success"
		return
	}

	var netErr *net.OpError

	if errors.As(err, &netErr) || errors.Is(err, net.ErrClosed) || errors.Is(err, net.ErrWriteToConnected) {
		res.StatCode = http.StatusRequestTimeout
		res.ErrMsg = err.Error()

		return
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.ErrNoProgress) {
		res.StatCode = http.StatusRequestTimeout
		res.ErrMsg = err.Error()

		return
	}

	res.StatCode = http.StatusBadRequest
	res.Message = err.Error()

	return
}
