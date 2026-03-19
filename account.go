package qai

import (
	"context"
	"fmt"
)

// BalanceResponse is the response from an account balance check.
type BalanceResponse struct {
	// UserID is the account's user identifier.
	UserID string `json:"user_id"`

	// CreditTicks is the current balance in ticks.
	CreditTicks int64 `json:"credit_ticks"`

	// CreditUSD is the current balance in US dollars.
	CreditUSD float64 `json:"credit_usd"`

	// TicksPerUSD is the number of ticks in one US dollar.
	TicksPerUSD int64 `json:"ticks_per_usd"`
}

// UsageEntry is a single entry from the usage ledger.
type UsageEntry struct {
	// ID is the ledger entry identifier.
	ID string `json:"id"`

	// RequestID is the API request that generated this entry.
	RequestID string `json:"request_id,omitempty"`

	// Model is the model used for this request.
	Model string `json:"model,omitempty"`

	// Provider is the upstream provider.
	Provider string `json:"provider,omitempty"`

	// Endpoint is the API endpoint called.
	Endpoint string `json:"endpoint,omitempty"`

	// DeltaTicks is the cost of this request in ticks (negative = debit).
	DeltaTicks int64 `json:"delta_ticks,omitempty"`

	// BalanceAfter is the account balance after this transaction.
	BalanceAfter int64 `json:"balance_after,omitempty"`

	// InputTokens is the number of input tokens consumed.
	InputTokens int64 `json:"input_tokens,omitempty"`

	// OutputTokens is the number of output tokens generated.
	OutputTokens int64 `json:"output_tokens,omitempty"`

	// CreatedAt is the ISO 8601 timestamp.
	CreatedAt string `json:"created_at,omitempty"`
}

// UsageResponse is a paginated list of usage entries.
type UsageResponse struct {
	// Entries is the list of usage entries.
	Entries []UsageEntry `json:"entries"`

	// HasMore indicates whether more entries exist after this page.
	HasMore bool `json:"has_more"`

	// NextCursor is the pagination cursor for the next page.
	NextCursor string `json:"next_cursor,omitempty"`
}

// UsageQuery controls pagination for the usage endpoint.
type UsageQuery struct {
	// Limit is the maximum number of entries per page (default 20, max 100).
	Limit int `json:"limit,omitempty"`

	// StartAfter is the cursor from a previous response's NextCursor.
	StartAfter string `json:"start_after,omitempty"`
}

// UsageSummaryMonth is a monthly aggregate of usage.
type UsageSummaryMonth struct {
	// Month is the month in "YYYY-MM" format.
	Month string `json:"month"`

	// TotalRequests is the number of API requests in this month.
	TotalRequests int64 `json:"total_requests"`

	// TotalInputTokens is the total input tokens consumed.
	TotalInputTokens int64 `json:"total_input_tokens"`

	// TotalOutputTokens is the total output tokens generated.
	TotalOutputTokens int64 `json:"total_output_tokens"`

	// TotalCostUSD is the total cost in US dollars.
	TotalCostUSD float64 `json:"total_cost_usd"`

	// TotalMarginUSD is the total margin in US dollars.
	TotalMarginUSD float64 `json:"total_margin_usd"`

	// ByProvider is the per-provider breakdown (opaque JSON).
	ByProvider []any `json:"by_provider,omitempty"`
}

// UsageSummaryResponse is the response from the usage summary endpoint.
type UsageSummaryResponse struct {
	// Months is the list of monthly summaries.
	Months []UsageSummaryMonth `json:"months"`
}

// PricingEntry is a single entry in the pricing table.
// JSON keys use PascalCase to match the API response.
type PricingEntry struct {
	// Provider is the upstream provider name.
	Provider string `json:"Provider"`

	// Model is the model identifier.
	Model string `json:"Model"`

	// DisplayName is the human-readable model name.
	DisplayName string `json:"DisplayName"`

	// InputPerMillion is the cost per million input tokens in USD.
	InputPerMillion float64 `json:"InputPerMillion"`

	// OutputPerMillion is the cost per million output tokens in USD.
	OutputPerMillion float64 `json:"OutputPerMillion"`

	// CachedPerMillion is the cost per million cached tokens in USD.
	CachedPerMillion float64 `json:"CachedPerMillion"`
}

// PricingResponse is the response from the pricing endpoint (map of model_id to entry).
type PricingResponse struct {
	// Pricing is a map of model ID to pricing entry.
	Pricing map[string]PricingEntry `json:"pricing"`
}

// AccountBalance returns the current account credit balance.
func (c *Client) AccountBalance(ctx context.Context) (*BalanceResponse, error) {
	var resp BalanceResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/account/balance", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AccountUsage returns paginated usage history.
func (c *Client) AccountUsage(ctx context.Context, query *UsageQuery) (*UsageResponse, error) {
	path := "/qai/v1/account/usage"
	if query != nil {
		sep := "?"
		if query.Limit > 0 {
			path += fmt.Sprintf("%slimit=%d", sep, query.Limit)
			sep = "&"
		}
		if query.StartAfter != "" {
			path += fmt.Sprintf("%sstart_after=%s", sep, query.StartAfter)
		}
	}
	var resp UsageResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AccountUsageSummary returns monthly usage summaries.
// If months is 0, the server default is used.
func (c *Client) AccountUsageSummary(ctx context.Context, months int) (*UsageSummaryResponse, error) {
	path := "/qai/v1/account/usage/summary"
	if months > 0 {
		path = fmt.Sprintf("%s?months=%d", path, months)
	}
	var resp UsageSummaryResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AccountPricing returns the full pricing table (model_id → pricing entry).
func (c *Client) AccountPricing(ctx context.Context) (*PricingResponse, error) {
	var resp PricingResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/pricing", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
