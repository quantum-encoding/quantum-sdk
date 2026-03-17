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
}

// VideoResponse is the response from video generation.
type VideoResponse struct {
	// Videos contains the generated videos.
	Videos []GeneratedVideo `json:"videos"`

	// Model is the model that generated the videos.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
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
	return &resp, nil
}
