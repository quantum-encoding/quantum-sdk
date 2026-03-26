package qai

import "context"

// TtsRequest is the request body for text-to-speech (sdk-graph canonical name).
type TtsRequest = TTSRequest

// TtsResponse is the response from text-to-speech (sdk-graph canonical name).
type TtsResponse = TTSResponse

// SttRequest is the request body for speech-to-text (sdk-graph canonical name).
type SttRequest = STTRequest

// SttResponse is the response from speech-to-text (sdk-graph canonical name).
type SttResponse = STTResponse

// TextToSpeechRequest is a canonical alias for TTSRequest (cross-SDK parity).
type TextToSpeechRequest = TTSRequest

// TextToSpeechResponse is a canonical alias for TTSResponse (cross-SDK parity).
type TextToSpeechResponse = TTSResponse

// SpeechToTextRequest is a canonical alias for STTRequest (cross-SDK parity).
type SpeechToTextRequest = STTRequest

// SpeechToTextResponse is a canonical alias for STTResponse (cross-SDK parity).
type SpeechToTextResponse = STTResponse

// TTSRequest is the request body for text-to-speech.
type TTSRequest struct {
	// Model is the TTS model (e.g. "tts-1", "eleven_multilingual_v2", "grok-3-tts").
	Model string `json:"model"`

	// Text is the text to synthesise into speech.
	Text string `json:"text"`

	// Voice is the voice to use (e.g. "alloy", "echo", "nova", "Rachel").
	Voice string `json:"voice,omitempty"`

	// OutputFormat is the audio format (e.g. "mp3", "wav", "opus"). Default: "mp3".
	OutputFormat string `json:"format,omitempty"`

	// Speed controls the speech rate (provider-dependent).
	Speed *float64 `json:"speed,omitempty"`
}

// TTSResponse is the response from text-to-speech.
type TTSResponse struct {
	// AudioBase64 is the base64-encoded audio data.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format (e.g. "mp3").
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// Model is the model that generated the audio.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// STTRequest is the request body for speech-to-text.
type STTRequest struct {
	// Model is the STT model (e.g. "whisper-1", "scribe_v2").
	Model string `json:"model"`

	// AudioBase64 is the base64-encoded audio data.
	AudioBase64 string `json:"audio_base64"`

	// Filename is the original filename (helps with format detection). Default: "audio.mp3".
	Filename string `json:"filename,omitempty"`

	// Language is the BCP-47 language code hint (e.g. "en", "de").
	Language string `json:"language,omitempty"`
}

// STTResponse is the response from speech-to-text.
type STTResponse struct {
	// Text is the transcribed text.
	Text string `json:"text"`

	// Model is the model that performed transcription.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// MusicRequest is the request body for music generation.
type MusicRequest struct {
	// Model is the music generation model (e.g. "lyria").
	Model string `json:"model"`

	// Prompt describes the music to generate.
	Prompt string `json:"prompt"`

	// DurationSeconds is the target duration in seconds (default 30).
	DurationSeconds int `json:"duration_seconds,omitempty"`
}

// MusicResponse is the response from music generation.
type MusicResponse struct {
	// AudioClips contains the generated music clips.
	AudioClips []MusicClip `json:"audio_clips"`

	// Model is the model that generated the music.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// MusicClip is a single generated music clip.
type MusicClip struct {
	// Base64Data is the base64-encoded audio data.
	Base64Data string `json:"base64"`

	// Format is the audio format (e.g. "mp3", "wav").
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// Index is the clip index within the batch.
	Index int `json:"index"`
}

// SoundEffectRequest is the request body for sound effects generation.
type SoundEffectRequest struct {
	// Prompt describes the sound effect to generate.
	Prompt string `json:"prompt"`

	// DurationSeconds is the optional target duration in seconds.
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
}

// SoundEffectResponse is the response from sound effects generation.
type SoundEffectResponse struct {
	// AudioBase64 is the base64-encoded audio data.
	AudioBase64 string `json:"audio_base64"`

	// Format is the audio format (e.g. "mp3").
	Format string `json:"format"`

	// SizeBytes is the audio file size.
	SizeBytes int `json:"size_bytes"`

	// Model is the model used for generation.
	Model string `json:"model"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// Speak generates speech from text.
func (c *Client) Speak(ctx context.Context, req *TTSRequest) (*TTSResponse, error) {
	var resp TTSResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/tts", req, &resp)
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

// Transcribe converts speech to text.
func (c *Client) Transcribe(ctx context.Context, req *STTRequest) (*STTResponse, error) {
	var resp STTResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/stt", req, &resp)
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

// SoundEffects generates sound effects from a text prompt.
func (c *Client) SoundEffects(ctx context.Context, req *SoundEffectRequest) (*SoundEffectResponse, error) {
	var resp SoundEffectResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/sound-effects", req, &resp)
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

// GenerateMusic generates music from a text prompt.
func (c *Client) GenerateMusic(ctx context.Context, req *MusicRequest) (*MusicResponse, error) {
	var resp MusicResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/audio/music", req, &resp)
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
