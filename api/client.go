// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package api provides a flexible HTTP API client with retry logic,
// rate limiting, and authentication support.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/patdeg/common"
	"golang.org/x/time/rate"
)

// Client represents an HTTP API client
type Client struct {
	baseURL     string
	httpClient  *http.Client
	auth        Authenticator
	rateLimiter *rate.Limiter
	retryConfig *RetryConfig
	headers     map[string]string
	mu          sync.RWMutex
}

// ClientConfig configures the API client
type ClientConfig struct {
	BaseURL     string
	Timeout     time.Duration
	Auth        Authenticator
	RateLimit   int // Requests per second
	RetryConfig *RetryConfig
	Headers     map[string]string
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
	RetryOn     []int // HTTP status codes to retry on
}

// Authenticator provides authentication for requests
type Authenticator interface {
	// Authenticate adds authentication to a request
	Authenticate(req *http.Request) error

	// Refresh refreshes authentication credentials if needed
	Refresh(ctx context.Context) error
}

// Request represents an API request
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers map[string]string
	Body    interface{}
}

// Response represents an API response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Error represents an API error
type Error struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// NewClient creates a new API client
func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryConfig == nil {
		config.RetryConfig = DefaultRetryConfig()
	}

	client := &Client{
		baseURL: strings.TrimRight(config.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		auth:        config.Auth,
		retryConfig: config.RetryConfig,
		headers:     config.Headers,
	}

	if config.RateLimit > 0 {
		client.rateLimiter = rate.NewLimiter(rate.Limit(config.RateLimit), 1)
	}

	return client
}

// Do executes an API request
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// Apply rate limiting
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %v", err)
		}
	}

	// Build HTTP request
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	// Add authentication
	if c.auth != nil {
		if err := c.auth.Authenticate(httpReq); err != nil {
			return nil, fmt.Errorf("authentication failed: %v", err)
		}
	}

	// Execute with retry
	return c.doWithRetry(ctx, httpReq)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, query url.Values) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: "GET",
		Path:   path,
		Query:  query,
	})
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: "POST",
		Path:   path,
		Body:   body,
	})
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: "PUT",
		Path:   path,
		Body:   body,
	})
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: "DELETE",
		Path:   path,
	})
}

// SetHeader sets a default header
func (c *Client) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}

// buildRequest builds an HTTP request
func (c *Client) buildRequest(ctx context.Context, req *Request) (*http.Request, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + req.Path)
	if err != nil {
		return nil, err
	}

	if req.Query != nil {
		u.RawQuery = req.Query.Encode()
	}

	// Build body
	var body io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %v", err)
		}
		body = bytes.NewReader(jsonData)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Set default headers
	c.mu.RLock()
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}
	c.mu.RUnlock()

	// Set request headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set content type for body
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	return httpReq, nil
}

// doWithRetry executes a request with retry logic
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*Response, error) {
	var lastErr error
	wait := c.retryConfig.InitialWait

	// Store the original request body bytes if present
	var bodyBytes []byte
	if req.Body != nil {
		// Check if body is seekable
		if _, ok := req.Body.(io.Seeker); ok {
			// If seekable, we can reset it on retries
			// Continue with normal flow
		} else {
			// Read the entire body to store for retries
			var err error
			bodyBytes, err = io.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %v", err)
			}
			// Set the body to a reader of the bytes for the first attempt
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}

			// Increase wait time
			wait = time.Duration(float64(wait) * c.retryConfig.Multiplier)
			if wait > c.retryConfig.MaxWait {
				wait = c.retryConfig.MaxWait
			}

			common.Debug("[API] Retrying request (attempt %d/%d)", attempt, c.retryConfig.MaxRetries)
		}

		// Clone request for retry
		reqCopy := req.Clone(ctx)
		if req.Body != nil {
			if bodyBytes != nil {
				// Reset body using stored bytes
				reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			} else if seeker, ok := req.Body.(io.Seeker); ok {
				// Reset seekable body
				seeker.Seek(0, io.SeekStart)
				reqCopy.Body = req.Body
			}
		}

		// Execute request
		resp, err := c.httpClient.Do(reqCopy)
		if err != nil {
			lastErr = err
			continue
		}

		// Read response body
		bodyData, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %v", err)
			continue
		}

		response := &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       bodyData,
		}

		// Check if we should retry
		if c.shouldRetry(resp.StatusCode) && attempt < c.retryConfig.MaxRetries {
			lastErr = fmt.Errorf("received status %d", resp.StatusCode)
			continue
		}

		// Check for error status
		if resp.StatusCode >= 400 {
			var apiErr Error
			if err := json.Unmarshal(bodyData, &apiErr); err != nil {
				apiErr = Error{
					StatusCode: resp.StatusCode,
					Message:    string(bodyData),
				}
			} else {
				apiErr.StatusCode = resp.StatusCode
			}
			return response, &apiErr
		}

		return response, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %v", c.retryConfig.MaxRetries, lastErr)
}

// shouldRetry checks if a status code should trigger a retry
func (c *Client) shouldRetry(statusCode int) bool {
	for _, code := range c.retryConfig.RetryOn {
		if statusCode == code {
			return true
		}
	}
	return false
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
		RetryOn:     []int{429, 500, 502, 503, 504},
	}
}

// BearerTokenAuth implements bearer token authentication
type BearerTokenAuth struct {
	Token string
}

func (a *BearerTokenAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}

func (a *BearerTokenAuth) Refresh(ctx context.Context) error {
	// Token refresh logic would go here
	return nil
}

// BasicAuth implements HTTP basic authentication
type BasicAuth struct {
	Username string
	Password string
}

func (a *BasicAuth) Authenticate(req *http.Request) error {
	req.SetBasicAuth(a.Username, a.Password)
	return nil
}

func (a *BasicAuth) Refresh(ctx context.Context) error {
	// No refresh needed for basic auth
	return nil
}

// APIKeyAuth implements API key authentication
type APIKeyAuth struct {
	Key    string
	Header string
}

func (a *APIKeyAuth) Authenticate(req *http.Request) error {
	if a.Header == "" {
		a.Header = "X-API-Key"
	}
	req.Header.Set(a.Header, a.Key)
	return nil
}

func (a *APIKeyAuth) Refresh(ctx context.Context) error {
	// No refresh needed for API key auth
	return nil
}

// RESTClient provides a higher-level REST API client
type RESTClient struct {
	client *Client
}

// NewRESTClient creates a new REST client
func NewRESTClient(config ClientConfig) *RESTClient {
	return &RESTClient{
		client: NewClient(config),
	}
}

// GetJSON performs a GET request and decodes JSON response
func (r *RESTClient) GetJSON(ctx context.Context, path string, query url.Values, result interface{}) error {
	resp, err := r.client.Get(ctx, path, query)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PostJSON performs a POST request with JSON body and decodes response
func (r *RESTClient) PostJSON(ctx context.Context, path string, body, result interface{}) error {
	resp, err := r.client.Post(ctx, path, body)
	if err != nil {
		return err
	}

	if result != nil {
		return json.Unmarshal(resp.Body, result)
	}
	return nil
}

// PutJSON performs a PUT request with JSON body and decodes response
func (r *RESTClient) PutJSON(ctx context.Context, path string, body, result interface{}) error {
	resp, err := r.client.Put(ctx, path, body)
	if err != nil {
		return err
	}

	if result != nil {
		return json.Unmarshal(resp.Body, result)
	}
	return nil
}

// DeleteJSON performs a DELETE request and decodes response
func (r *RESTClient) DeleteJSON(ctx context.Context, path string, result interface{}) error {
	resp, err := r.client.Delete(ctx, path)
	if err != nil {
		return err
	}

	if result != nil && len(resp.Body) > 0 {
		return json.Unmarshal(resp.Body, result)
	}
	return nil
}

// Paginate handles paginated API responses
func (r *RESTClient) Paginate(ctx context.Context, path string, pageSize int, handler func(page interface{}) error) error {
	page := 1
	for {
		query := url.Values{
			"page":      []string{fmt.Sprintf("%d", page)},
			"page_size": []string{fmt.Sprintf("%d", pageSize)},
		}

		var result struct {
			Data     json.RawMessage `json:"data"`
			HasMore  bool            `json:"has_more"`
			NextPage int             `json:"next_page"`
		}

		if err := r.GetJSON(ctx, path, query, &result); err != nil {
			return err
		}

		if err := handler(result.Data); err != nil {
			return err
		}

		if !result.HasMore {
			break
		}

		page = result.NextPage
		if page == 0 {
			page++
		}
	}

	return nil
}
