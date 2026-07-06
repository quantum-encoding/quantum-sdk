package qai

import (
	"errors"
	"fmt"
	"net/http"
)

// ApiError is an alias for APIError (sdk-graph canonical name).
type ApiError = APIError

// Error is an alias for APIError representing SDK errors (sdk-graph canonical name).
// In Go, errors are returned as the built-in error interface; APIError is the concrete type.
type Error = APIError

// CodeInsufficientBalance is the stable machine-readable code the backend
// emits on a 402 from billing-bearing routes (chat, images, video, tts,
// missions, etc.). It maps to HTTP 402 Payment Required and means the
// caller's wallet cannot cover the pessimistic next-step cost estimate.
// Switch on this code rather than substring-matching the message.
//
// Mirror of internal/server/errors.go CodeInsufficientBalance.
const CodeInsufficientBalance = "INSUFFICIENT_BALANCE"

// ErrInsufficientBalance is a typed sentinel returned (wrapped) when the
// gateway responds 402 with CodeInsufficientBalance. Use errors.Is(err,
// ErrInsufficientBalance) to detect "top up your wallet and retry" in
// generic code paths where you can't type-assert *APIError.
//
// Note: doJSON returns *APIError directly (not wrapped); callers can use
// either IsInsufficientBalance(err) or errors.Is(err, ErrInsufficientBalance)
// — the helper covers both the typed sentinel and the raw *APIError 402.
var ErrInsufficientBalance = errors.New("qai: insufficient balance")

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

// IsInsufficientBalance returns true if the error is a 402 Payment Required
// response carrying the backend's INSUFFICIENT_BALANCE code — i.e. the
// caller's wallet cannot cover the call. The gateway emits this from
// billing-bearing routes (chat, images/generate, video, tts, missions) via
// the StepGuard pre-flight check. On a 402 the request was NOT executed,
// so it is safe to top up and retry the same request (with the same
// Idempotency-Key).
//
// Matches on EITHER the HTTP status (402) OR the stable code string, so it
// keeps working even if a route emits 402 with a legacy type field but no
// code, or vice-versa.
func (e *APIError) IsInsufficientBalance() bool {
	if e.StatusCode == http.StatusPaymentRequired {
		return true
	}
	return e.Code == CodeInsufficientBalance
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

// IsInsufficientBalance reports whether err is a 402 / INSUFFICIENT_BALANCE
// response from the gateway. It recognises both a raw *APIError returned by
// doJSON and a wrapped ErrInsufficientBalance sentinel, so generic call
// sites can branch without type-asserting:
//
//	if qai.IsInsufficientBalance(err) { /* prompt user to top up, then retry */ }
func IsInsufficientBalance(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInsufficientBalance) {
		return true
	}
	if e, ok := err.(*APIError); ok {
		return e.IsInsufficientBalance()
	}
	return false
}
