package sdk_dto

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type (
	HttpClientConfigs struct {
		RetryWaitMin time.Duration
		RetryWaitMax time.Duration
		RetryMax     int
		CheckRetry   retryablehttp.CheckRetry
		Backoff      retryablehttp.Backoff
		ErrorHandler retryablehttp.ErrorHandler
		PrepareRetry retryablehttp.PrepareRetry
	}

	HttpClientOptions struct {
		Ctx       context.Context
		Url       string
		Method    string
		Body      any
		Headers   map[string]string
		Configs   HttpClientConfigs
		Transport http.RoundTripper
		standart  *http.Client
	}
)
