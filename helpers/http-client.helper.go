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

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type httpClient struct {
	standart *http.Client
}

func NewHttpClient(options sdk_dto.HttpClientOptions) (res sdk_opt.Response) {
	client := retryablehttp.NewClient()

	parser := NewParser()
	standart := client.StandardClient()

	if options.Transport != nil {
		client.HTTPClient.Transport = options.Transport
	}

	httpClient := httpClient{standart: standart}
	httpRes := new(http.Response)
	httpErr := error(nil)

	client.RetryWaitMin = 3 * time.Second
	client.RetryWaitMax = 5 * time.Second
	client.RetryMax = 5
	client.Backoff = retryablehttp.DefaultBackoff

	httpClient.httpClientConfig(client, options.Configs)

	client.RequestLogHook = func(l retryablehttp.Logger, r *http.Request, i int) {
		lgr, err := httputil.DumpRequestOut(r, true)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Info(fmt.Sprintf("\nHTTP REQUEST: \n%s", string(lgr)))
	}

	client.ResponseLogHook = func(l retryablehttp.Logger, r *http.Response) {
		lgr, err := httputil.DumpResponse(r, true)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Info(fmt.Sprintf("\nHTTP RESPONSE: \n%s", string(lgr)))
	}

	if options.Body == nil {
		options.Body = bytes.NewReader(nil)
	} else {
		body, err := parser.Marshal(options.Body)
		if err != nil {
			res.StatCode = http.StatusBadRequest
			res.ErrMsg = err.Error()

			return
		}
		options.Body = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(options.Ctx, options.Method, options.Url, options.Body.(io.Reader))
	if res = httpClient.httpClientError(err); res.StatCode >= http.StatusBadRequest {
		return
	}

	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	if options.Transport != nil {
		options.Transport.RoundTrip(req)
	}

	httpRes, httpErr = standart.Do(req)
	if res = httpClient.httpClientError(httpErr); res.StatCode >= http.StatusBadRequest {
		return
	}

	body, err := io.ReadAll(httpRes.Body)
	if res = httpClient.httpClientError(err); res.StatCode >= http.StatusBadRequest {
		return
	}
	defer httpRes.Body.Close()

	if strings.Contains(strings.ToLower(string(body)), "html") {
		res.StatCode = http.StatusRequestTimeout
		res.ErrMsg = "Network Error"

		return
	}

	res.StatCode = http.StatusOK
	res.Data = io.NopCloser(bytes.NewBuffer(body))

	return
}

func (h httpClient) httpClientConfig(client *retryablehttp.Client, options sdk_dto.HttpClientConfigs) {
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

func (h httpClient) httpClientError(err error) (res sdk_opt.Response) {
	if err == nil {
		res.StatCode = http.StatusOK
		res.Message = "Success"
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
