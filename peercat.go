// Package peercat provides the official Go SDK for the PeerCat AI image generation API.
//
// # Quick Start
//
//	client, err := peercat.New("pcat_live_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	result, err := client.Generate(ctx, &peercat.GenerateParams{
//	    Prompt: "A beautiful sunset over mountains",
//	    Model:  "stable-diffusion-xl",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Image URL:", result.ImageURL)
//
// # Configuration
//
//	client, err := peercat.New("pcat_live_xxx",
//	    peercat.WithBaseURL("https://custom.api.url"),
//	    peercat.WithTimeout(30 * time.Second),
//	    peercat.WithMaxRetries(5),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Demo Mode
//
// Use demo mode to test without spending credits:
//
//	result, err := client.Generate(ctx, &peercat.GenerateParams{
//	    Prompt: "Test prompt",
//	    Mode:   peercat.ModeDemo,
//	})
package peercat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default API base URL
	DefaultBaseURL = "https://api.peerc.at"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 60 * time.Second
	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3
	// Version is the SDK version
	Version = "0.1.0"
)

// Client is the PeerCat API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// Option is a functional option for configuring the client
type Option func(*Client)

// ErrEmptyAPIKey is returned when an empty API key is provided
var ErrEmptyAPIKey = fmt.Errorf("peercat: API key is required")

// New creates a new PeerCat client with the given API key
// Returns an error if apiKey is empty
func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}

	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		maxRetries: DefaultMaxRetries,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// WithBaseURL sets a custom base URL
// Trailing slashes are automatically trimmed to prevent double-slash issues
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(retries int) Option {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// Generate creates an image from a text prompt
func (c *Client) Generate(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	var result GenerateResult
	if err := c.post(ctx, "/v1/generate", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetModels returns available image generation models
func (c *Client) GetModels(ctx context.Context) ([]Model, error) {
	var response ModelsResponse
	if err := c.get(ctx, "/v1/models", nil, &response); err != nil {
		return nil, err
	}
	return response.Models, nil
}

// GetPrices returns current pricing for all models
func (c *Client) GetPrices(ctx context.Context) (*PriceResponse, error) {
	var response PriceResponse
	if err := c.get(ctx, "/v1/price", nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetBalance returns current credit balance
func (c *Client) GetBalance(ctx context.Context) (*Balance, error) {
	var balance Balance
	if err := c.get(ctx, "/v1/balance", nil, &balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetHistory returns usage history
func (c *Client) GetHistory(ctx context.Context, params *HistoryParams) (*HistoryResponse, error) {
	query := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			query.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			query.Set("offset", strconv.Itoa(params.Offset))
		}
	}

	var response HistoryResponse
	if err := c.get(ctx, "/v1/history", query, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateKey creates a new API key (requires wallet signature)
func (c *Client) CreateKey(ctx context.Context, params *CreateKeyParams) (*CreateKeyResult, error) {
	var result CreateKeyResult
	if err := c.post(ctx, "/v1/keys", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListKeys returns all API keys for the authenticated wallet
func (c *Client) ListKeys(ctx context.Context) (*KeysResponse, error) {
	var response KeysResponse
	if err := c.get(ctx, "/v1/keys", nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// RevokeKey revokes an API key
func (c *Client) RevokeKey(ctx context.Context, keyID string) error {
	var response successResponse
	return c.delete(ctx, "/v1/keys/"+keyID, &response)
}

// UpdateKeyName updates an API key's name
func (c *Client) UpdateKeyName(ctx context.Context, keyID, name string) error {
	var response successResponse
	return c.patch(ctx, "/v1/keys/"+keyID, map[string]string{"name": name}, &response)
}

// SubmitPrompt submits a prompt for on-chain payment
func (c *Client) SubmitPrompt(ctx context.Context, params *SubmitPromptParams) (*PromptSubmission, error) {
	var result PromptSubmission
	if err := c.post(ctx, "/v1/prompts", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOnChainStatus returns the status of an on-chain generation
func (c *Client) GetOnChainStatus(ctx context.Context, txSignature string) (*OnChainGenerationStatus, error) {
	var status OnChainGenerationStatus
	if err := c.get(ctx, "/v1/generate/"+txSignature, nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Internal HTTP methods

func (c *Client) get(ctx context.Context, path string, query url.Values, result interface{}) error {
	return c.request(ctx, http.MethodGet, path, query, nil, result)
}

func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	return c.request(ctx, http.MethodPost, path, nil, body, result)
}

func (c *Client) patch(ctx context.Context, path string, body, result interface{}) error {
	return c.request(ctx, http.MethodPatch, path, nil, body, result)
}

func (c *Client) delete(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodDelete, path, nil, nil, result)
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, body, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := c.doRequest(ctx, method, path, query, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (except rate limits)
		if apiErr, ok := err.(*Error); ok {
			if apiErr.Status >= 400 && apiErr.Status < 500 && apiErr.Status != 429 {
				return err
			}
		}

		// Don't retry on context errors
		if ctx.Err() != nil {
			return err
		}

		// Exponential backoff (or use Retry-After for rate limits)
		if attempt < c.maxRetries {
			delay := time.Duration(1<<attempt) * time.Second
			if delay > 10*time.Second {
				delay = 10 * time.Second
			}

			// Use Retry-After header if available for rate limit errors
			if apiErr, ok := err.(*Error); ok && apiErr.RateLimitInfo != nil && apiErr.RateLimitInfo.RetryAfter > 0 {
				delay = time.Duration(apiErr.RateLimitInfo.RetryAfter) * time.Second
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, result interface{}) error {
	// Build URL
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	// Build request body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "peercat-go/"+Version)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse rate limit headers
	rateLimitInfo := parseRateLimitHeaders(resp.Header)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		var errResp apiErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return &Error{
				Status:        resp.StatusCode,
				Type:          "unknown",
				Code:          "parse_error",
				Message:       string(respBody),
				RateLimitInfo: rateLimitInfo,
			}
		}
		return errorFromResponse(resp.StatusCode, &errResp, rateLimitInfo)
	}

	// Parse successful response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// parseRateLimitHeaders parses rate limit information from response headers
func parseRateLimitHeaders(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}
	hasInfo := false

	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			info.Limit = v
			hasInfo = true
		}
	}

	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if v, err := strconv.Atoi(remaining); err == nil {
			info.Remaining = v
			hasInfo = true
		}
	}

	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if v, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.Reset = v
			hasInfo = true
		}
	}

	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if v, err := strconv.Atoi(retryAfter); err == nil {
			info.RetryAfter = v
			hasInfo = true
		}
	}

	if !hasInfo {
		return nil
	}
	return info
}

type successResponse struct {
	Success bool `json:"success"`
}
