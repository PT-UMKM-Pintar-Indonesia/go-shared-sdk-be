package sdk_helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices" // Go 1.21+
	"strings"
	"sync"
	"time"
)

const (
	VerifyURL          = "https://www.google.com/recaptcha/api/siteverify"
	DefaultTimeout     = 10 * time.Second
	DefaultMaxRespSize = 10 * 1024 // 10KB
)

var (
	ErrRateLimited = errors.New("rate limit exceeded")
	errorMap       = map[string]error{
		"missing-input-secret":   errors.New("missing secret parameter"),
		"invalid-input-secret":   errors.New("invalid secret parameter"),
		"missing-input-response": errors.New("missing response token"),
		"invalid-input-response": errors.New("invalid or malformed token"),
		"bad-request":            errors.New("bad request"),
		"timeout-or-duplicate":   errors.New("token expired or already used"),
	}
)

type (
	Config struct {
		Secret      string
		HTTPClient  *http.Client
		Timeout     time.Duration
		VerifyURL   string
		MaxRespSize int64
		RateLimit   int
		RateWindow  time.Duration
	}

	VerifyOptions struct {
		Response         string
		RemoteIP         string
		ExpectedHostname []string
		ExpectedAction   []string
		MinScore         float64
	}

	Response struct {
		Success     bool     `json:"success"`
		ChallengeTS string   `json:"challenge_ts"`
		Hostname    string   `json:"hostname"`
		Score       float64  `json:"score"`
		Action      string   `json:"action"`
		ErrorCodes  []string `json:"error-codes"`
	}

	Client struct {
		secret        string
		verifyURL     string
		httpClient    *http.Client
		maxRespSize   int64
		mu            sync.RWMutex
		rateLimit     int
		rateWindow    time.Duration
		requestCounts map[string][]time.Time
		stopCleanup   chan struct{}
	}
)

func NewGoogleRecaptcha(config *Config) (*Client, error) {
	if config.Secret == "" {
		return nil, errors.New("secret key is required")
	}

	c := &Client{
		secret:        config.Secret,
		verifyURL:     config.VerifyURL,
		maxRespSize:   config.MaxRespSize,
		requestCounts: make(map[string][]time.Time),
		rateLimit:     config.RateLimit,
		rateWindow:    config.RateWindow,
		stopCleanup:   make(chan struct{}),
	}

	if c.verifyURL == "" {
		c.verifyURL = VerifyURL
	}
	if c.maxRespSize == 0 {
		c.maxRespSize = DefaultMaxRespSize
	}
	if c.rateLimit == 0 {
		c.rateLimit = 100
	}
	if c.rateWindow == 0 {
		c.rateWindow = time.Minute
	}

	if config.HTTPClient != nil {
		c.httpClient = config.HTTPClient
	} else {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}
		c.httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	go c.cleanupLoop()
	return c, nil
}

func (c *Client) checkRateLimit(ip string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-c.rateWindow)

	c.requestCounts[ip] = slices.DeleteFunc(c.requestCounts[ip], func(t time.Time) bool {
		return t.Before(cutoff)
	})

	if len(c.requestCounts[ip]) >= c.rateLimit {
		return ErrRateLimited
	}

	c.requestCounts[ip] = append(c.requestCounts[ip], now)
	return nil
}

func (c *Client) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-c.rateWindow)

			for ip, times := range c.requestCounts {
				newTimes := slices.DeleteFunc(times, func(t time.Time) bool {
					return t.Before(cutoff)
				})
				if len(newTimes) == 0 {
					delete(c.requestCounts, ip)
				} else {
					c.requestCounts[ip] = newTimes
				}
			}

			c.mu.Unlock()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Client) Verify(ctx context.Context, opts VerifyOptions) (*Response, error) {
	if opts.Response == "" {
		return nil, errorMap["missing-input-response"]
	}

	if opts.RemoteIP != "" {
		if err := c.checkRateLimit(opts.RemoteIP); err != nil {
			return nil, err
		}
	}

	val := url.Values{"secret": {c.secret}, "response": {opts.Response}}
	if opts.RemoteIP != "" {
		val.Set("remoteip", opts.RemoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.verifyURL, strings.NewReader(val.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var res Response
	if err := json.NewDecoder(io.LimitReader(resp.Body, c.maxRespSize)).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.ErrorCodes) > 0 {
		return &res, c.parseErrors(res.ErrorCodes)
	}

	return &res, nil
}

func (c *Client) parseErrors(codes []string) error {
	var errs []error
	for _, code := range codes {
		if err, ok := errorMap[code]; ok {
			errs = append(errs, err)
		} else {
			errs = append(errs, fmt.Errorf("unknown error: %s", code))
		}
	}
	return errors.Join(errs...)
}

func (r *Response) ValidateV3(minScore float64, expectedActions, expectedHosts []string) error {
	if !r.Success {
		return errors.New("verification failed")
	}

	if r.Score < minScore {
		return fmt.Errorf("score %.2f below threshold %.2f", r.Score, minScore)
	}

	if len(expectedActions) > 0 && !slices.Contains(expectedActions, r.Action) {
		return fmt.Errorf("invalid action: %s", r.Action)
	}

	if len(expectedHosts) > 0 && !slices.Contains(expectedHosts, r.Hostname) {
		return fmt.Errorf("invalid hostname: %s", r.Hostname)
	}

	return nil
}

func (c *Client) Close() {
	close(c.stopCleanup)
	c.httpClient.CloseIdleConnections()
}
