package qai

import (
	"context"
	"fmt"
)

// VoiceInfo describes an available voice.
type VoiceInfo struct {
	// VoiceID is the voice identifier used in TTS requests.
	VoiceID string `json:"voice_id"`

	// Name is the human-readable voice name.
	Name string `json:"name"`

	// Category is the voice category (e.g. "premade", "cloned").
	Category string `json:"category"`

	// Description describes the voice characteristics.
	Description string `json:"description,omitempty"`

	// PreviewURL is a URL to preview the voice.
	PreviewURL string `json:"preview_url,omitempty"`
}

// VoicesResponse is the response from listing voices.
type VoicesResponse struct {
	// Voices is the list of available voices.
	Voices []VoiceInfo `json:"voices"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// CloneVoiceRequest is the request body for instant voice cloning.
type CloneVoiceRequest struct {
	// Name is the display name for the cloned voice (required).
	Name string `json:"name"`

	// Description describes the voice (optional).
	Description string `json:"description,omitempty"`

	// AudioSamples is a list of base64-encoded audio files for cloning (required).
	AudioSamples []string `json:"audio_samples"`
}

// CloneVoiceResponse is the response from voice cloning.
type CloneVoiceResponse struct {
	// VoiceID is the identifier for the newly cloned voice.
	VoiceID string `json:"voice_id"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// ListVoices returns all available voices (ElevenLabs).
func (c *Client) ListVoices(ctx context.Context) (*VoicesResponse, error) {
	var resp VoicesResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/voices", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CloneVoice creates an instant voice clone from audio samples.
func (c *Client) CloneVoice(ctx context.Context, req *CloneVoiceRequest) (*CloneVoiceResponse, error) {
	var resp CloneVoiceResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/voices/clone", req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.RequestID == "" {
		resp.RequestID = meta.RequestID
	}
	return &resp, nil
}

// DeleteVoice removes a cloned voice by its ID.
func (c *Client) DeleteVoice(ctx context.Context, id string) error {
	_, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/qai/v1/voices/%s", id), nil, nil)
	return err
}
