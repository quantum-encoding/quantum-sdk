package qai

import "context"

// ImageRequest is the request body for image generation.
type ImageRequest struct {
	// Model is the image generation model (e.g. "grok-imagine-image", "gpt-image-1", "dall-e-3").
	Model string `json:"model"`

	// Prompt describes the image to generate.
	Prompt string `json:"prompt"`

	// Count is the number of images to generate (default 1).
	Count int `json:"count,omitempty"`

	// Size specifies the output dimensions (e.g. "1024x1024", "1536x1024").
	Size string `json:"size,omitempty"`

	// AspectRatio specifies the aspect ratio (e.g. "16:9", "1:1").
	AspectRatio string `json:"aspect_ratio,omitempty"`

	// Quality is the quality level (e.g. "standard", "hd").
	Quality string `json:"quality,omitempty"`

	// OutputFormat is the image format (e.g. "png", "jpeg", "webp").
	OutputFormat string `json:"output_format,omitempty"`

	// Style is the style preset (e.g. "vivid", "natural"). DALL-E 3 specific.
	Style string `json:"style,omitempty"`

	// Background is the background mode (e.g. "auto", "transparent", "opaque"). GPT-Image specific.
	Background string `json:"background,omitempty"`

	// ImageURL is the image URL or data URI for image-to-3D conversion (Meshy).
	ImageURL string `json:"image_url,omitempty"`

	// Topology is the mesh topology: "triangle" or "quad".
	Topology string `json:"topology,omitempty"`

	// TargetPolycount is the target polygon count (100-300,000).
	TargetPolycount int `json:"target_polycount,omitempty"`

	// SymmetryMode is the symmetry mode: "auto", "on", or "off".
	SymmetryMode string `json:"symmetry_mode,omitempty"`

	// PoseMode is the pose mode: "", "a-pose", or "t-pose".
	PoseMode string `json:"pose_mode,omitempty"`

	// EnablePBR generates PBR texture maps (base_color, metallic, roughness, normal).
	EnablePBR *bool `json:"enable_pbr,omitempty"`
}

// ImageResponse is the response from image generation.
type ImageResponse struct {
	// Images contains the generated images.
	Images []GeneratedImage `json:"images"`

	// Model is the model that generated the images.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// GeneratedImage is a single generated image.
type GeneratedImage struct {
	// Base64Data is the base64-encoded image data (or a URL for 3D models).
	Base64Data string `json:"base64"`

	// Format is the image format (e.g. "png", "jpeg").
	Format string `json:"format"`

	// Index is the image index within the batch.
	Index int `json:"index"`
}

// ImageEditRequest is the request body for image editing.
type ImageEditRequest struct {
	// Model is the editing model (e.g. "gpt-image-1", "grok-imagine-image").
	Model string `json:"model"`

	// Prompt describes the desired edit.
	Prompt string `json:"prompt"`

	// InputImages is a list of base64-encoded input images.
	InputImages []string `json:"input_images"`

	// Count is the number of edited images to generate (default 1).
	Count int `json:"count,omitempty"`

	// Size specifies the output dimensions.
	Size string `json:"size,omitempty"`
}

// ImageEditResponse is the response from image editing (same shape as generation).
type ImageEditResponse = ImageResponse

// GenerateImage generates images from a text prompt.
func (c *Client) GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResponse, error) {
	var resp ImageResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/images/generate", req, &resp)
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

// EditImage edits images using an AI model.
func (c *Client) EditImage(ctx context.Context, req *ImageEditRequest) (*ImageEditResponse, error) {
	var resp ImageEditResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/images/edit", req, &resp)
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
