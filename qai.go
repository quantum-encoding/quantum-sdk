package qai

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
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
		// 60s used to abort long buffered media generation (image/video return a
		// single JSON blob only when the provider finishes — no bytes flow during
		// generation). 600s clears the backend's 5-minute media deadline, so the
		// server returns a proper error before the client gives up. Override with
		// WithTimeout. Streaming uses a separate no-timeout client.
		http: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// responseMeta holds common response metadata parsed from HTTP headers.
type responseMeta struct {
	CostTicks    int64
	RequestID    string
	Model        string
	BalanceAfter int64 // X-QAI-Balance-After: wallet balance after this call (signed — claw-back can make it negative)
}

// idempotentRequest is implemented by billing-bearing request structs that
// carry a caller-overridable Idempotency-Key. When the caller hasn't set one,
// doJSON/doStreamRaw auto-generates a random key so a network retry of the
// same in-flight request is deduped server-side instead of double-charging.
//
// The field is `json:"-"` so it never appears in the request body; it only
// surfaces as the Idempotency-Key HTTP header.
type idempotentRequest interface {
	idempotencyKey() string
}

// newIdempotencyKey returns a fresh 128-bit hex idempotency key. Uses
// crypto/rand so the keys are unguessable (an attacker who could predict
// them could pre-seat keys and block legitimate retries).
func newIdempotencyKey() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand should never fail on a sane host; fall back to a
		// time-derived key so the header is still set rather than dropped.
		return fmt.Sprintf("qai_%x", time.Now().UnixNano())
	}
	return "qai_" + hex.EncodeToString(b[:])
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

	// Auto-attach an Idempotency-Key on billing-bearing POSTs so a retry
	// of the same in-flight request is deduped server-side. Callers can
	// override by setting IdempotencyKey on the request struct.
	if method == http.MethodPost {
		if ik, ok := body.(idempotentRequest); ok {
			req.Header.Set("Idempotency-Key", ik.idempotencyKey())
		}
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
	// X-QAI-Balance-After is the wallet balance after this call. It is
	// signed: a claw-back (refund reversal) can push it negative. Absent
	// on routes that don't touch the wallet (models list, etc.) — leave 0.
	if v := resp.Header.Get("X-QAI-Balance-After"); v != "" {
		meta.BalanceAfter, _ = strconv.ParseInt(v, 10, 64)
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

	// Billing-bearing streaming routes (chat) get an auto-generated
	// Idempotency-Key too — a stream is still a chargeable generation.
	if ik, ok := body.(idempotentRequest); ok {
		req.Header.Set("Idempotency-Key", ik.idempotencyKey())
	}

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
	if v := resp.Header.Get("X-QAI-Balance-After"); v != "" {
		meta.BalanceAfter, _ = strconv.ParseInt(v, 10, 64)
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
