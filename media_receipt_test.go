package qai

// Tests for the media receipt fields on video and music generation:
// duration_seconds on both, and usage on video.
//
// Decoding is exercised directly against the response structs, since these
// fields are body-only — no header carries them. Wire shapes are pinned to
// routes_media.go (videoGenResponse, musicGenResponse, mediaTokenUsage,
// mediaTokenUsageOf).

import (
	"encoding/json"
	"testing"
)

// A token-billed video: on Gemini Omni the tokens are the whole cost basis,
// so every bucket has to survive the wire.
func TestVideoReceiptCarriesDurationAndEveryUsageBucket(t *testing.T) {
	body := `{
		"videos": [{"base64":"AAAA","format":"mp4","size_bytes":184320,"index":0}],
		"model": "gemini-omni-video",
		"duration_seconds": 8.5,
		"usage": {
			"prompt_tokens": 412,
			"completion_tokens": 49232,
			"reasoning_tokens": 96,
			"cached_tokens": 128,
			"total_tokens": 49868
		},
		"cost_ticks": 1247000000,
		"balance_after": 73,
		"request_id": "qai_req_2f1c8ab0-91d"
	}`

	var resp VideoResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.DurationSeconds != 8.5 {
		t.Errorf("DurationSeconds = %v, want 8.5", resp.DurationSeconds)
	}
	if resp.Usage == nil {
		t.Fatal("Usage is nil, want the reported buckets")
	}
	if got := *resp.Usage; got != (MediaTokenUsage{
		PromptTokens:     412,
		CompletionTokens: 49232,
		ReasoningTokens:  96,
		CachedTokens:     128,
		TotalTokens:      49868,
	}) {
		t.Errorf("Usage = %+v, want the five reported buckets", got)
	}
}

// A per-second model reports no tokens and the gateway sends no usage object
// at all. A zeroed struct here would read as a token-billed call that spent
// nothing, and would report a 0% cache hit rate on a model that has no cache.
// A gateway predating these fields must still decode.
func TestVideoReceiptWithoutThemLeavesUsageNil(t *testing.T) {
	body := `{
		"videos": [{"base64":"AAAA","format":"mp4","size_bytes":184320,"index":0}],
		"model": "veo-2",
		"cost_ticks": 3200000000,
		"balance_after": 41,
		"request_id": "qai_req_7d5e0c14-33a"
	}`

	var resp VideoResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.Usage != nil {
		t.Errorf("Usage = %+v, want nil — a usage block was invented", resp.Usage)
	}
	if resp.DurationSeconds != 0 {
		t.Errorf("DurationSeconds = %v, want 0 — a duration was invented", resp.DurationSeconds)
	}
}

// A usage object the provider did populate round-trips back to the same wire
// shape: buckets it reported nothing for stay off the wire rather than
// re-emerging as zeros.
func TestUsageRoundTripsWithoutInventingBuckets(t *testing.T) {
	var resp VideoResponse
	in := `{"videos":[],"model":"gemini-omni-video","duration_seconds":4,` +
		`"usage":{"completion_tokens":23168},` +
		`"cost_ticks":1,"request_id":"r"}`
	if err := json.Unmarshal([]byte(in), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	out, err := json.Marshal(resp.Usage)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"completion_tokens":23168}` {
		t.Errorf("usage re-encoded as %s, want only the reported bucket", out)
	}
}

// Music is duration-metered, so the generated length is the basis of the
// charge and has to survive the wire.
func TestMusicReceiptCarriesTheGeneratedDuration(t *testing.T) {
	body := `{
		"audio_clips": [{"base64":"SUQz","format":"mp3","size_bytes":2941184,"index":0}],
		"model": "lyria-002",
		"duration_seconds": 184.0,
		"cost_ticks": 1840000000,
		"balance_after": 57,
		"request_id": "qai_req_bb31f907-4c1"
	}`

	var resp MusicResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.DurationSeconds != 184.0 {
		t.Errorf("DurationSeconds = %v, want 184", resp.DurationSeconds)
	}
}

// A provider that reports no length leaves the field off the wire entirely,
// and a receipt from a gateway predating it must still decode.
func TestMusicReceiptWithoutADurationDecodes(t *testing.T) {
	body := `{
		"audio_clips": [],
		"model": "eleven-music",
		"cost_ticks": 600000000,
		"request_id": "r"
	}`

	var resp MusicResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.DurationSeconds != 0 {
		t.Errorf("DurationSeconds = %v, want 0 — a duration was invented", resp.DurationSeconds)
	}
}
