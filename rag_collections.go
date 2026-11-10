package qai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

// Collection is a user-scoped xAI collection (proxied through quantum-ai).
type Collection struct {
	// ID is the collection ID (xAI-issued).
	ID string `json:"id"`

	// Name is the human-readable collection name.
	Name string `json:"name"`

	// Description is an optional collection description.
	Description string `json:"description,omitempty"`

	// DocumentCount is the number of documents in the collection.
	DocumentCount int64 `json:"document_count,omitempty"`

	// Owner is the user ID or "shared".
	Owner string `json:"owner,omitempty"`

	// Provider is the backend provider (e.g. "xai").
	Provider string `json:"provider,omitempty"`

	// CreatedAt is the ISO timestamp of creation.
	CreatedAt string `json:"created_at,omitempty"`
}

// CollectionDocument is a document within a collection.
type CollectionDocument struct {
	// ID is the gateway-issued document identifier.
	ID string `json:"id"`

	// CollectionID is the collection the document belongs to.
	CollectionID string `json:"collection_id"`

	// FileID is the provider's own file ID.
	FileID string `json:"file_id"`

	// Filename is the uploaded filename.
	Filename string `json:"filename"`

	// Status is the indexing status. The upload route records "indexed" as
	// soon as the provider accepts the file.
	Status string `json:"status"`

	// Chunks is the number of chunks the document was split into. The upload
	// route never sets it, so it reads 0.
	Chunks int `json:"chunks"`

	// UploadedAt is the RFC 3339 upload timestamp.
	UploadedAt string `json:"uploaded_at"`
}

// CollectionDetail is one collection with its documents, as
// GET /qai/v1/rag/collections/{id} returns it.
type CollectionDetail struct {
	// Collection is the collection record.
	Collection Collection `json:"collection"`

	// Documents are the documents uploaded to it.
	Documents []CollectionDocument `json:"documents"`
}

// CollectionSearchResult is one chunk matched by a collection search.
type CollectionSearchResult struct {
	// Content is the matched chunk text.
	Content string `json:"content"`

	// Score is the relevance score; results arrive highest first.
	Score float64 `json:"score"`

	// Collection is the name of the collection the chunk came from.
	Collection string `json:"collection"`

	// CollectionID is the ID of the collection the chunk came from.
	CollectionID string `json:"collection_id"`

	// DocumentID is the provider's document ID, when it reported one.
	DocumentID string `json:"document_id,omitempty"`

	// Filename is the source filename, when the provider reported one.
	Filename string `json:"filename,omitempty"`

	// IsShared is true when the chunk came from a shared collection rather
	// than one the caller owns.
	IsShared bool `json:"is_shared"`
}

// CollectionSearchRequest is the request body for collection search.
type CollectionSearchRequest struct {
	// Query is the search query (required).
	Query string `json:"query"`

	// CollectionIDs limits the search to these collections. Empty searches
	// every collection the caller can read, their own and the shared ones.
	// IDs the caller cannot read are skipped without an error.
	CollectionIDs []string `json:"collection_ids,omitempty"`

	// MaxChunks caps the results across all collections. 0 lets the gateway
	// apply its default of 10.
	MaxChunks int `json:"max_chunks,omitempty"`
}

// CollectionUploadResult is the document record an upload creates.
type CollectionUploadResult = CollectionDocument

// collectionsListResponse is the API response wrapper for collection listing.
type collectionsListResponse struct {
	Collections []Collection `json:"collections"`
}

// collectionSearchResponse is the API response wrapper for collection search.
type collectionSearchResponse struct {
	Results []CollectionSearchResult `json:"results"`
}

// createCollectionRequest is the request body for creating a collection.
type createCollectionRequest struct {
	Name string `json:"name"`
}

// CollectionsList lists the user's collections plus shared collections.
func (c *Client) CollectionsList(ctx context.Context) ([]Collection, error) {
	var resp collectionsListResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/rag/collections", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Collections, nil
}

// CollectionsCreate creates a new user-owned collection.
func (c *Client) CollectionsCreate(ctx context.Context, name string) (*Collection, error) {
	req := createCollectionRequest{Name: name}
	var resp Collection
	_, err := c.doJSON(ctx, "POST", "/qai/v1/rag/collections", &req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CollectionsGet gets a single collection and its documents. The caller must
// own the collection or it must be shared.
func (c *Client) CollectionsGet(ctx context.Context, id string) (*CollectionDetail, error) {
	var resp CollectionDetail
	_, err := c.doJSON(ctx, "GET", "/qai/v1/rag/collections/"+id, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CollectionsDelete deletes a collection (owner only).
func (c *Client) CollectionsDelete(ctx context.Context, id string) error {
	_, err := c.doJSON(ctx, "DELETE", "/qai/v1/rag/collections/"+id, nil, nil)
	return err
}

// CollectionsUpload uploads a file to a collection. The server handles the
// two-step xAI upload (files API + management API) with the master key.
func (c *Client) CollectionsUpload(ctx context.Context, collectionID, filename string, content []byte) (*CollectionUploadResult, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("qai: create multipart field: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return nil, fmt.Errorf("qai: write multipart content: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("qai: close multipart writer: %w", err)
	}

	path := "/qai/v1/rag/collections/" + collectionID + "/upload"
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("qai: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qai: POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	meta := &responseMeta{
		RequestID: resp.Header.Get("X-QAI-Request-Id"),
	}
	if v := resp.Header.Get("X-QAI-Cost-Ticks"); v != "" {
		meta.CostTicks, _ = strconv.ParseInt(v, 10, 64)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp, meta.RequestID)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("qai: read response: %w", err)
	}

	var result CollectionUploadResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("qai: decode response: %w", err)
	}

	return &result, nil
}

// CollectionsSearch searches the requested collections, or every collection
// the caller can read when req.CollectionIDs is empty, and returns the matched
// chunks merged across collections, highest score first, capped at
// req.MaxChunks.
func (c *Client) CollectionsSearch(ctx context.Context, req *CollectionSearchRequest) ([]CollectionSearchResult, error) {
	var resp collectionSearchResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/rag/collections/search", req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}
