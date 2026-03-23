package qai

import "context"

// SurrealRagProvider is an alias for SurrealRAGProviderInfo (sdk-graph canonical name).
type SurrealRagProvider = SurrealRAGProviderInfo

// SurrealRagProvidersResponse is an alias for SurrealRAGProvidersResponse (sdk-graph canonical name).
type SurrealRagProvidersResponse = SurrealRAGProvidersResponse

// SurrealRAGProviderInfo describes a documentation provider in the SurrealDB RAG system.
type SurrealRAGProviderInfo struct {
	// Provider is the provider name (e.g. "xai", "claude", "heygen").
	Provider string `json:"provider"`

	// Chunks is the number of embedded document chunks for this provider.
	Chunks int `json:"chunks"`
}

// SurrealRAGProvidersResponse is the response from listing RAG providers.
type SurrealRAGProvidersResponse struct {
	// Providers is the list of documentation providers with chunk counts.
	Providers []SurrealRAGProviderInfo `json:"providers"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// SurrealRAGProviders lists all documentation providers in the SurrealDB RAG system.
func (c *Client) SurrealRAGProviders(ctx context.Context) (*SurrealRAGProvidersResponse, error) {
	var resp SurrealRAGProvidersResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/rag/surreal/providers", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
