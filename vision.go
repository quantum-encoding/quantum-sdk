package qai

import "context"

// ---------------------------------------------------------------------------
// Request
// ---------------------------------------------------------------------------

// VisionRequest is the request body for vision analysis endpoints.
type VisionRequest struct {
	// Base64-encoded image (with or without data: prefix).
	ImageBase64 string `json:"image_base64,omitempty"`

	// Image URL (fetched by the model provider).
	ImageURL string `json:"image_url,omitempty"`

	// Model to use. Default: gemini-2.5-flash.
	Model string `json:"model,omitempty"`

	// Analysis profile: "combined" (default), "scene", "objects", "ocr", "quality".
	Profile string `json:"profile,omitempty"`

	// Domain context for relevance checking.
	Context *VisionContext `json:"context,omitempty"`
}

// VisionContext provides domain context for relevance analysis.
type VisionContext struct {
	// Installation type (e.g. "solar", "heat_pump", "ev_charger").
	InstallationType string `json:"installation_type,omitempty"`

	// Phase (e.g. "pre_install", "installation", "post_install").
	Phase string `json:"phase,omitempty"`

	// Expected items for relevance checking.
	ExpectedItems []string `json:"expected_items,omitempty"`
}

// ---------------------------------------------------------------------------
// Response
// ---------------------------------------------------------------------------

// VisionResponse is the full vision analysis response.
type VisionResponse struct {
	// Scene description.
	Caption string `json:"caption,omitempty"`

	// Suggested tags (lowercase_snake_case).
	Tags []string `json:"tags"`

	// Detected objects with bounding boxes.
	Objects []DetectedObject `json:"objects"`

	// Image quality assessment.
	Quality *QualityAssessment `json:"quality,omitempty"`

	// Relevance check against context.
	Relevance *RelevanceCheck `json:"relevance,omitempty"`

	// Extracted text and overlay metadata.
	OCR *OCRResult `json:"ocr,omitempty"`

	// Model used.
	Model string `json:"model"`

	// Cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// Request identifier.
	RequestID string `json:"request_id"`
}

// DetectedObject is an object detected in the image with a bounding box.
type DetectedObject struct {
	// Object label.
	Label string `json:"label"`

	// Detection confidence (0.0 - 1.0).
	Confidence float64 `json:"confidence"`

	// Bounding box: [y_min, x_min, y_max, x_max] normalised to 0-1000.
	BoundingBox [4]int32 `json:"bounding_box"`
}

// QualityAssessment is an image quality assessment.
type QualityAssessment struct {
	// Overall rating: "good", "acceptable", "poor".
	Overall string `json:"overall"`

	// Quality score (0.0 - 1.0).
	Score float64 `json:"score"`

	// Blur level: "none", "slight", "significant".
	Blur string `json:"blur"`

	// Lighting: "well_lit", "dim", "dark".
	Darkness string `json:"darkness"`

	// Resolution: "high", "adequate", "low".
	Resolution string `json:"resolution"`

	// Exposure: "correct", "over", "under".
	Exposure string `json:"exposure"`

	// Specific issues found.
	Issues []string `json:"issues"`
}

// RelevanceCheck is a relevance check against expected content.
type RelevanceCheck struct {
	// Whether the image is relevant to the context.
	Relevant bool `json:"relevant"`

	// Relevance score (0.0 - 1.0).
	Score float64 `json:"score"`

	// Items expected based on context.
	ExpectedItems []string `json:"expected_items"`

	// Items actually found in the image.
	FoundItems []string `json:"found_items"`

	// Expected but not found.
	MissingItems []string `json:"missing_items"`

	// Found but not expected.
	UnexpectedItems []string `json:"unexpected_items"`

	// Additional notes.
	Notes string `json:"notes,omitempty"`
}

// OCRResult is the OCR / text extraction result.
type OCRResult struct {
	// All extracted text concatenated.
	Text string `json:"text,omitempty"`

	// Extracted metadata (GPS, timestamp, address, etc.).
	Metadata map[string]string `json:"metadata"`

	// Individual text overlays with positions.
	Overlays []TextOverlay `json:"overlays"`
}

// TextOverlay is a detected text region in the image.
type TextOverlay struct {
	// Extracted text content.
	Text string `json:"text"`

	// Bounding box: [y_min, x_min, y_max, x_max] normalised to 0-1000.
	BoundingBox *[4]int32 `json:"bounding_box,omitempty"`

	// Overlay type: "gps", "timestamp", "address", "label", "other".
	Type string `json:"type,omitempty"`
}

// ---------------------------------------------------------------------------
// Client methods
// ---------------------------------------------------------------------------

// VisionAnalyze performs a full combined vision analysis
// (scene + objects + quality + OCR + relevance).
func (c *Client) VisionAnalyze(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	var resp VisionResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/vision/analyze", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VisionDetect performs object detection with bounding boxes.
func (c *Client) VisionDetect(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	var resp VisionResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/vision/detect", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VisionDescribe returns a scene description and tags.
func (c *Client) VisionDescribe(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	var resp VisionResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/vision/describe", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VisionOCR extracts text and overlay metadata from an image.
func (c *Client) VisionOCR(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	var resp VisionResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/vision/ocr", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VisionQuality performs an image quality assessment.
func (c *Client) VisionQuality(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	var resp VisionResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/vision/quality", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
