package qai

import "context"

// ChunkDocumentRequest is the request body for document chunking.
// The file is sent as base64-encoded data in the JSON body.
type ChunkDocumentRequest struct {
	// FileBase64 is the base64-encoded file content.
	FileBase64 string `json:"file_base64"`

	// Filename is the original filename (helps determine the file type).
	Filename string `json:"filename"`

	// ChunkSize is the target chunk size in characters (optional).
	ChunkSize int `json:"chunk_size,omitempty"`

	// ChunkOverlap is the overlap between chunks in characters (optional).
	ChunkOverlap int `json:"chunk_overlap,omitempty"`
}

// ChunkDocumentResponse is the response from document chunking.
type ChunkDocumentResponse struct {
	// Chunks is the list of text chunks.
	Chunks []DocumentChunk `json:"chunks"`

	// Meta contains provider-specific metadata.
	Meta map[string]any `json:"meta,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// DocumentChunk is a single chunk of extracted document text.
type DocumentChunk struct {
	// Text is the chunk content.
	Text string `json:"text"`

	// Index is the chunk position within the document.
	Index int `json:"index"`

	// PageNumber is the source page number (if applicable).
	PageNumber int `json:"page_number,omitempty"`
}

// ProcessDocumentRequest is the request body for document processing.
// The file is sent as base64-encoded data in the JSON body.
type ProcessDocumentRequest struct {
	// FileBase64 is the base64-encoded file content.
	FileBase64 string `json:"file_base64"`

	// Filename is the original filename (helps determine the file type).
	Filename string `json:"filename"`

	// OutputFormat is the desired output format (e.g. "markdown", "text").
	OutputFormat string `json:"output_format,omitempty"`

	// ExtractImages controls whether images are embedded in the output.
	ExtractImages bool `json:"extract_images,omitempty"`
}

// ProcessDocumentResponse is the response from document processing.
type ProcessDocumentResponse struct {
	// Content is the processed document content.
	Content string `json:"content"`

	// Format is the format of the processed content.
	Format string `json:"format"`

	// Meta contains provider-specific metadata.
	Meta map[string]any `json:"meta,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// ChunkDocument extracts and chunks a document into smaller text segments.
func (c *Client) ChunkDocument(ctx context.Context, req *ChunkDocumentRequest) (*ChunkDocumentResponse, error) {
	var resp ChunkDocumentResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/documents/chunk", req, &resp)
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

// ProcessDocument extracts and processes a document with OCR and image embedding.
func (c *Client) ProcessDocument(ctx context.Context, req *ProcessDocumentRequest) (*ProcessDocumentResponse, error) {
	var resp ProcessDocumentResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/documents/process", req, &resp)
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
