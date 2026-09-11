package qai

import (
	"bytes"
	"encoding/json"
	"testing"
)

// prompt_cache_key rides the request only when set, so a caller who never
// names one keeps the gateway's identity-derived default.
func TestChatRequest_PromptCacheKey(t *testing.T) {
	req := &ChatRequest{Model: "gpt-5.6", Messages: []ChatMessage{{Role: "user", Content: "hi"}}}

	var bare map[string]any
	mustRoundTrip(t, req, &bare)
	if _, ok := bare["prompt_cache_key"]; ok {
		t.Error("prompt_cache_key sent on a request that set none")
	}

	req.PromptCacheKey = "conv-7f3a"
	var keyed map[string]any
	mustRoundTrip(t, req, &keyed)
	if keyed["prompt_cache_key"] != "conv-7f3a" {
		t.Errorf("prompt_cache_key = %v, want conv-7f3a", keyed["prompt_cache_key"])
	}
}

// A "reasoning" block decodes with its payload untouched and re-encodes
// byte-for-byte, in the position it arrived in among the tool calls.
func TestContentBlock_ReasoningRoundTripsVerbatim(t *testing.T) {
	wire := `{"id":"req_1","model":"gpt-5.6","stop_reason":"tool_use","content":[
		{"type":"reasoning","reasoning":{"id":"rs_abc","summary":[],"encrypted_content":"Zm9v"},"minted_by":"gpt-5.6"},
		{"type":"tool_use","id":"call_1","name":"lookup","input":{"q":"x"}}
	]}`

	var resp ChatResponse
	if err := json.Unmarshal([]byte(wire), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Content) != 2 {
		t.Fatalf("content blocks = %d, want 2", len(resp.Content))
	}

	r := resp.Content[0]
	if r.Type != "reasoning" {
		t.Errorf("block[0].Type = %q, want reasoning", r.Type)
	}
	if r.MintedBy != "gpt-5.6" {
		t.Errorf("MintedBy = %q, want gpt-5.6", r.MintedBy)
	}
	// Opaque: the SDK must not have reshaped the provider's item.
	var payload map[string]any
	if err := json.Unmarshal(r.Reasoning, &payload); err != nil {
		t.Fatalf("reasoning payload: %v", err)
	}
	if payload["encrypted_content"] != "Zm9v" || payload["id"] != "rs_abc" {
		t.Errorf("reasoning payload = %v, want the provider's item verbatim", payload)
	}

	// Echoed back on the next turn's assistant message, unchanged and in
	// the same order — position is the state.
	echoed, err := json.Marshal(ChatMessage{Role: "assistant", ContentBlocks: resp.Content})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var back struct {
		ContentBlocks []struct {
			Type      string          `json:"type"`
			Reasoning json.RawMessage `json:"reasoning"`
			MintedBy  string          `json:"minted_by"`
		} `json:"content_blocks"`
	}
	if err := json.Unmarshal(echoed, &back); err != nil {
		t.Fatalf("decode echo: %v", err)
	}
	if back.ContentBlocks[0].Type != "reasoning" || back.ContentBlocks[1].Type != "tool_use" {
		t.Fatalf("echo order = %q,%q — reasoning must stay before its tool call",
			back.ContentBlocks[0].Type, back.ContentBlocks[1].Type)
	}
	if !bytes.Equal(back.ContentBlocks[0].Reasoning, r.Reasoning) {
		t.Errorf("reasoning re-encoded as %s, want the bytes that arrived: %s",
			back.ContentBlocks[0].Reasoning, r.Reasoning)
	}
	if back.ContentBlocks[0].MintedBy != "gpt-5.6" {
		t.Errorf("minted_by lost on echo")
	}
}

// A block with no reasoning state omits both fields rather than sending
// nulls a provider would reject.
func TestContentBlock_NoReasoningOmitsTheFields(t *testing.T) {
	var m map[string]any
	mustRoundTrip(t, ContentBlock{Type: "text", Text: "hello"}, &m)
	for _, k := range []string{"reasoning", "minted_by", "thought_signature"} {
		if _, ok := m[k]; ok {
			t.Errorf("%s sent on a plain text block", k)
		}
	}
}

// Gemini 3 signs a turn that ends in TEXT, not only its tool calls.
func TestContentBlock_ThoughtSignatureOnTextBlock(t *testing.T) {
	var resp ChatResponse
	// "c2ln" is base64("sig") — the wire spelling of the []byte field.
	wire := `{"id":"r","model":"gemini-3.5-flash","stop_reason":"stop","content":[
		{"type":"text","text":"hi","thought_signature":"c2ln"}]}`
	if err := json.Unmarshal([]byte(wire), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := string(resp.Content[0].ThoughtSignature); got != "sig" {
		t.Fatalf("text block signature = %q, want sig", got)
	}

	var m map[string]any
	mustRoundTrip(t, resp.Content[0], &m)
	if m["thought_signature"] != "c2ln" {
		t.Errorf("signature re-encoded as %v, want c2ln", m["thought_signature"])
	}
}

// The gateway sends a thought_signature SSE event just before done.
func TestStreamEvent_ThoughtSignature(t *testing.T) {
	var raw rawStreamEvent
	if err := json.Unmarshal([]byte(`{"type":"thought_signature","thought_signature":"c2ln"}`), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(raw.ThoughtSignature) != "sig" {
		t.Errorf("raw signature = %q, want sig", raw.ThoughtSignature)
	}
}

// provider_options is an open map: an unknown provider key, and an unknown
// key under a documented provider, both survive the round trip.
func TestChatRequest_ProviderOptionsAreOpen(t *testing.T) {
	req := &ChatRequest{
		Model:    "gpt-5.6",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
		ProviderOptions: map[string]any{
			"openai": map[string]any{
				"reasoning_summary":             "detailed",
				"reasoning_mode":                "pro",
				"verbosity":                     "low",
				"text_format":                   "json_object",
				"a_key_this_sdk_never_heard_of": 42,
			},
			"xai": map[string]any{"native_files": true},
		},
	}
	var m map[string]any
	mustRoundTrip(t, req, &m)
	opts := m["provider_options"].(map[string]any)
	openai := opts["openai"].(map[string]any)
	if openai["reasoning_mode"] != "pro" || openai["text_format"] != "json_object" {
		t.Errorf("openai options = %v", openai)
	}
	if openai["a_key_this_sdk_never_heard_of"] != float64(42) {
		t.Errorf("an unnamed key was dropped: %v", openai)
	}
	if opts["xai"].(map[string]any)["native_files"] != true {
		t.Errorf("xai.native_files = %v, want true", opts["xai"])
	}
}

// Every tier the gateway validates, "max" included, rides as typed.
func TestChatRequest_ReasoningEffortTiers(t *testing.T) {
	for _, tier := range []string{"none", "low", "medium", "high", "xhigh", "max"} {
		var m map[string]any
		mustRoundTrip(t, &ChatRequest{
			Model:           "gpt-5.6",
			Messages:        []ChatMessage{{Role: "user", Content: "hi"}},
			ReasoningEffort: tier,
		}, &m)
		if m["reasoning_effort"] != tier {
			t.Errorf("reasoning_effort = %v, want %q", m["reasoning_effort"], tier)
		}
	}
}

func mustRoundTrip(t *testing.T, v any, out any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
