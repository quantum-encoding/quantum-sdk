package qai

import "context"

// DialogueRequest is the request body for multi-speaker dialogue generation.
type DialogueRequest struct {
	// Text is the dialogue script with speaker annotations (required).
	Text string `json:"text"`

	// Voices maps speaker names to voice IDs (required).
	Voices []DialogueVoice `json:"voices"`

	// Model is the dialogue model (e.g. "eleven_v3"). Optional.
	Model string `json:"model,omitempty"`

	// OutputFormat is the audio format (e.g. "mp3"). Optional.
	OutputFormat string `json:"output_format,omitempty"`

	// Seed for reproducible generation. Optional.
	Seed int `json:"seed,omitempty"`
}

// DialogueVoice maps a speaker name to a voice ID.
type DialogueVoice struct {
	// VoiceID is the ElevenLabs voice identifier.
	VoiceID string `json:"voice_id"`

	// Name is the speaker name as it appears in the text.
	Name string `json:"name"`
}

// DialogueResponse is the response from dialogue generation.
type DialogueResponse struct {
	// AudioBase64 is the base64-encoded audio data.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format (e.g. "mp3").
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// Model is the model used.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// SpeechToSpeechRequest is the request body for voice conversion.
type SpeechToSpeechRequest struct {
	// VoiceID is the target voice identifier (required).
	VoiceID string `json:"voice_id"`

	// AudioBase64 is the base64-encoded source audio (required).
	AudioBase64 string `json:"audio_base64"`
}

// SpeechToSpeechResponse is the response from speech-to-speech conversion.
type SpeechToSpeechResponse struct {
	// AudioBase64 is the base64-encoded converted audio.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format.
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// Model is the model used.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// IsolateVoiceRequest is the request body for voice isolation.
type IsolateVoiceRequest struct {
	// AudioBase64 is the base64-encoded audio with background noise (required).
	AudioBase64 string `json:"audio_base64"`

	// Filename is the original filename (helps with format detection).
	Filename string `json:"filename,omitempty"`
}

// IsolateVoiceResponse is the response from voice isolation.
type IsolateVoiceResponse struct {
	// AudioBase64 is the base64-encoded clean speech.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format.
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// RemixVoiceRequest is the request body for voice remixing.
type RemixVoiceRequest struct {
	// AudioBase64 is the base64-encoded source audio (required).
	AudioBase64 string `json:"audio_base64"`

	// Filename is the original filename.
	Filename string `json:"filename,omitempty"`

	// Gender is the target gender ("male", "female").
	Gender string `json:"gender,omitempty"`

	// Accent is the target accent (e.g. "british", "american").
	Accent string `json:"accent,omitempty"`

	// Style is the speaking style (e.g. "casual", "formal").
	Style string `json:"style,omitempty"`

	// Pacing is the speech pacing (e.g. "slow", "fast").
	Pacing string `json:"pacing,omitempty"`

	// AudioQuality is the output quality (e.g. "high", "low").
	AudioQuality string `json:"audio_quality,omitempty"`

	// PromptStrength controls how much the voice attributes are applied.
	PromptStrength string `json:"prompt_strength,omitempty"`

	// Script is the text to speak with the remixed voice.
	Script string `json:"script,omitempty"`
}

// RemixVoiceResponse is the response from voice remixing.
type RemixVoiceResponse struct {
	// AudioBase64 is the base64-encoded remixed audio.
	AudioBase64 string `json:"audio_base64,omitempty"`

	// Format is the audio format.
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// VoiceID is the generated voice ID (if a new voice was created).
	VoiceID string `json:"voice_id,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// DubRequest is the request body for audio/video dubbing.
type DubRequest struct {
	// AudioBase64 is the base64-encoded source audio (provide this or SourceURL).
	AudioBase64 string `json:"audio_base64,omitempty"`

	// Filename is the original filename.
	Filename string `json:"filename,omitempty"`

	// SourceURL is a URL to the source media (provide this or AudioBase64).
	SourceURL string `json:"source_url,omitempty"`

	// SourceLang is the source language code (auto-detected if empty).
	SourceLang string `json:"source_lang,omitempty"`

	// TargetLang is the target language code (required).
	TargetLang string `json:"target_lang"`

	// NumSpeakers is the expected number of speakers (optional).
	NumSpeakers int `json:"num_speakers,omitempty"`

	// HighestResolution enables highest quality processing.
	HighestResolution bool `json:"highest_resolution,omitempty"`
}

// DubResponse is the response from dubbing.
type DubResponse struct {
	// DubbingID is the dubbing job identifier.
	DubbingID string `json:"dubbing_id"`

	// AudioBase64 is the base64-encoded dubbed audio.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format.
	Format string `json:"format"`

	// TargetLang is the target language.
	TargetLang string `json:"target_lang"`

	// Status is the dubbing status.
	Status string `json:"status"`

	// ProcessingTimeSeconds is the time taken to process.
	ProcessingTimeSeconds float64 `json:"processing_time_seconds"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// AlignRequest is the request body for forced alignment (word-level timestamps).
type AlignRequest struct {
	// AudioBase64 is the base64-encoded audio (required).
	AudioBase64 string `json:"audio_base64"`

	// Filename is the original filename.
	Filename string `json:"filename,omitempty"`

	// Text is the transcript to align against the audio (required).
	Text string `json:"text"`

	// Language is the BCP-47 language code hint.
	Language string `json:"language,omitempty"`
}

// AlignResponse is the response from forced alignment.
type AlignResponse struct {
	// Alignment is the list of aligned words with timestamps.
	Alignment []AlignedWord `json:"alignment"`

	// Model is the model used.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// AlignedWord is a single word with timing information.
type AlignedWord struct {
	// Text is the word text.
	Text string `json:"text"`

	// StartTime is the start time in seconds.
	StartTime float64 `json:"start_time"`

	// EndTime is the end time in seconds.
	EndTime float64 `json:"end_time"`

	// Confidence is the alignment confidence score.
	Confidence float64 `json:"confidence"`
}

// VoiceDesignRequest is the request body for generating voice previews from a description.
type VoiceDesignRequest struct {
	// VoiceDescription describes the desired voice characteristics (required).
	VoiceDescription string `json:"voice_description"`

	// SampleText is the text to speak in the preview (required).
	SampleText string `json:"sample_text"`
}

// VoiceDesignResponse is the response from voice design.
type VoiceDesignResponse struct {
	// Previews is the list of generated voice previews.
	Previews []VoicePreview `json:"previews"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// VoicePreview is a single voice preview from voice design.
type VoicePreview struct {
	// GeneratedVoiceID is the ID that can be used to save this voice.
	GeneratedVoiceID string `json:"generated_voice_id"`

	// AudioBase64 is the base64-encoded preview audio.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format.
	Format string `json:"format"`
}

// StarfishTTSRequest is the request body for HeyGen's Starfish text-to-speech.
type StarfishTTSRequest struct {
	// Text is the text to speak (required).
	Text string `json:"text"`

	// VoiceID is the HeyGen voice identifier (required).
	VoiceID string `json:"voice_id"`

	// InputType is the input type (e.g. "text", "ssml").
	InputType string `json:"input_type,omitempty"`

	// Speed controls the speech rate.
	Speed float64 `json:"speed,omitempty"`

	// Language is the BCP-47 language code.
	Language string `json:"language,omitempty"`

	// Locale is the locale code.
	Locale string `json:"locale,omitempty"`
}

// StarfishTTSResponse is the response from Starfish TTS.
type StarfishTTSResponse struct {
	// AudioBase64 is the base64-encoded audio (may be empty if URL is provided).
	AudioBase64 string `json:"audio_base64,omitempty"`

	// URL is a direct URL to the audio file (may be provided instead of base64).
	URL string `json:"url,omitempty"`

	// Format is the audio format.
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// Duration is the audio duration in seconds.
	Duration float64 `json:"duration"`

	// Model is the model used.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// Dialogue generates multi-speaker dialogue audio from a script.
func (c *Client) Dialogue(ctx context.Context, req *DialogueRequest) (*DialogueResponse, error) {
	var resp DialogueResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/dialogue", req, &resp)
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

// SpeechToSpeech converts speech audio to a different voice.
func (c *Client) SpeechToSpeech(ctx context.Context, req *SpeechToSpeechRequest) (*SpeechToSpeechResponse, error) {
	var resp SpeechToSpeechResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/speech-to-speech", req, &resp)
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

// IsolateVoice removes background noise and returns clean speech.
func (c *Client) IsolateVoice(ctx context.Context, req *IsolateVoiceRequest) (*IsolateVoiceResponse, error) {
	var resp IsolateVoiceResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/isolate", req, &resp)
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

// RemixVoice transforms a voice by modifying attributes like gender, accent, and style.
func (c *Client) RemixVoice(ctx context.Context, req *RemixVoiceRequest) (*RemixVoiceResponse, error) {
	var resp RemixVoiceResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/remix", req, &resp)
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

// Dub submits audio or video for dubbing into a target language.
func (c *Client) Dub(ctx context.Context, req *DubRequest) (*DubResponse, error) {
	var resp DubResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/dub", req, &resp)
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

// Align performs forced alignment to get word-level timestamps for audio and text.
func (c *Client) Align(ctx context.Context, req *AlignRequest) (*AlignResponse, error) {
	var resp AlignResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/align", req, &resp)
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

// VoiceDesign generates voice previews from a text description of desired characteristics.
func (c *Client) VoiceDesign(ctx context.Context, req *VoiceDesignRequest) (*VoiceDesignResponse, error) {
	var resp VoiceDesignResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/voice-design", req, &resp)
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

// StarfishTTS generates speech using HeyGen's Starfish text-to-speech model.
func (c *Client) StarfishTTS(ctx context.Context, req *StarfishTTSRequest) (*StarfishTTSResponse, error) {
	var resp StarfishTTSResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/starfish-tts", req, &resp)
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
