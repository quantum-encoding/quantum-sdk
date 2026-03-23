package qai

import "context"

// CreditPack is a credit pack available for purchase.
type CreditPack struct {
	// ID is the unique pack identifier.
	ID string `json:"id"`

	// Name is the display name (e.g. "Starter Pack").
	Name string `json:"name,omitempty"`

	// PriceUSD is the price in US dollars.
	PriceUSD float64 `json:"price_usd"`

	// CreditTicks is the number of credit ticks included.
	CreditTicks int64 `json:"credit_ticks"`

	// Description of the pack.
	Description string `json:"description,omitempty"`
}

// CreditPacksResponse is the response from listing credit packs.
type CreditPacksResponse struct {
	// Packs is the list of available credit packs.
	Packs []CreditPack `json:"packs"`
}

// CreditPurchaseRequest is the request to purchase a credit pack.
type CreditPurchaseRequest struct {
	// PackID is the pack to purchase.
	PackID string `json:"pack_id"`

	// SuccessURL is the URL to redirect to after successful payment.
	SuccessURL string `json:"success_url,omitempty"`

	// CancelURL is the URL to redirect to if payment is cancelled.
	CancelURL string `json:"cancel_url,omitempty"`
}

// CreditPurchaseResponse is the response from purchasing a credit pack.
type CreditPurchaseResponse struct {
	// CheckoutURL is the URL to redirect the user to for payment.
	CheckoutURL string `json:"checkout_url"`
}

// CreditBalanceResponse is the response from checking credit balance.
type CreditBalanceResponse struct {
	// BalanceTicks is the balance in ticks.
	BalanceTicks int64 `json:"balance_ticks"`

	// BalanceUSD is the balance in US dollars.
	BalanceUSD float64 `json:"balance_usd"`
}

// CreditTier is a pricing tier.
type CreditTier struct {
	// Name of the tier.
	Name string `json:"name,omitempty"`

	// MinBalance is the minimum balance for this tier.
	MinBalance int64 `json:"min_balance,omitempty"`

	// DiscountPercent is the discount percentage.
	DiscountPercent float64 `json:"discount_percent,omitempty"`

	// Extra contains additional tier data.
	Extra map[string]any `json:"extra,omitempty"`
}

// CreditTiersResponse is the response from listing credit tiers.
type CreditTiersResponse struct {
	// Tiers is the list of available tiers.
	Tiers []CreditTier `json:"tiers"`
}

// DevProgramApplyRequest is the request to apply for the developer program.
type DevProgramApplyRequest struct {
	// UseCase describes the intended use case.
	UseCase string `json:"use_case"`

	// Company name (optional).
	Company string `json:"company,omitempty"`

	// ExpectedUSD is the expected monthly spend in USD (optional).
	ExpectedUSD float64 `json:"expected_usd,omitempty"`

	// Website URL (optional).
	Website string `json:"website,omitempty"`
}

// DevProgramApplyResponse is the response from a dev program application.
type DevProgramApplyResponse struct {
	// Status of the application (e.g. "submitted", "approved").
	Status string `json:"status"`
}

// CreditPacks lists available credit packs. No authentication required.
func (c *Client) CreditPacks(ctx context.Context) (*CreditPacksResponse, error) {
	var resp CreditPacksResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/credits/packs", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreditPurchase purchases a credit pack. Returns a checkout URL for payment.
func (c *Client) CreditPurchase(ctx context.Context, req *CreditPurchaseRequest) (*CreditPurchaseResponse, error) {
	var resp CreditPurchaseResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/credits/purchase", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreditBalance returns the current credit balance.
func (c *Client) CreditBalance(ctx context.Context) (*CreditBalanceResponse, error) {
	var resp CreditBalanceResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/credits/balance", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreditTiers lists available credit tiers. No authentication required.
func (c *Client) CreditTiers(ctx context.Context) (*CreditTiersResponse, error) {
	var resp CreditTiersResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/credits/tiers", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DevProgramApply applies for the developer program.
func (c *Client) DevProgramApply(ctx context.Context, req *DevProgramApplyRequest) (*DevProgramApplyResponse, error) {
	var resp DevProgramApplyResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/credits/dev-program", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
