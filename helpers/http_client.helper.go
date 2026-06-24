package sdk_helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type RequestLogFunc func(req *http.Request, body []byte)

type ResponseLogFunc func(res *http.Response, body []byte)

type HttpClientOptions struct {
	MaxRetry            int
	RetryWaitMin        time.Duration
	RetryWaitMax        time.Duration
	MaxResponseBodySize int64
	Logger              *slog.Logger
	OnRequestLog        RequestLogFunc
	OnResponseLog       ResponseLogFunc
}

type RequestOptions struct {
	Method  string
	URL     string
	Body    any
	Headers map[string]string
}

type HttpClient struct {
	client *retryablehttp.Client
	opt    HttpClientOptions
}

func NewHttpClient(opt HttpClientOptions) *HttpClient {
	if opt.MaxResponseBodySize == 0 {
		opt.MaxResponseBodySize = 10 * 1024 * 1024
	}
	if opt.Logger == nil {
		opt.Logger = slog.Default()
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = opt.MaxRetry
	retryClient.RetryWaitMin = opt.RetryWaitMin
	retryClient.RetryWaitMax = opt.RetryWaitMax
	retryClient.Logger = nil

	return &HttpClient{
		client: retryClient,
		opt:    opt,
	}
}

func (c *HttpClient) Do(ctx context.Context, reqOpt RequestOptions) ([]byte, error) {
	bodyReader, contentType, err := c.buildRequestBody(reqOpt.Body)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if bodyReader != nil {
		if br, ok := bodyReader.(*bytes.Reader); ok {
			bodyBytes = make([]byte, br.Len())
			currPos, _ := br.Seek(0, io.SeekCurrent)
			_, _ = br.ReadAt(bodyBytes, 0)
			_, _ = br.Seek(currPos, io.SeekStart)
		}
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, reqOpt.Method, reqOpt.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range reqOpt.Headers {
		req.Header.Set(k, v)
	}

	if c.opt.OnRequestLog != nil {
		c.opt.OnRequestLog(req.Request, bodyBytes)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limitReader := io.LimitReader(resp.Body, c.opt.MaxResponseBodySize)
	resBody, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, err
	}

	if c.opt.OnResponseLog != nil {
		c.opt.OnResponseLog(resp, resBody)
	}

	if resp.StatusCode >= 400 {
		return resBody, fmt.Errorf("http_error_%d", resp.StatusCode)
	}

	return resBody, nil
}

func (c *HttpClient) buildRequestBody(data any) (io.Reader, string, error) {
	if data == nil {
		return nil, "", nil
	}

	switch v := data.(type) {
	case []byte:
		return bytes.NewReader(v), "application/octet-stream", nil
	case string:
		return bytes.NewReader([]byte(v)), "text/plain", nil
	case url.Values:
		return bytes.NewReader([]byte(v.Encode())), "application/x-www-form-urlencoded", nil
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(jsonData), "application/json", nil
	}
}
