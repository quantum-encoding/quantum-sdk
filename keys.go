package qai

import (
	"context"
	"fmt"
)

// CreateKeyRequest is the request body for creating a scoped API key.
type CreateKeyRequest struct {
	// Name is a human-readable name for the key (required).
	Name string `json:"name"`

	// Endpoints is a list of allowed endpoint prefixes (e.g. ["/qai/v1/chat"]).
	// If empty, all endpoints are allowed.
	Endpoints []string `json:"endpoints,omitempty"`

	// SpendCapUSD is the maximum USD spend for this key (converted to ticks server-side).
	SpendCapUSD float64 `json:"spend_cap_usd,omitempty"`

	// RateLimit is the maximum requests per minute for this key.
	RateLimit int `json:"rate_limit,omitempty"`
}

// CreateKeyResponse is the response from creating an API key.
type CreateKeyResponse struct {
	// Key is the raw API key string. It is only shown once at creation time.
	Key string `json:"key"`

	// Details contains the key metadata.
	Details KeyDetails `json:"details"`
}

// KeyDetails contains metadata about an API key.
type KeyDetails struct {
	// ID is the key identifier.
	ID string `json:"id"`

	// Name is the human-readable key name.
	Name string `json:"name"`

	// Prefix is the visible prefix of the key (e.g. "qai_...abc").
	Prefix string `json:"prefix"`

	// Endpoints is the list of allowed endpoint prefixes.
	Endpoints []string `json:"endpoints,omitempty"`

	// SpendCapTicks is the maximum spend in ticks.
	SpendCapTicks int64 `json:"spend_cap_ticks,omitempty"`

	// SpentTicks is the total ticks spent by this key.
	SpentTicks int64 `json:"spent_ticks,omitempty"`

	// RateLimit is the maximum requests per minute.
	RateLimit int `json:"rate_limit,omitempty"`

	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// Revoked indicates whether the key has been revoked.
	Revoked bool `json:"revoked,omitempty"`
}

// ListKeysResponse is the response from listing API keys.
type ListKeysResponse struct {
	// Keys is the list of API keys for the authenticated user.
	Keys []KeyDetails `json:"keys"`
}

// CreateKey creates a new scoped API key.
// The returned key string is only shown once — store it securely.
func (c *Client) CreateKey(ctx context.Context, req *CreateKeyRequest) (*CreateKeyResponse, error) {
	var resp CreateKeyResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/keys", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListKeys returns all API keys for the authenticated user.
func (c *Client) ListKeys(ctx context.Context) (*ListKeysResponse, error) {
	var resp ListKeysResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/keys", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// RevokeKey revokes an API key by its ID.
func (c *Client) RevokeKey(ctx context.Context, id string) error {
	_, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/qai/v1/keys/%s", id), nil, nil)
	return err
}
