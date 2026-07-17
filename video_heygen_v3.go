package qai

// HeyGen v3: template detail/render + batch videos.

import (
	"context"
	"fmt"
	"net/url"
)

// ---------------------------------------------------------------------------
// HeyGen Template v3 (variable schema + render)
// ---------------------------------------------------------------------------

// VideoTemplateSceneVariable is a variable slot referenced by a template scene.
type VideoTemplateSceneVariable struct {
	// Name is the variable name (key into the template's variables map).
	Name string `json:"name"`

	// VariableType is the variable kind (e.g. "text", "image", "character", "voice").
	VariableType string `json:"variable_type"`
}

// VideoTemplateScene is a scene in a template, in template order.
type VideoTemplateScene struct {
	// SceneID is the scene identifier (usable in a generate request's SceneIDs).
	SceneID string `json:"scene_id"`

	// Script is the scene script with placeholders unreplaced
	// (e.g. "Introducing {{headline}}...").
	Script string `json:"script"`

	// Variables are the variables referenced by this scene.
	Variables []VideoTemplateSceneVariable `json:"variables"`
}

// VideoTemplateDetail is the detailed template info: variable schema + scenes.
//
// Each Variables[name] value is a discriminated union on its "type" field
// ("text" | "image" | "video" | "audio" | "voice" | "character"; unknown
// future types round-trip verbatim), returned in the exact shape a generate
// request accepts — replace defaults and submit back.
type VideoTemplateDetail struct {
	// ID is the template identifier.
	ID string `json:"id"`

	// Name is the template name.
	Name string `json:"name"`

	// AspectRatio is the aspect ratio (e.g. "16:9").
	AspectRatio string `json:"aspect_ratio"`

	// Variables is the variable schema keyed by variable name (union values
	// kept as raw JSON objects so unknown future variable types round-trip
	// verbatim).
	Variables map[string]any `json:"variables"`

	// Scenes are the scenes in template order.
	Scenes []VideoTemplateScene `json:"scenes"`
}

// VideoTemplateDetailResponse is the response from inspecting a template's
// variable schema (unbilled).
type VideoTemplateDetailResponse struct {
	// Template is the template detail.
	Template VideoTemplateDetail `json:"template"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// VideoTemplateDimension is the output dimension for a template render.
// Both values must be even, each 128-4096, and keep the template aspect ratio.
type VideoTemplateDimension struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// VideoSubtitlePosition is the subtitle position for burned-in captions.
type VideoSubtitlePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// VideoTemplateSubtitles holds subtitle options for a template render
// (implies captions).
type VideoTemplateSubtitles struct {
	// PresetName is the subtitle preset (e.g. "classic", "bold", "bright"). Required.
	PresetName string `json:"preset_name"`

	// Alignment (default 2).
	Alignment *int `json:"alignment,omitempty"`

	// DisableHighlight disables word highlighting.
	DisableHighlight *bool `json:"disable_highlight,omitempty"`

	// FontSize is the font size.
	FontSize *int `json:"font_size,omitempty"`

	// Position is the subtitle position.
	Position *VideoSubtitlePosition `json:"position,omitempty"`
}

// VideoTemplateGenerateRequest is the request body for rendering a video
// from a template (async job).
type VideoTemplateGenerateRequest struct {
	// Variables holds variable overrides keyed by name (at least one
	// required). Values use the same union shapes returned by the template
	// detail route; omitted variables keep the template defaults.
	Variables map[string]any `json:"variables"`

	// Title names the generated video.
	Title string `json:"title,omitempty"`

	// SceneIDs restricts the render to these scenes, in order (repeats allowed).
	SceneIDs []string `json:"scene_ids,omitempty"`

	// Dimension is the output dimension (must keep the template aspect ratio).
	Dimension *VideoTemplateDimension `json:"dimension,omitempty"`

	// FPS is the frames per second: 25 (default), 30, or 60.
	FPS int `json:"fps,omitempty"`

	// Caption burns captions (default false).
	Caption *bool `json:"caption,omitempty"`

	// Subtitles holds subtitle options (implies captions).
	Subtitles *VideoTemplateSubtitles `json:"subtitles,omitempty"`

	// ReorderMusic: background audio moves with scenes (default true).
	ReorderMusic *bool `json:"reorder_music,omitempty"`

	// KeepTextVerticallyCentered keeps text vertically centered (default false).
	KeepTextVerticallyCentered *bool `json:"keep_text_vertically_centered,omitempty"`

	// IncludeGIF includes a GIF preview in the webhook payload.
	IncludeGIF *bool `json:"include_gif,omitempty"`

	// EnableSharing enables a public share page.
	EnableSharing *bool `json:"enable_sharing,omitempty"`

	// FolderID is the HeyGen folder id.
	FolderID string `json:"folder_id,omitempty"`

	// BrandVoiceID is the brand voice id.
	BrandVoiceID string `json:"brand_voice_id,omitempty"`
}

// ---------------------------------------------------------------------------
// HeyGen Batch videos
// ---------------------------------------------------------------------------

// VideoBatchSubmitRequest is the request body for submitting a batch of videos.
type VideoBatchSubmitRequest struct {
	// Videos holds 1-100 raw HeyGen POST /v3/videos request bodies, passed
	// through verbatim. Each is polymorphic, discriminated by its "type"
	// field ("avatar" | "image" | "cinematic_avatar"), so items are kept as
	// opaque JSON objects.
	Videos []map[string]any `json:"videos"`

	// Title is the display name for the batch in the HeyGen app.
	Title string `json:"title,omitempty"`
}

// VideoBatchSubmitResponse is the response from submitting a video batch
// (202 Accepted).
type VideoBatchSubmitResponse struct {
	// BatchID is the batch id — poll VideoBatchStatus with it.
	BatchID string `json:"batch_id"`

	// Status is always "processing" at submit.
	Status string `json:"status"`

	// TotalItems is the count of submitted items.
	TotalItems int `json:"total_items"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// VideoBatchStatusQuery holds the query parameters for the batch status page.
type VideoBatchStatusQuery struct {
	// Limit is the page size (1-100; upstream default 100). Zero means unset.
	Limit int

	// Token is an opaque cursor from a previous response's next_token.
	Token string
}

// VideoBatchItemError is the per-item error detail in a batch status page.
type VideoBatchItemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// VideoBatchItem is one item of a batch status page, ordered by ItemIndex.
type VideoBatchItem struct {
	// ItemIndex is the zero-based position in the submitted videos array.
	ItemIndex int `json:"item_index"`

	// Status is "queued" | "processing" | "completed" | "failed".
	Status string `json:"status"`

	// VideoID is present once the item's video exists.
	VideoID string `json:"video_id,omitempty"`

	// VideoURL is present only when BillingStatus == "settled" and the
	// item completed.
	VideoURL string `json:"video_url,omitempty"`

	// Error is present only when the item failed.
	Error *VideoBatchItemError `json:"error,omitempty"`
}

// VideoBatchStatusResponse is the response from a batch status check
// (one cursor-paginated page of items).
//
// Billing settles the first time a GET observes a terminal batch status;
// video_url values are withheld until BillingStatus == "settled" — keep
// polling until then to obtain URLs.
type VideoBatchStatusResponse struct {
	// BatchID is the batch id.
	BatchID string `json:"batch_id"`

	// Title is the batch display name (may be empty).
	Title string `json:"title"`

	// Status is the batch-level status: "processing" | "completed" | "failed".
	Status string `json:"status"`

	// TotalItems is the count of submitted items.
	TotalItems int `json:"total_items"`

	// CountsByStatus holds per-item-status counts across the whole batch.
	CountsByStatus map[string]int `json:"counts_by_status"`

	// CreatedAt is the batch creation time in unix seconds (upstream HeyGen
	// timestamp — NOT RFC3339).
	CreatedAt int64 `json:"created_at"`

	// Items is one page of items, ordered by ItemIndex.
	Items []VideoBatchItem `json:"items"`

	// HasMore indicates more item pages exist.
	HasMore bool `json:"has_more"`

	// NextToken should be passed as Token for the next page (may be empty).
	NextToken string `json:"next_token"`

	// BillingStatus is "unsettled" | "settlement_pending" | "settled".
	BillingStatus string `json:"billing_status"`

	// CostTicks is the total ticks charged for the batch; 0 until settled.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// ---------------------------------------------------------------------------
// Client methods
// ---------------------------------------------------------------------------

// VideoTemplateDetail inspects a HeyGen template's variable schema and
// scenes (unbilled).
//
// Only draft-v4 templates with variables are supported upstream; an unknown
// template id surfaces as a provider_error.
func (c *Client) VideoTemplateDetail(ctx context.Context, templateID string) (*VideoTemplateDetailResponse, error) {
	var resp VideoTemplateDetailResponse
	_, err := c.doJSON(ctx, "GET", fmt.Sprintf("/qai/v1/video/template/%s", templateID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoTemplateGenerate renders a video from a HeyGen template (async job
// type "video/template-v3").
//
// Returns the accepted-job envelope — poll with GetJob / PollJob (or SSE via
// StreamJob) until "completed"/"failed", then read result.video_url. Deep
// validation happens at execution time, so violations surface as a failed
// job rather than a 4xx at submit.
func (c *Client) VideoTemplateGenerate(ctx context.Context, templateID string, req *VideoTemplateGenerateRequest) (*JobAcceptedResponse, error) {
	var resp JobAcceptedResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/video/template/%s", templateID), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoBatchSubmit submits 1-100 raw HeyGen video payloads as one batch
// (202 Accepted).
//
// Poll VideoBatchStatus for progress and delivery.
func (c *Client) VideoBatchSubmit(ctx context.Context, req *VideoBatchSubmitRequest) (*VideoBatchSubmitResponse, error) {
	var resp VideoBatchSubmitResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/video/batch", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoBatchStatus gets a batch's status plus one cursor-paginated page
// of items.
//
// Poll (~5s) until Status is terminal, then keep polling until
// BillingStatus == "settled" — per-item video_url values are withheld until
// settlement. Collect URLs across pages via NextToken.
func (c *Client) VideoBatchStatus(ctx context.Context, batchID string, query *VideoBatchStatusQuery) (*VideoBatchStatusResponse, error) {
	path := fmt.Sprintf("/qai/v1/video/batch/%s", batchID)
	if query != nil {
		params := url.Values{}
		if query.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", query.Limit))
		}
		if query.Token != "" {
			params.Set("token", query.Token)
		}
		if encoded := params.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	var resp VideoBatchStatusResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
