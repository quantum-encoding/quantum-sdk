package qai

import "context"

// DocumentRequest is the request body for document extraction.
// The file is sent as base64-encoded data in the JSON body.
type DocumentRequest struct {
	// FileBase64 is the base64-encoded file content.
	FileBase64 string `json:"file_base64"`

	// Filename is the original filename (helps determine the file type).
	Filename string `json:"filename"`

	// OutputFormat is the desired output format (e.g. "markdown", "text").
	OutputFormat string `json:"output_format,omitempty"`
}

// DocumentResponse is the response from document extraction.
type DocumentResponse struct {
	// Content is the extracted text content.
	Content string `json:"content"`

	// Format is the format of the extracted content (e.g. "markdown").
	Format string `json:"format"`

	// Meta contains provider-specific metadata about the document.
	Meta map[string]any `json:"meta,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// ExtractDocument extracts text content from a document (PDF, image, etc.).
func (c *Client) ExtractDocument(ctx context.Context, req *DocumentRequest) (*DocumentResponse, error) {
	var resp DocumentResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/documents/extract", req, &resp)
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
