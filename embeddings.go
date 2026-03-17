package qai

import "context"

// EmbedRequest is the request body for text embeddings.
type EmbedRequest struct {
	// Model is the embedding model (e.g. "text-embedding-3-small", "text-embedding-3-large").
	Model string `json:"model"`

	// Input is the list of texts to embed.
	Input []string `json:"input"`
}

// EmbedResponse is the response from text embedding.
type EmbedResponse struct {
	// Embeddings is the list of embedding vectors, one per input string.
	Embeddings [][]float64 `json:"embeddings"`

	// Model is the model that generated the embeddings.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// Embed generates text embeddings for the given inputs.
func (c *Client) Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	var resp EmbedResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/embeddings", req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.CostTicks == 0 {
		resp.CostTicks = meta.CostTicks
	}
	if resp.RequestID == "" {
		resp.RequestID = meta.RequestID
	}
	return &resp, nil
}
