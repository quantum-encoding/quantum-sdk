package qai

// HeyGen Avatar Realtime (Broadcast) sessions.
//
// A realtime session makes an avatar speak live and publishes a plain HLS
// stream (720p). Sessions are PREPAID: the entire max_duration_seconds block
// is charged at create time and is NOT refunded on early cancel (cancelling
// only stops the upstream meter).
//
// Recommended flow:
//  1. CreateAvatarRealtimeSession -> stream_id
//  2. Poll GetAvatarRealtimeSession (~2s) until status == "streaming",
//     then play HLSURL
//  3. For "text_stream" sessions, append text with SendAvatarRealtimeText
//     and close with IsFinal: true (idle timeout is ~30s without new text)
//  4. CancelAvatarRealtimeSession as soon as you're done
//
// Not to be confused with the WebSocket voice realtime API in realtime.go.

import (
	"context"
	"fmt"
)

// AvatarAudioInput is the audio input union for "audio"-type realtime
// sessions, discriminated by InputType (wire field "type").
type AvatarAudioInput struct {
	// InputType is the input kind: "url" | "asset_id" | "base64".
	InputType string `json:"type"`

	// URL is a publicly accessible HTTPS URL (when InputType == "url").
	URL string `json:"url,omitempty"`

	// AssetID is a HeyGen asset id from an asset upload (when InputType == "asset_id").
	AssetID string `json:"asset_id,omitempty"`

	// MediaType is the MIME type, e.g. "audio/mpeg" (when InputType == "base64").
	MediaType string `json:"media_type,omitempty"`

	// Data is the base64-encoded audio bytes (when InputType == "base64").
	Data string `json:"data,omitempty"`
}

// AvatarRealtimeRequest is the request body for creating a live avatar
// session (prepaid).
type AvatarRealtimeRequest struct {
	// SessionType is the session kind: "tts" | "audio" | "text_stream".
	// Wire field: "type".
	SessionType string `json:"type"`

	// AvatarID is the HeyGen photo-avatar / motion-avatar look id
	// (required for all kinds).
	AvatarID string `json:"avatar_id"`

	// VoiceID is required for "tts" and "text_stream", must be omitted
	// for "audio".
	VoiceID string `json:"voice_id,omitempty"`

	// Text is the fixed script ("tts") or the initial non-empty seed
	// ("text_stream").
	Text string `json:"text,omitempty"`

	// Audio is required for "audio", must be omitted for "tts"/"text_stream".
	Audio *AvatarAudioInput `json:"audio,omitempty"`

	// MaxDurationSeconds is the prepaid block in seconds (1-3600). The whole
	// block is charged at create time; early cancel does NOT refund.
	MaxDurationSeconds int `json:"max_duration_seconds"`
}

// AvatarRealtimeCreateResponse is the response from creating a live avatar session.
type AvatarRealtimeCreateResponse struct {
	// StreamID is the session id — use in the status/text/cancel calls.
	StreamID string `json:"stream_id"`

	// Status is always "pending" at create.
	Status string `json:"status"`

	// PrepaidSeconds is the echo of max_duration_seconds.
	PrepaidSeconds int `json:"prepaid_seconds"`

	// CostTicks is the ticks charged for the prepaid block.
	CostTicks int64 `json:"cost_ticks"`

	// BalanceAfter is the post-deduction credit balance in ticks (from the
	// X-QAI-Balance-After header; Receipt Pattern).
	BalanceAfter int64 `json:"balance_after,omitempty"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// AvatarRealtimeStatusResponse is the response from a session status check.
type AvatarRealtimeStatusResponse struct {
	// StreamID is the session id.
	StreamID string `json:"stream_id"`

	// Status is "pending" | "streaming" | "completed" | "error".
	Status string `json:"status"`

	// HLSURL is the HLS .m3u8 playback URL (720p); present once streaming.
	HLSURL string `json:"hls_url,omitempty"`

	// ErrorMessage is the failure detail when Status == "error".
	ErrorMessage string `json:"error_message,omitempty"`

	// EndReason on completed text_stream sessions: "final_marker" | "idle_timeout".
	EndReason string `json:"end_reason,omitempty"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// AvatarRealtimeTextRequest is the request body for appending a text delta
// to a "text_stream" session.
type AvatarRealtimeTextRequest struct {
	// Delta is the text fragment to append (a token or coalesced batch).
	// Required unless IsFinal is true, in which case it may be empty.
	Delta string `json:"delta,omitempty"`

	// IsFinal true closes the text input (appending afterwards fails
	// upstream with a 410 provider_error). Wire field: "final".
	IsFinal bool `json:"final"`
}

// AvatarRealtimeTextResponse is the response from appending a text delta.
type AvatarRealtimeTextResponse struct {
	// OK is always true on success.
	OK bool `json:"ok"`

	// BufferedBytes is the total text bytes buffered for the session so far.
	BufferedBytes int64 `json:"buffered_bytes"`

	// IsFinal is the echo of the request's "final" flag. Wire field: "final".
	IsFinal bool `json:"final"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// AvatarRealtimeCancelResponse is the response from cancelling a session early.
type AvatarRealtimeCancelResponse struct {
	// StreamID is the session id.
	StreamID string `json:"stream_id"`

	// Cancelled true = this call initiated cancellation; false = the session
	// was already terminal (cancel is idempotent).
	Cancelled bool `json:"cancelled"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// CreateAvatarRealtimeSession creates a live avatar realtime session
// (HeyGen Broadcast).
//
// PREPAID: the entire MaxDurationSeconds block (1-3600 s) is charged at
// create time; cancelling early does NOT refund.
func (c *Client) CreateAvatarRealtimeSession(ctx context.Context, req *AvatarRealtimeRequest) (*AvatarRealtimeCreateResponse, error) {
	var resp AvatarRealtimeCreateResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/avatar/realtime", req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.CostTicks == 0 {
		resp.CostTicks = meta.CostTicks
	}
	if resp.BalanceAfter == 0 {
		resp.BalanceAfter = meta.BalanceAfter
	}
	if resp.RequestID == "" {
		resp.RequestID = meta.RequestID
	}
	return &resp, nil
}

// GetAvatarRealtimeSession gets the live status of an avatar realtime session.
//
// Poll (~2s) until Status == "streaming", then play HLSURL.
// "completed" and "error" are terminal.
func (c *Client) GetAvatarRealtimeSession(ctx context.Context, streamID string) (*AvatarRealtimeStatusResponse, error) {
	var resp AvatarRealtimeStatusResponse
	_, err := c.doJSON(ctx, "GET", fmt.Sprintf("/qai/v1/avatar/realtime/%s", streamID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendAvatarRealtimeText appends a text delta to a "text_stream" session
// (or closes it with an empty Delta and IsFinal: true).
func (c *Client) SendAvatarRealtimeText(ctx context.Context, streamID string, req *AvatarRealtimeTextRequest) (*AvatarRealtimeTextResponse, error) {
	var resp AvatarRealtimeTextResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/avatar/realtime/%s/text", streamID), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelAvatarRealtimeSession terminates an avatar realtime session early
// (idempotent; no refund — this only stops HeyGen's upstream meter).
func (c *Client) CancelAvatarRealtimeSession(ctx context.Context, streamID string) (*AvatarRealtimeCancelResponse, error) {
	var resp AvatarRealtimeCancelResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/avatar/realtime/%s/cancel", streamID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
