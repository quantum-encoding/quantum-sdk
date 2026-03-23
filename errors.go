package qai

import (
	"fmt"
	"net/http"
)

// ApiError is an alias for APIError (sdk-graph canonical name).
type ApiError = APIError

// Error is an alias for APIError representing SDK errors (sdk-graph canonical name).
// In Go, errors are returned as the built-in error interface; APIError is the concrete type.
type Error = APIError

// APIError is returned when the API responds with a non-2xx status code.
type APIError struct {
	// StatusCode is the HTTP status code from the response.
	StatusCode int `json:"-"`

	// Code is the error type from the API (e.g. "invalid_request", "rate_limit").
	Code string `json:"code"`

	// Message is the human-readable error description.
	Message string `json:"message"`

	// RequestID is the unique request identifier from the X-QAI-Request-Id header.
	RequestID string `json:"-"`
}

// apiErrorBody matches the JSON error envelope from the API.
type apiErrorBody struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("qai: %d %s: %s (request_id=%s)", e.StatusCode, e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("qai: %d %s: %s", e.StatusCode, e.Code, e.Message)
}

// IsRateLimit returns true if the error is a 429 rate limit response.
func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsAuth returns true if the error is a 401 or 403 authentication/authorization failure.
func (e *APIError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true if the error is a 404 not found response.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimitError checks whether an error is a rate limit APIError.
func IsRateLimitError(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.IsRateLimit()
	}
	return false
}

// IsAuthError checks whether an error is an authentication APIError.
func IsAuthError(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.IsAuth()
	}
	return false
}

// IsNotFoundError checks whether an error is a not found APIError.
func IsNotFoundError(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.IsNotFound()
	}
	return false
}
