package qai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ContactRequest is the request body for the contact form.
type ContactRequest struct {
	// Name is the sender's name (required).
	Name string `json:"name"`

	// Email is the sender's email address (required).
	Email string `json:"email"`

	// Subject is the message subject.
	Subject string `json:"subject,omitempty"`

	// Message is the contact message (required).
	Message string `json:"message"`
}

// Contact sends a contact form message. This endpoint does not require authentication.
func (c *Client) Contact(ctx context.Context, req *ContactRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("qai: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/qai/v1/contact", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("qai: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// No Authorization header — contact is a public endpoint.

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("qai: POST /qai/v1/contact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       http.StatusText(resp.StatusCode),
			Message:    string(body),
		}
	}

	return nil
}
