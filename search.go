package qai

import "context"

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

// SearchOptions configures web search requests.
type SearchOptions struct {
	// Count is the number of results to return (default server-side).
	Count int `json:"count,omitempty"`

	// Offset is the zero-based result offset for pagination.
	Offset int `json:"offset,omitempty"`

	// Country filters results by country code (e.g. "US", "GB").
	Country string `json:"country,omitempty"`

	// Language filters results by language code (e.g. "en", "fr").
	Language string `json:"language,omitempty"`

	// Freshness limits results to a time range (e.g. "24h", "7d", "30d").
	Freshness string `json:"freshness,omitempty"`

	// SafeSearch controls adult content filtering ("off", "moderate", "strict").
	SafeSearch string `json:"safe_search,omitempty"`
}

// ContextOptions configures LLM context search requests.
type ContextOptions struct {
	// Count is the number of context chunks to return.
	Count int `json:"count,omitempty"`

	// Country filters results by country code.
	Country string `json:"country,omitempty"`

	// Language filters results by language code.
	Language string `json:"language,omitempty"`

	// Freshness limits results to a time range.
	Freshness string `json:"freshness,omitempty"`
}

// SearchMessage is a message in a search-answer conversation.
type SearchMessage struct {
	// Role is "user" or "assistant".
	Role string `json:"role"`

	// Content is the message text.
	Content string `json:"content"`
}

// ---------------------------------------------------------------------------
// Internal request bodies
// ---------------------------------------------------------------------------

type webSearchRequest struct {
	Query      string `json:"query"`
	Count      int    `json:"count,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Country    string `json:"country,omitempty"`
	Language   string `json:"language,omitempty"`
	Freshness  string `json:"freshness,omitempty"`
	SafeSearch string `json:"safe_search,omitempty"`
}

type contextSearchRequest struct {
	Query     string `json:"query"`
	Count     int    `json:"count,omitempty"`
	Country   string `json:"country,omitempty"`
	Language  string `json:"language,omitempty"`
	Freshness string `json:"freshness,omitempty"`
}

type searchAnswerRequest struct {
	Messages []SearchMessage `json:"messages"`
	Model    string          `json:"model"`
}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

// WebSearchResponse is the response from a web search request.
type WebSearchResponse struct {
	// Query is the original search query.
	Query string `json:"query"`

	// Web contains organic web results.
	Web []WebResult `json:"web,omitempty"`

	// News contains news results.
	News []NewsResult `json:"news,omitempty"`

	// Videos contains video results.
	Videos []VideoResult `json:"videos,omitempty"`

	// Infobox contains an infobox if one was found.
	Infobox *Infobox `json:"infobox,omitempty"`

	// Discussions contains forum/discussion results.
	Discussions []Discussion `json:"discussions,omitempty"`

	// CostTicks is the total cost from the X-QAI-Cost-Ticks header.
	CostTicks int64 `json:"-"`

	// RequestID is from the X-QAI-Request-Id header.
	RequestID string `json:"-"`
}

// WebResult is a single organic web search result.
type WebResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Age         string `json:"age,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
}

// NewsResult is a single news search result.
type NewsResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Age         string `json:"age,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	Source      string `json:"source,omitempty"`
}

// VideoResult is a single video search result.
type VideoResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Age         string `json:"age,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
}

// Infobox contains structured information about a search topic.
type Infobox struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	URL         string            `json:"url,omitempty"`
	Thumbnail   string            `json:"thumbnail,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// Discussion is a forum or discussion result.
type Discussion struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Age         string `json:"age,omitempty"`
}

// LLMContextResponse is the response from a context search request.
type LLMContextResponse struct {
	// Query is the original search query.
	Query string `json:"query"`

	// Chunks contains the context chunks suitable for LLM consumption.
	Chunks []ContextChunk `json:"chunks,omitempty"`

	// Sources lists the source URLs used.
	Sources []string `json:"sources,omitempty"`

	// CostTicks is the total cost from the X-QAI-Cost-Ticks header.
	CostTicks int64 `json:"-"`

	// RequestID is from the X-QAI-Request-Id header.
	RequestID string `json:"-"`
}

// ContextChunk is a single chunk of context from a web page.
type ContextChunk struct {
	Content     string  `json:"content"`
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Score       float64 `json:"score,omitempty"`
	ContentType string  `json:"content_type,omitempty"`
}

// SearchAnswerResponse is the response from a search-answer request.
type SearchAnswerResponse struct {
	// ID is the unique request identifier.
	ID string `json:"id"`

	// Model is the model that generated the answer.
	Model string `json:"model"`

	// Choices contains the generated answers.
	Choices []SearchAnswerChoice `json:"choices"`

	// Citations are the sources referenced in the answer.
	Citations []Citation `json:"citations,omitempty"`

	// CostTicks is the total cost from the X-QAI-Cost-Ticks header.
	CostTicks int64 `json:"-"`

	// RequestID is from the X-QAI-Request-Id header.
	RequestID string `json:"-"`
}

// SearchAnswerChoice is a single answer choice.
type SearchAnswerChoice struct {
	Index        int           `json:"index"`
	Message      SearchMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// Citation is a source referenced in a search answer.
type Citation struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet,omitempty"`
}

// ---------------------------------------------------------------------------
// Client methods
// ---------------------------------------------------------------------------

// WebSearch performs a web search using the Brave Search API.
//
//	results, err := client.WebSearch(ctx, "Go 1.24 release notes", &qai.SearchOptions{
//	    Count: 10,
//	    Country: "US",
//	})
func (c *Client) WebSearch(ctx context.Context, query string, opts *SearchOptions) (*WebSearchResponse, error) {
	reqBody := webSearchRequest{Query: query}
	if opts != nil {
		reqBody.Count = opts.Count
		reqBody.Offset = opts.Offset
		reqBody.Country = opts.Country
		reqBody.Language = opts.Language
		reqBody.Freshness = opts.Freshness
		reqBody.SafeSearch = opts.SafeSearch
	}

	var resp WebSearchResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/search/web", reqBody, &resp)
	if err != nil {
		return nil, err
	}

	resp.CostTicks = meta.CostTicks
	resp.RequestID = meta.RequestID

	return &resp, nil
}

// SearchContext retrieves LLM-optimised context chunks for a query using web search.
//
//	ctx, err := client.SearchContext(ctx, "quantum computing basics", &qai.ContextOptions{
//	    Count: 5,
//	})
func (c *Client) SearchContext(ctx context.Context, query string, opts *ContextOptions) (*LLMContextResponse, error) {
	reqBody := contextSearchRequest{Query: query}
	if opts != nil {
		reqBody.Count = opts.Count
		reqBody.Country = opts.Country
		reqBody.Language = opts.Language
		reqBody.Freshness = opts.Freshness
	}

	var resp LLMContextResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/search/context", reqBody, &resp)
	if err != nil {
		return nil, err
	}

	resp.CostTicks = meta.CostTicks
	resp.RequestID = meta.RequestID

	return &resp, nil
}

// SearchAnswer generates an AI answer grounded in web search results.
//
//	answer, err := client.SearchAnswer(ctx, []qai.SearchMessage{
//	    {Role: "user", Content: "What is the latest Go version?"},
//	}, "grok-3-mini")
func (c *Client) SearchAnswer(ctx context.Context, messages []SearchMessage, model string) (*SearchAnswerResponse, error) {
	reqBody := searchAnswerRequest{
		Messages: messages,
		Model:    model,
	}

	var resp SearchAnswerResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/search/answer", reqBody, &resp)
	if err != nil {
		return nil, err
	}

	resp.CostTicks = meta.CostTicks
	resp.RequestID = meta.RequestID
	if resp.Model == "" {
		resp.Model = meta.Model
	}

	return &resp, nil
}
