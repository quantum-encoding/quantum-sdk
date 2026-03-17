package qai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	// DefaultBaseURL is the production Quantum AI API endpoint.
	DefaultBaseURL = "https://api.quantumencoding.ai"

	// TicksPerUSD is the number of ticks in one US dollar (10 billion).
	TicksPerUSD int64 = 10_000_000_000
)

// Client is the Quantum AI API client.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// New creates a new Quantum AI API client.
//
// The apiKey is sent as a Bearer token on every request.
// Use functional options to customise the base URL, HTTP client, or timeout.
//
//	client := qai.New("qai_key_xxx",
//	    qai.WithBaseURL("http://localhost:8080"),
//	    qai.WithTimeout(30 * time.Second),
//	)
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// responseMeta holds common response metadata parsed from HTTP headers.
type responseMeta struct {
	CostTicks int64
	RequestID string
	Model     string
}

// doJSON sends a JSON request and decodes the JSON response into dst.
// It returns the response metadata (cost ticks, request ID) from headers.
func (c *Client) doJSON(ctx context.Context, method, path string, body any, dst any) (*responseMeta, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("qai: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("qai: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qai: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	meta := &responseMeta{
		RequestID: resp.Header.Get("X-QAI-Request-Id"),
		Model:     resp.Header.Get("X-QAI-Model"),
	}
	if v := resp.Header.Get("X-QAI-Cost-Ticks"); v != "" {
		meta.CostTicks, _ = strconv.ParseInt(v, 10, 64)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return meta, parseAPIError(resp, meta.RequestID)
	}

	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			return meta, fmt.Errorf("qai: decode response: %w", err)
		}
	}

	return meta, nil
}

// doStreamRaw sends a JSON request expecting an SSE (text/event-stream) response.
// It returns the raw http.Response for the caller to read SSE events from.
// The caller is responsible for closing the response body.
func (c *Client) doStreamRaw(ctx context.Context, path string, body any) (*http.Response, *responseMeta, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("qai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("qai: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Use a client without timeout for streaming — context controls cancellation.
	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("qai: POST %s: %w", path, err)
	}

	meta := &responseMeta{
		RequestID: resp.Header.Get("X-QAI-Request-Id"),
		Model:     resp.Header.Get("X-QAI-Model"),
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, meta, parseAPIError(resp, meta.RequestID)
	}

	return resp, meta, nil
}

// parseAPIError reads the response body and returns an *APIError.
func parseAPIError(resp *http.Response, requestID string) *APIError {
	body, _ := io.ReadAll(resp.Body)

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RequestID:  requestID,
		Code:       http.StatusText(resp.StatusCode),
		Message:    string(body),
	}

	var errBody apiErrorBody
	if json.Unmarshal(body, &errBody) == nil && errBody.Error.Message != "" {
		apiErr.Message = errBody.Error.Message
		if errBody.Error.Code != "" {
			apiErr.Code = errBody.Error.Code
		} else if errBody.Error.Type != "" {
			apiErr.Code = errBody.Error.Type
		}
	}

	return apiErr
}
