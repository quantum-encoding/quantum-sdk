package qai

import "context"

// RAGSearchRequest is the request body for Vertex AI RAG search.
type RAGSearchRequest struct {
	// Query is the search query.
	Query string `json:"query"`

	// Corpus filters by corpus name or ID (fuzzy match). Omit to search all corpora.
	Corpus string `json:"corpus,omitempty"`

	// TopK is the maximum number of results to return (default 10).
	TopK int `json:"top_k,omitempty"`
}

// RAGSearchResponse is the response from RAG search.
type RAGSearchResponse struct {
	// Results contains the matching document chunks.
	Results []RAGResult `json:"results"`

	// Query is the original search query.
	Query string `json:"query"`

	// Corpora lists the corpora that were searched.
	Corpora []string `json:"corpora,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// RAGResult is a single result from RAG search.
type RAGResult struct {
	// SourceURI is the source document URI.
	SourceURI string `json:"source_uri"`

	// SourceName is the display name of the source.
	SourceName string `json:"source_name"`

	// Text is the matching text chunk.
	Text string `json:"text"`

	// Score is the relevance score.
	Score float64 `json:"score"`

	// Distance is the vector distance (lower is more similar).
	Distance float64 `json:"distance"`
}

// RAGCorpus describes an available RAG corpus.
type RAGCorpus struct {
	// Name is the full resource name.
	Name string `json:"name"`

	// DisplayName is the human-readable name.
	DisplayName string `json:"displayName"`

	// Description describes the corpus contents.
	Description string `json:"description"`

	// State is the corpus state (e.g. "ACTIVE").
	State string `json:"state"`
}

// ragCorporaResponse is the API response wrapper for corpus listing.
type ragCorporaResponse struct {
	Corpora   []RAGCorpus `json:"corpora"`
	RequestID string      `json:"request_id"`
}

// SurrealRAGSearchRequest is the request body for SurrealDB-backed RAG search.
type SurrealRAGSearchRequest struct {
	// Query is the search query.
	Query string `json:"query"`

	// Provider filters by documentation provider (e.g. "xai", "claude", "heygen").
	Provider string `json:"provider,omitempty"`

	// Limit is the maximum number of results (default 10, max 50).
	Limit int `json:"limit,omitempty"`
}

// SurrealRAGSearchResponse is the response from SurrealDB RAG search.
type SurrealRAGSearchResponse struct {
	// Results contains the matching documentation chunks.
	Results []SurrealRAGResult `json:"results"`

	// Query is the original search query.
	Query string `json:"query"`

	// Provider is the provider filter that was applied.
	Provider string `json:"provider,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// SurrealRAGResult is a single result from SurrealDB RAG search.
type SurrealRAGResult struct {
	// Provider is the documentation provider.
	Provider string `json:"provider"`

	// Title is the document title.
	Title string `json:"title"`

	// Heading is the section heading.
	Heading string `json:"heading"`

	// SourceFile is the original source file path.
	SourceFile string `json:"source_file"`

	// Content is the matching text chunk.
	Content string `json:"content"`

	// Score is the cosine similarity score.
	Score float64 `json:"score"`
}

// RAGSearch searches Vertex AI RAG corpora for relevant documentation.
func (c *Client) RAGSearch(ctx context.Context, req *RAGSearchRequest) (*RAGSearchResponse, error) {
	var resp RAGSearchResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/rag/search", req, &resp)
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

// RAGCorpora lists available Vertex AI RAG corpora.
func (c *Client) RAGCorpora(ctx context.Context) ([]RAGCorpus, error) {
	var resp ragCorporaResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/rag/corpora", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Corpora, nil
}

// SurrealRAGSearch searches provider API documentation via SurrealDB vector search.
func (c *Client) SurrealRAGSearch(ctx context.Context, req *SurrealRAGSearchRequest) (*SurrealRAGSearchResponse, error) {
	var resp SurrealRAGSearchResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/rag/surreal/search", req, &resp)
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
