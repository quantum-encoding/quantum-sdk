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
	// Model is the TTS model. Empty = the gateway default,
	// "gemini-3.1-flash-tts-preview", paired with the "Laomedeia" voice — so
	// Text alone is a complete request. Also "gemini-2.5-flash-preview-tts",
	// "gemini-2.5-pro-preview-tts", OpenAI "openai-tts-1" / "gpt-4o-mini-tts",
	// xAI "grok-tts", ElevenLabs "eleven_*".
	Model string `json:"model,omitempty"`

	// Text is the text to synthesise into speech. May carry inline audio tags
	// ("[whispers]", "[excited]", …) and, for dialogue, the speaker labels
	// named in Speakers.
	Text string `json:"text"`

	// Voice is the voice to use — an id from ListVoices. Gemini's default is
	// "Laomedeia". Ignored when Speakers is set.
	Voice string `json:"voice,omitempty"`

	// OutputFormat is the audio format: "mp3" (default), "wav", "opus", "pcm".
	OutputFormat string `json:"format,omitempty"`

	// Speed is the speech rate, 0.7–1.5. xAI only — on Gemini, ask for it in
	// Instructions ("at a slow, measured pace").
	Speed *float64 `json:"speed,omitempty"`

	// Instructions is style direction: tone, pace, accent, character. On
	// Gemini it is prepended to the prompt and is the main way to steer a
	// read, since Gemini exposes no knobs for any of it. On OpenAI only
	// gpt-4o-mini-tts honours it — tts-1/tts-1-hd reject the field and the
	// gateway drops it for them.
	Instructions string `json:"instructions,omitempty"`

	// Language is a BCP-47 tag, e.g. "en-GB", "es-ES", or "auto". Gemini
	// detects the language on its own; set this to pin the pronunciation or
	// accent family. Also drives xAI pronunciation, where an English default
	// sounds robotic on other languages.
	Language string `json:"language,omitempty"`

	// SampleRate is the output sample rate in Hz, e.g. 24000 or 44100. xAI only.
	SampleRate int `json:"sample_rate,omitempty"`

	// BitRate is the output bit rate in bits/sec, e.g. 128000. xAI only.
	BitRate int `json:"bit_rate,omitempty"`

	// VoiceSettings tunes ElevenLabs synthesis. Ignored by every other
	// provider.
	VoiceSettings *TTSVoiceSettings `json:"voice_settings,omitempty"`

	// Speakers turns Text into a two-voice dialogue on Gemini TTS. Each entry
	// pairs a speaker label used in Text ("Lacey: …") with the prebuilt voice
	// that reads it. EXACTLY TWO — the gateway rejects any other count with a
	// 400 — and Voice is then ignored.
	Speakers []TTSSpeaker `json:"speakers,omitempty"`

	// IdempotencyKey is sent as the Idempotency-Key header; auto-generated if empty.
	IdempotencyKey string `json:"-"`
}

// TTSVoiceSettings tunes ElevenLabs synthesis.
//
// Every field is a pointer because an absent knob leaves the provider default
// alone, which is not the same as sending 0: 0.0 stability is a real setting
// the provider honours, so a zeroed struct silently retunes the voice.
type TTSVoiceSettings struct {
	// Stability is 0.0–1.0. Lower is more expressive and less consistent.
	Stability *float64 `json:"stability,omitempty"`

	// SimilarityBoost is 0.0–1.0 — how closely to track the original voice.
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`

	// Style is 0.0–1.0 style exaggeration.
	Style *float64 `json:"style,omitempty"`

	// UseSpeakerBoost boosts resemblance to the original speaker.
	UseSpeakerBoost *bool `json:"use_speaker_boost,omitempty"`
}

// TTSSpeaker is one voice in a Gemini two-speaker dialogue.
type TTSSpeaker struct {
	// Name is the label this speaker's lines carry in the text, e.g. "Lacey"
	// for lines written as "Lacey: …".
	Name string `json:"name"`

	// Voice is the prebuilt voice that reads those lines, e.g. "Laomedeia".
	Voice string `json:"voice"`
}

// idempotencyKey returns the caller-set key, auto-generating one if empty.
func (r *TTSRequest) idempotencyKey() string {
	if r.IdempotencyKey == "" {
		r.IdempotencyKey = newIdempotencyKey()
	}
	return r.IdempotencyKey
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

	// BalanceAfter is the wallet balance after this call, from the
	// X-QAI-Balance-After header. Signed — a claw-back can make it negative.
	BalanceAfter int64 `json:"-"`

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

	// DurationSeconds is the length actually generated, when the provider
	// reports it. Music is duration-metered — Lyria per 30 seconds,
	// ElevenLabs per minute — and settlement prefers this over the requested
	// length, so it is the basis of CostTicks. Zero when the provider reports
	// no length.
	DurationSeconds float64 `json:"duration_seconds,omitempty"`

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
	resp.BalanceAfter = meta.BalanceAfter
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
