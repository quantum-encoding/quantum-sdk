package qai

import "context"

// StudioClip is a clip in a studio video.
type StudioClip struct {
	// AvatarID is the avatar ID.
	AvatarID string `json:"avatar_id,omitempty"`

	// VoiceID is the voice ID.
	VoiceID string `json:"voice_id,omitempty"`

	// Script is the script text for this clip.
	Script string `json:"script,omitempty"`

	// Background is the background settings.
	Background any `json:"background,omitempty"`
}

// StudioVideoRequest is the request body for HeyGen studio video creation.
type StudioVideoRequest struct {
	// Title is the video title.
	Title string `json:"title,omitempty"`

	// Clips is the video clips.
	Clips []StudioClip `json:"clips"`

	// Dimension is the video dimensions.
	Dimension string `json:"dimension,omitempty"`

	// AspectRatio is the aspect ratio.
	AspectRatio string `json:"aspect_ratio,omitempty"`
}

// TranslateRequest is the request body for video translation.
type TranslateRequest struct {
	// VideoURL is the URL of the video to translate.
	VideoURL string `json:"video_url,omitempty"`

	// VideoBase64 is the base64-encoded video (alternative to URL).
	VideoBase64 string `json:"video_base64,omitempty"`

	// TargetLanguage is the target language code.
	TargetLanguage string `json:"target_language"`

	// SourceLanguage is the source language code (auto-detected if omitted).
	SourceLanguage string `json:"source_language,omitempty"`
}

// Avatar is a HeyGen avatar.
type Avatar struct {
	// AvatarID is the avatar identifier.
	AvatarID string `json:"avatar_id"`

	// Name is the avatar name.
	Name string `json:"name,omitempty"`

	// Gender is the avatar gender.
	Gender string `json:"gender,omitempty"`

	// PreviewURL is the preview image URL.
	PreviewURL string `json:"preview_url,omitempty"`

	// Extra contains additional fields.
	Extra map[string]any `json:"extra,omitempty"`
}

// AvatarsResponse is the response from listing HeyGen avatars.
type AvatarsResponse struct {
	Avatars []Avatar `json:"avatars"`
}

// VideoTemplate is a HeyGen video template.
type VideoTemplate struct {
	// TemplateID is the template identifier.
	TemplateID string `json:"template_id"`

	// Name is the template name.
	Name string `json:"name,omitempty"`

	// PreviewURL is the preview image URL.
	PreviewURL string `json:"preview_url,omitempty"`

	// Extra contains additional fields.
	Extra map[string]any `json:"extra,omitempty"`
}

// VideoTemplatesResponse is the response from listing HeyGen video templates.
type VideoTemplatesResponse struct {
	Templates []VideoTemplate `json:"templates"`
}

// HeyGenVoice is a HeyGen voice.
type HeyGenVoice struct {
	// VoiceID is the voice identifier.
	VoiceID string `json:"voice_id"`

	// Name is the voice name.
	Name string `json:"name,omitempty"`

	// Language is the voice language.
	Language string `json:"language,omitempty"`

	// Gender is the voice gender.
	Gender string `json:"gender,omitempty"`

	// Extra contains additional fields.
	Extra map[string]any `json:"extra,omitempty"`
}

// VideoStudioRequest is the request body for creating a HeyGen talking-head video.
// This is a long-running operation routed through the async job system.
type VideoStudioRequest struct {
	// AvatarID is the HeyGen avatar to use (required).
	AvatarID string `json:"avatar_id"`

	// Script is the text for the avatar to speak (required).
	Script string `json:"script"`

	// VoiceID is the voice for the avatar (required).
	VoiceID string `json:"voice_id"`
}

// JobAcceptedResponse is the response for async job endpoints.
// The client should poll the job status using GetJob.
type JobAcceptedResponse struct {
	// JobID is the unique job identifier for polling.
	JobID string `json:"job_id"`

	// Status is the initial job status (e.g. "pending").
	Status string `json:"status"`

	// Type is the job type.
	Type string `json:"type"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// VideoTranslateRequest is the request body for translating a video to another language.
type VideoTranslateRequest struct {
	// VideoURL is the URL of the source video (required).
	VideoURL string `json:"video_url"`

	// OutputLanguage is the target language code (required).
	OutputLanguage string `json:"output_language"`

	// SourceLanguage is the source language code (auto-detected if empty).
	SourceLanguage string `json:"source_language,omitempty"`

	// Title is an optional title for the translated video.
	Title string `json:"title,omitempty"`
}

// PhotoAvatarRequest is the request body for creating a photo avatar video.
type PhotoAvatarRequest struct {
	// PhotoBase64 is the base64-encoded photo.
	PhotoBase64 string `json:"photo_base64,omitempty"`

	// Script is the script text for the avatar to speak.
	Script string `json:"script,omitempty"`

	// VoiceID is the voice ID.
	VoiceID string `json:"voice_id,omitempty"`

	// AspectRatio is the aspect ratio.
	AspectRatio string `json:"aspect_ratio,omitempty"`

	// Name is the avatar name.
	Name string `json:"name,omitempty"`

	// Age is the target age (e.g. "young", "middle-aged").
	Age string `json:"age,omitempty"`

	// Gender is the avatar gender.
	Gender string `json:"gender,omitempty"`

	// Ethnicity is the avatar ethnicity.
	Ethnicity string `json:"ethnicity,omitempty"`

	// Orientation is the face orientation.
	Orientation string `json:"orientation,omitempty"`

	// Pose is the avatar pose.
	Pose string `json:"pose,omitempty"`

	// Style is the visual style.
	Style string `json:"style,omitempty"`

	// Appearance describes the desired appearance.
	Appearance string `json:"appearance,omitempty"`
}

// DigitalTwinRequest is the request body for digital twin video generation.
type DigitalTwinRequest struct {
	// AvatarID is the digital twin / avatar ID.
	AvatarID string `json:"avatar_id,omitempty"`

	// Script is the script text.
	Script string `json:"script,omitempty"`

	// VoiceID is the voice ID (uses twin's default voice if omitted).
	VoiceID string `json:"voice_id,omitempty"`

	// AspectRatio is the aspect ratio.
	AspectRatio string `json:"aspect_ratio,omitempty"`

	// VideoURL is the URL of the training video.
	VideoURL string `json:"video_url,omitempty"`

	// ConsentVideoURL is the URL of the consent video.
	ConsentVideoURL string `json:"consent_video_url,omitempty"`

	// Name is the digital twin name.
	Name string `json:"name,omitempty"`

	// Description describes the digital twin.
	Description string `json:"description,omitempty"`

	// AvatarGroupID is the avatar group to add the twin to.
	AvatarGroupID string `json:"avatar_group_id,omitempty"`

	// CallbackURL is a webhook URL for completion notification.
	CallbackURL string `json:"callback_url,omitempty"`
}

// HeyGenAvatarsResponse is the response from listing HeyGen avatars.
type HeyGenAvatarsResponse struct {
	// Avatars is the list of available avatars.
	Avatars []any `json:"avatars"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// HeyGenTemplatesResponse is the response from listing HeyGen templates.
type HeyGenTemplatesResponse struct {
	// Templates is the list of available templates.
	Templates []any `json:"templates"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// HeyGenVoicesResponse is the response from listing HeyGen voices.
type HeyGenVoicesResponse struct {
	// Voices is the list of available HeyGen voices.
	Voices []any `json:"voices"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// VideoStudio creates a HeyGen talking-head video.
// Returns a job that should be polled for completion.
func (c *Client) VideoStudio(ctx context.Context, req *VideoStudioRequest) (*JobAcceptedResponse, error) {
	var resp JobAcceptedResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/video/studio", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoTranslate submits a video for translation into another language.
// Returns a job that should be polled for completion.
func (c *Client) VideoTranslate(ctx context.Context, req *VideoTranslateRequest) (*JobAcceptedResponse, error) {
	var resp JobAcceptedResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/video/translate", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoPhotoAvatar submits a photo avatar creation job.
// Returns a job that should be polled for completion.
func (c *Client) VideoPhotoAvatar(ctx context.Context, req *PhotoAvatarRequest) (*JobAcceptedResponse, error) {
	var resp JobAcceptedResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/video/photo-avatar", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoDigitalTwin submits a digital twin creation job.
// Returns a job that should be polled for completion.
func (c *Client) VideoDigitalTwin(ctx context.Context, req *DigitalTwinRequest) (*JobAcceptedResponse, error) {
	var resp JobAcceptedResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/video/digital-twin", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoAvatars lists all available HeyGen avatars.
func (c *Client) VideoAvatars(ctx context.Context) (*HeyGenAvatarsResponse, error) {
	var resp HeyGenAvatarsResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/video/avatars", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoTemplates lists all available HeyGen video templates.
func (c *Client) VideoTemplates(ctx context.Context) (*HeyGenTemplatesResponse, error) {
	var resp HeyGenTemplatesResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/video/templates", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// VideoHeyGenVoices lists all available HeyGen voices.
func (c *Client) VideoHeyGenVoices(ctx context.Context) (*HeyGenVoicesResponse, error) {
	var resp HeyGenVoicesResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/video/heygen-voices", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
