package qai

// HeyGen sounds search (background music + sound effects).

import (
	"context"
	"fmt"
	"net/url"
)

// AudioSoundsQuery holds the query parameters for searching the sounds catalog.
type AudioSoundsQuery struct {
	// Query is a natural-language description of the sound wanted (required).
	Query string

	// SoundType is the catalog to search: "music" | "sound_effects"
	// (API default: "music"). Wire param: "type".
	SoundType string

	// Limit is the max results, 1-50 (API default 10). Zero means unset.
	Limit int

	// MinScore is the minimum similarity score, 0-1 (API default 0.7).
	// Nil means unset.
	MinScore *float64

	// Token is an opaque cursor from a previous response's next_token.
	Token string
}

// AudioSound is a track from the sounds catalog.
type AudioSound struct {
	// ID is the track identifier.
	ID string `json:"id"`

	// Name is the track name.
	Name string `json:"name"`

	// Description is the track description.
	Description string `json:"description"`

	// AudioURL is a pre-signed WAV URL with a limited lifetime — download
	// promptly, do not cache.
	AudioURL string `json:"audio_url"`

	// Duration is the duration in seconds.
	Duration float64 `json:"duration"`

	// Score is the similarity score 0-1 (best first).
	Score float64 `json:"score"`

	// SoundType is "music" | "sound_effects". Wire field: "type".
	SoundType string `json:"type"`
}

// AudioSoundsResponse is the response from searching the sounds catalog (unbilled).
type AudioSoundsResponse struct {
	// Sounds contains the matching tracks, best score first (empty page -> []).
	Sounds []AudioSound `json:"sounds"`

	// HasMore indicates more pages exist.
	HasMore bool `json:"has_more"`

	// NextToken should be passed as Token for the next page (may be empty).
	NextToken string `json:"next_token"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// SearchAudioSounds searches HeyGen's background-music and sound-effects
// catalogs (semantic ranking, best score first). Unbilled catalog route.
func (c *Client) SearchAudioSounds(ctx context.Context, query *AudioSoundsQuery) (*AudioSoundsResponse, error) {
	params := url.Values{}
	params.Set("query", query.Query)
	if query.SoundType != "" {
		params.Set("type", query.SoundType)
	}
	if query.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", query.Limit))
	}
	if query.MinScore != nil {
		params.Set("min_score", fmt.Sprintf("%g", *query.MinScore))
	}
	if query.Token != "" {
		params.Set("token", query.Token)
	}
	path := "/qai/v1/audio/sounds?" + params.Encode()

	var resp AudioSoundsResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
