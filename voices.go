package qai

import (
	"context"
	"fmt"
	"net/url"
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

// ---------------------------------------------------------------------------
// Voice Library (shared/community voices)
// ---------------------------------------------------------------------------

// SharedVoice is a voice from the shared voice library.
type SharedVoice struct {
	PublicOwnerID    string  `json:"public_owner_id"`
	VoiceID          string  `json:"voice_id"`
	Name             string  `json:"name"`
	Category         string  `json:"category,omitempty"`
	Description      string  `json:"description,omitempty"`
	PreviewURL       string  `json:"preview_url,omitempty"`
	Gender           string  `json:"gender,omitempty"`
	Age              string  `json:"age,omitempty"`
	Accent           string  `json:"accent,omitempty"`
	Language         string  `json:"language,omitempty"`
	UseCase          string  `json:"use_case,omitempty"`
	Rate             float64 `json:"rate,omitempty"`
	ClonedByCount    int64   `json:"cloned_by_count,omitempty"`
	FreeUsersAllowed bool    `json:"free_users_allowed,omitempty"`
}

// SharedVoicesResponse is the response from browsing the voice library.
type SharedVoicesResponse struct {
	Voices     []SharedVoice `json:"voices"`
	NextCursor string        `json:"next_cursor,omitempty"`
	HasMore    bool          `json:"has_more"`
}

// VoiceLibraryQuery contains optional filters for browsing the voice library.
type VoiceLibraryQuery struct {
	Query    string
	PageSize int
	Cursor   string
	Gender   string
	Language string
	UseCase  string
}

// AddVoiceFromLibraryRequest is the request body for adding a shared voice.
type AddVoiceFromLibraryRequest struct {
	PublicOwnerID string `json:"public_owner_id"`
	VoiceID       string `json:"voice_id"`
	Name          string `json:"name,omitempty"`
}

// AddVoiceFromLibraryResponse is the response from adding a shared voice.
type AddVoiceFromLibraryResponse struct {
	VoiceID string `json:"voice_id"`
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

// VoiceLibrary browses the shared voice library with optional filters.
func (c *Client) VoiceLibrary(ctx context.Context, q *VoiceLibraryQuery) (*SharedVoicesResponse, error) {
	path := "/qai/v1/voices/library"
	if q != nil {
		params := url.Values{}
		if q.Query != "" {
			params.Set("query", q.Query)
		}
		if q.PageSize > 0 {
			params.Set("page_size", fmt.Sprintf("%d", q.PageSize))
		}
		if q.Cursor != "" {
			params.Set("cursor", q.Cursor)
		}
		if q.Gender != "" {
			params.Set("gender", q.Gender)
		}
		if q.Language != "" {
			params.Set("language", q.Language)
		}
		if q.UseCase != "" {
			params.Set("use_case", q.UseCase)
		}
		if encoded := params.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}

	var resp SharedVoicesResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddVoiceFromLibrary adds a shared voice from the library to the user's account.
func (c *Client) AddVoiceFromLibrary(ctx context.Context, req *AddVoiceFromLibraryRequest) (*AddVoiceFromLibraryResponse, error) {
	var resp AddVoiceFromLibraryResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/voices/library/add", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
