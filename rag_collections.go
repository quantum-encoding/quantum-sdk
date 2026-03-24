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
	// FileID is the document file ID.
	FileID string `json:"file_id"`

	// Name is the document filename.
	Name string `json:"name"`

	// SizeBytes is the document size in bytes.
	SizeBytes int64 `json:"size_bytes,omitempty"`

	// ContentType is the MIME content type.
	ContentType string `json:"content_type,omitempty"`

	// ProcessingStatus is the current processing status.
	ProcessingStatus string `json:"processing_status,omitempty"`

	// DocumentStatus is the document lifecycle status.
	DocumentStatus string `json:"document_status,omitempty"`

	// Indexed indicates whether the document has been indexed.
	Indexed *bool `json:"indexed,omitempty"`

	// CreatedAt is the ISO timestamp of creation.
	CreatedAt string `json:"created_at,omitempty"`
}

// CollectionSearchResult is a single result from collection search.
type CollectionSearchResult struct {
	// Content is the matching text content.
	Content string `json:"content"`

	// Score is the relevance score.
	Score float64 `json:"score,omitempty"`

	// FileID is the source file ID.
	FileID string `json:"file_id,omitempty"`

	// CollectionID is the source collection ID.
	CollectionID string `json:"collection_id,omitempty"`

	// Metadata contains additional result metadata.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// CollectionSearchRequest is the request body for collection search.
type CollectionSearchRequest struct {
	// Query is the search query.
	Query string `json:"query"`

	// CollectionIDs is the list of collection IDs to search.
	CollectionIDs []string `json:"collection_ids"`

	// Mode is the search mode (e.g. "hybrid", "semantic", "keyword").
	Mode string `json:"mode,omitempty"`

	// MaxResults is the maximum number of results to return.
	MaxResults int `json:"max_results,omitempty"`
}

// CollectionUploadResult is the result of uploading a document to a collection.
type CollectionUploadResult struct {
	// FileID is the uploaded file ID.
	FileID string `json:"file_id"`

	// Filename is the original filename.
	Filename string `json:"filename"`

	// Bytes is the uploaded file size in bytes.
	Bytes int64 `json:"bytes,omitempty"`
}

// collectionsListResponse is the API response wrapper for collection listing.
type collectionsListResponse struct {
	Collections []Collection `json:"collections"`
}

// collectionDocumentsResponse is the API response wrapper for document listing.
type collectionDocumentsResponse struct {
	Documents []CollectionDocument `json:"documents"`
}

// collectionSearchResponse is the API response wrapper for collection search.
type collectionSearchResponse struct {
	Results []CollectionSearchResult `json:"results"`
}

// deleteCollectionResponse is the API response wrapper for collection deletion.
type deleteCollectionResponse struct {
	Message string `json:"message"`
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

// CollectionsGet gets details for a single collection (must be owned or shared).
func (c *Client) CollectionsGet(ctx context.Context, id string) (*Collection, error) {
	var resp Collection
	_, err := c.doJSON(ctx, "GET", "/qai/v1/rag/collections/"+id, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CollectionsDelete deletes a collection (owner only).
func (c *Client) CollectionsDelete(ctx context.Context, id string) error {
	var resp deleteCollectionResponse
	_, err := c.doJSON(ctx, "DELETE", "/qai/v1/rag/collections/"+id, nil, &resp)
	return err
}

// CollectionsDocuments lists documents in a collection.
func (c *Client) CollectionsDocuments(ctx context.Context, collectionID string) ([]CollectionDocument, error) {
	var resp collectionDocumentsResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/rag/collections/"+collectionID+"/documents", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Documents, nil
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

// CollectionsSearch searches across collections (user's + shared) with
// hybrid/semantic/keyword mode.
func (c *Client) CollectionsSearch(ctx context.Context, req *CollectionSearchRequest) ([]CollectionSearchResult, error) {
	var resp collectionSearchResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/rag/search/collections", req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}
