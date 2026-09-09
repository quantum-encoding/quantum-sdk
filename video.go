package qai

import "context"

// VideoRequest is the request body for video generation.
type VideoRequest struct {
	// Model is the video generation model (e.g. "heygen", "grok-imagine-video", "sora-2", "veo-2").
	Model string `json:"model"`

	// Prompt describes the video to generate.
	Prompt string `json:"prompt"`

	// DurationSeconds is the target video duration in seconds (default 8).
	DurationSeconds int `json:"duration_seconds,omitempty"`

	// AspectRatio specifies the video aspect ratio (e.g. "16:9", "9:16").
	AspectRatio string `json:"aspect_ratio,omitempty"`

	// IdempotencyKey is sent as the Idempotency-Key header; auto-generated if empty.
	IdempotencyKey string `json:"-"`
}

// idempotencyKey returns the caller-set key, auto-generating one if empty.
func (r *VideoRequest) idempotencyKey() string {
	if r.IdempotencyKey == "" {
		r.IdempotencyKey = newIdempotencyKey()
	}
	return r.IdempotencyKey
}

// VideoResponse is the response from video generation.
type VideoResponse struct {
	// Videos contains the generated videos.
	Videos []GeneratedVideo `json:"videos"`

	// Model is the model that generated the videos.
	Model string `json:"model"`

	// DurationSeconds is the length of the video actually produced, when the
	// provider reports it. This is the quantity a per-second model is billed
	// on — settlement prefers it over the requested duration — so it is the
	// basis of CostTicks. Zero when the provider reports no length.
	DurationSeconds float64 `json:"duration_seconds,omitempty"`

	// Usage is the token counts behind a TOKEN-billed video charge. Gemini
	// Omni is the only such model: it meters output by modality at ~5,792
	// tokens per second of 720p, so on that path tokens are the whole cost
	// basis. Nil for per-second and per-clip models, whose cost is a function
	// of duration instead — a zeroed struct would assert a token basis the
	// charge does not have.
	Usage *MediaTokenUsage `json:"usage,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// BalanceAfter is the wallet balance after this call, from the
	// X-QAI-Balance-After header. Signed — a claw-back can make it negative.
	BalanceAfter int64 `json:"-"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// MediaTokenUsage is the token breakdown behind a token-billed media charge.
//
// The gateway omits every bucket it has nothing to report for, and sends no
// object at all when all of them would be zero, so a nil Usage means the
// charge has no token basis rather than a token basis of nothing.
type MediaTokenUsage struct {
	// PromptTokens is the input billed at the prompt rate.
	PromptTokens int `json:"prompt_tokens,omitempty"`

	// CompletionTokens is the output. For Gemini Omni this is the
	// modality-metered video output, which is most of the charge.
	CompletionTokens int `json:"completion_tokens,omitempty"`

	// ReasoningTokens is billed at the output rate.
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`

	// CachedTokens is the input served from cache, billed at the cache-read
	// rate. Omitted by providers that report no cache hits.
	CachedTokens int `json:"cached_tokens,omitempty"`

	// TotalTokens is the provider-reported total across the buckets.
	TotalTokens int `json:"total_tokens,omitempty"`
}

// GeneratedVideo is a single generated video.
type GeneratedVideo struct {
	// Base64Data is the base64-encoded video data (or a URL).
	Base64Data string `json:"base64"`

	// Format is the video format (e.g. "mp4").
	Format string `json:"format"`

	// SizeBytes is the video file size.
	SizeBytes int `json:"size_bytes"`

	// Index is the video index within the batch.
	Index int `json:"index"`
}

// GenerateVideo generates a video from a text prompt.
//
// Video generation is slow (30s-5min). For production use, consider submitting
// via the Jobs API instead.
func (c *Client) GenerateVideo(ctx context.Context, req *VideoRequest) (*VideoResponse, error) {
	var resp VideoResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/video/generate", req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.CostTicks == 0 {
		resp.CostTicks = meta.CostTicks
	}
	if resp.RequestID == "" {
		resp.RequestID = meta.RequestID
	}
	resp.BalanceAfter = meta.BalanceAfter
	return &resp, nil
}
