// Package qai provides a Go client for the Quantum AI API.
//
// The client supports text generation (with streaming), image/video/audio generation,
// embeddings, RAG search, and model listing — all through a single API endpoint.
//
// Usage:
//
//	client := qai.New("your-api-key")
//	resp, err := client.Chat(ctx, &qai.ChatRequest{
//	    Model:    "claude-sonnet-4-6",
//	    Messages: []qai.ChatMessage{{Role: "user", Content: "Hello!"}},
//	})
package qai

import (
	"net/http"
	"time"
)

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client for all requests.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.http = hc
	}
}

// WithTimeout sets the timeout for non-streaming HTTP requests.
// Streaming requests use the context deadline instead.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.http.Timeout = d
	}
}
