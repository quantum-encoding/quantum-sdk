package qai

import "context"

// ModelInfo describes an available model.
type ModelInfo struct {
	// ID is the model identifier used in API requests.
	ID string `json:"id"`

	// Provider is the upstream provider (e.g. "anthropic", "xai", "openai").
	Provider string `json:"provider"`

	// DisplayName is the human-readable model name.
	DisplayName string `json:"display_name"`

	// InputPerMillion is the cost per million input tokens in USD.
	InputPerMillion float64 `json:"input_per_million"`

	// OutputPerMillion is the cost per million output tokens in USD.
	OutputPerMillion float64 `json:"output_per_million"`
}

// modelsResponse is the API response wrapper for model listing.
type modelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// pricingResponse is the API response wrapper for pricing.
type pricingResponse struct {
	Pricing []PricingInfo `json:"pricing"`
}

// PricingInfo contains pricing details for a model.
type PricingInfo struct {
	// ID is the model identifier.
	ID string `json:"id"`

	// Provider is the upstream provider.
	Provider string `json:"provider"`

	// DisplayName is the human-readable model name.
	DisplayName string `json:"display_name"`

	// InputPerMillion is the cost per million input tokens in USD.
	InputPerMillion float64 `json:"input_per_million"`

	// OutputPerMillion is the cost per million output tokens in USD.
	OutputPerMillion float64 `json:"output_per_million"`
}

// ListModels returns all available models with provider and pricing information.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	var resp modelsResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/models", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Models, nil
}

// GetPricing returns the complete pricing table for all models.
func (c *Client) GetPricing(ctx context.Context) ([]PricingInfo, error) {
	var resp pricingResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/pricing", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Pricing, nil
}
