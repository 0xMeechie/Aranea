package sdk

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	defaultUserAgent   = "aranea-sdk/0.1"
)

// HTTPClient wraps *http.Client with defaults useful for SDK callers:
// a request timeout, a default User-Agent, and a set of headers applied
// to every request unless the caller overrides them.
type HTTPClient struct {
	client    *http.Client
	userAgent string
	headers   map[string]string
}

// HTTPClientOption configures an HTTPClient.
type HTTPClientOption func(*HTTPClient)

// WithTimeout sets the per-request timeout. Zero disables the timeout.
func WithTimeout(d time.Duration) HTTPClientOption {
	return func(c *HTTPClient) { c.client.Timeout = d }
}

// WithUserAgent overrides the default User-Agent header. An empty string
// disables it.
func WithUserAgent(ua string) HTTPClientOption {
	return func(c *HTTPClient) { c.userAgent = ua }
}

// WithHeader adds (or replaces) a default header applied to every request.
func WithHeader(key, value string) HTTPClientOption {
	return func(c *HTTPClient) {
		if c.headers == nil {
			c.headers = map[string]string{}
		}
		c.headers[key] = value
	}
}

// WithHTTPClient replaces the underlying *http.Client. Use to inject a
// pre-configured transport, cookie jar, etc. The client's Timeout is
// preserved unless overridden by WithTimeout.
func WithHTTPClient(hc *http.Client) HTTPClientOption {
	return func(c *HTTPClient) {
		if hc != nil {
			c.client = hc
		}
	}
}

// NewHTTPClient builds an HTTPClient with sensible defaults applied,
// then layered options.
func NewHTTPClient(opts ...HTTPClientOption) *HTTPClient {
	hc := &HTTPClient{
		client:    &http.Client{Timeout: defaultHTTPTimeout},
		userAgent: defaultUserAgent,
	}
	for _, opt := range opts {
		opt(hc)
	}
	return hc
}

// Do executes req, applying the client's default User-Agent and headers
// where the request does not already set them.
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	for k, v := range c.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	return c.client.Do(req)
}

// Request builds and executes a request. body may be empty.
func (c *HTTPClient) Request(method, url, body string) (*http.Response, error) {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.Do(req)
}

// Fetch is a string-in/string-out convenience: it builds a request, runs it,
// reads the full body, and returns it as a string. The response status is
// returned alongside so callers can branch on it without inspecting headers.
func (c *HTTPClient) Fetch(method, url, body string) (status int, respBody string, err error) {
	resp, err := c.Request(method, url, body)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("read response: %w", err)
	}
	return resp.StatusCode, string(data), nil
}
