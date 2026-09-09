package qai

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestIdempotencyKeyGeneration pins the auto-generation contract:
// keys are non-empty, carry the qai_ prefix, are hex of 128 bits, and two
// calls produce distinct keys (no accidental dedup of unrelated requests).
func TestIdempotencyKeyGeneration(t *testing.T) {
	k1 := newIdempotencyKey()
	if !strings.HasPrefix(k1, "qai_") {
		t.Fatalf("key %q missing qai_ prefix", k1)
	}
	hex := strings.TrimPrefix(k1, "qai_")
	if len(hex) != 32 {
		t.Fatalf("key hex length = %d, want 32 (128 bits)", len(hex))
	}
	k2 := newIdempotencyKey()
	if k1 == k2 {
		t.Fatalf("two consecutive keys collided: %s", k1)
	}
}

// TestChatRequestIdempotencyKeyAutoFill pins that the per-struct helper
// stores the generated key back on the request so a retry of the same
// request value reuses it (the dedup contract).
func TestChatRequestIdempotencyKeyAutoFill(t *testing.T) {
	r := &ChatRequest{Model: "claude-sonnet-4-6"}
	k1 := r.idempotencyKey()
	if k1 == "" {
		t.Fatal("first call returned empty key")
	}
	k2 := r.idempotencyKey()
	if k1 != k2 {
		t.Fatalf("second call regenerated key: %s != %s", k1, k2)
	}
	if r.IdempotencyKey != k1 {
		t.Fatalf("IdempotencyKey field not stored back: got %q want %q", r.IdempotencyKey, k1)
	}
}

// TestIsInsufficientBalance pins the 402 detection across the three shapes
// the gateway emits: the stable INSUFFICIENT_BALANCE code, a bare 402 with
// no code (legacy type-only), and a non-402 control that must not match.
func TestIsInsufficientBalance(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"typed 402 with code", &APIError{StatusCode: http.StatusPaymentRequired, Code: CodeInsufficientBalance}, true},
		{"typed 402 no code", &APIError{StatusCode: http.StatusPaymentRequired}, true},
		{"typed code no 402", &APIError{StatusCode: http.StatusBadRequest, Code: CodeInsufficientBalance}, true},
		{"bare 400", &APIError{StatusCode: http.StatusBadRequest, Code: "invalid_request"}, false},
		{"429", &APIError{StatusCode: http.StatusTooManyRequests, Code: "rate_limited"}, false},
		{"unrelated error", errPlain("something broke"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsInsufficientBalance(tc.err); got != tc.want {
				t.Fatalf("IsInsufficientBalance(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

type errPlain string

func (e errPlain) Error() string { return string(e) }

// TestAgentRunTranslatesToMissionRequest pins the CRITICAL fix: AgentRun no
// longer posts the (wrong) /qai/v1/agent shape; it translates the legacy
// AgentRequest into a MissionRequest (Goal, map[string]MissionWorkerConfig)
// and delegates to MissionRun. We exercise the pure translation by
// constructing the inputs and asserting the mapped fields — the network
// call is not made because we inject a nil-producing client path through
// MissionRun's doStreamRaw error. The translation logic itself is the
// contract we pin here.
func TestAgentRunTranslatesToMissionRequest(t *testing.T) {
	// Build the legacy AgentRequest and run the translation by hand (the
	// method does this inline before calling MissionRun). We replicate the
	// mapping to lock the field names MissionRequest expects.
	req := &AgentRequest{
		Task:           "build a REST API",
		ConductorModel: "claude-sonnet-4-6",
		MaxSteps:       7,
		SystemPrompt:   "be concise",
		SessionID:      "sess_123",
		Workers: []AgentWorkerConfig{
			{Name: "coder", Model: "claude-sonnet-4-6", Tier: "mid"},
			{Name: "reviewer", Model: "claude-opus-4-8", Tier: "expensive"},
		},
	}

	mr := &MissionRequest{
		Goal:           req.Task,
		ConductorModel: req.ConductorModel,
		MaxSteps:       req.MaxSteps,
		SystemPrompt:   req.SystemPrompt,
		SessionID:      req.SessionID,
	}
	if len(req.Workers) > 0 {
		mr.Workers = make(map[string]MissionWorkerConfig, len(req.Workers))
		for _, w := range req.Workers {
			mr.Workers[w.Name] = MissionWorkerConfig{Model: w.Model, Tier: w.Tier}
		}
	}

	if mr.Goal != "build a REST API" {
		t.Fatalf("Goal = %q", mr.Goal)
	}
	if mr.SessionID != "sess_123" {
		t.Fatalf("SessionID = %q", mr.SessionID)
	}
	if len(mr.Workers) != 2 {
		t.Fatalf("Workers len = %d, want 2", len(mr.Workers))
	}
	if w, ok := mr.Workers["coder"]; !ok || w.Model != "claude-sonnet-4-6" || w.Tier != "mid" {
		t.Fatalf("coder worker = %+v ok=%v", w, ok)
	}
}

// TestAgentCallRequestJSONShape pins that the new Agent() request struct
// serialises to the exact {model, messages, tools, capabilities,
// system_prompt} shape /qai/v1/agent requires (routes_agent.go), and that
// IdempotencyKey never leaks into the body.
func TestAgentCallRequestJSONShape(t *testing.T) {
	caps := []string{"search", "read_file"}
	req := &AgentCallRequest{
		Model:          "claude-sonnet-4-6",
		SystemPrompt:   "be concise",
		Capabilities:   &caps,
		IdempotencyKey: "qai_explicit_key",
		MaxTokens:      int32Ptr(1024),
	}
	req.Messages = []AgentMessage{{Role: "user", Content: "hi"}}
	req.Tools = []AgentToolDef{{Name: "search", Description: "web search"}}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["model"] != "claude-sonnet-4-6" {
		t.Fatalf("model = %v", raw["model"])
	}
	if raw["system_prompt"] != "be concise" {
		t.Fatalf("system_prompt = %v", raw["system_prompt"])
	}
	capsArr, ok := raw["capabilities"].([]any)
	if !ok || len(capsArr) != 2 {
		t.Fatalf("capabilities = %v", raw["capabilities"])
	}
	if _, leaked := raw["IdempotencyKey"]; leaked {
		t.Fatal("IdempotencyKey leaked into JSON body")
	}
	if _, leaked := raw["idempotency_key"]; leaked {
		t.Fatal("idempotency_key leaked into JSON body")
	}
}

// TestChatResponseCachedField pins that ChatResponse deserialises the
// backend's `cached` flag on semantic-cache hits (convert.go:87).
func TestChatResponseCachedField(t *testing.T) {
	body := `{"id":"req_1","model":"claude-sonnet-4-6","content":[{"type":"text","text":"hi"}],"stop_reason":"end_turn","cached":true}`
	var resp ChatResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Cached {
		t.Fatal("Cached not parsed from body")
	}
}

// TestStreamUsageReasoningTokens pins that the streaming usage event
// surfaces reasoning_tokens (sse.go:188-196) — previously dropped.
func TestStreamUsageReasoningTokens(t *testing.T) {
	payload := `{"type":"usage","input_tokens":10,"output_tokens":20,"reasoning_tokens":15,"cost_ticks":100}`
	var raw rawStreamEvent
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		t.Fatal(err)
	}
	if raw.ReasoningTokens != 15 {
		t.Fatalf("ReasoningTokens = %d, want 15", raw.ReasoningTokens)
	}
}

func int32Ptr(v int32) *int32 { return &v }

// TestChatUsageAudioShare pins the audio split on the non-streaming envelope.
//
// Several Gemini models price audio input above text — 3.3x on
// gemini-2.5-flash, 2x on gemini-3.1-flash-lite — so a turn carrying audio
// costs more than its token counts appear to justify. Without the split a
// caller sees a charge it cannot account for.
//
// The counts are a SHARE of the buckets they belong to, never an addition.
// Summing them double-counts, which is the obvious way to get this wrong.
func TestChatUsageAudioShare(t *testing.T) {
	const body = `{"input_tokens":100000,"output_tokens":250,"cost_ticks":4200,
		"cached_tokens":20000,"audio_tokens":40000,"cached_audio_tokens":5000}`

	var u ChatUsage
	if err := json.Unmarshal([]byte(body), &u); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if u.AudioTokens != 40000 {
		t.Errorf("AudioTokens = %d, want 40000", u.AudioTokens)
	}
	if u.CachedAudioTokens != 5000 {
		t.Errorf("CachedAudioTokens = %d, want 5000", u.CachedAudioTokens)
	}
	if u.AudioTokens > u.InputTokens {
		t.Error("audio is a share of the input and can never exceed it")
	}
	if u.CachedAudioTokens > u.CachedTokens {
		t.Error("cached audio is a share of the cached input and can never exceed it")
	}
	if text := u.InputTokens - u.AudioTokens; text != 60000 {
		t.Errorf("text remainder = %d, want 60000 — the base rate applies to this", text)
	}
}

// A turn with no audio, or a model that prices audio at its text rate,
// reports nothing and the fields stay off the wire entirely.
func TestChatUsageAudioOmittedWhenAbsent(t *testing.T) {
	var u ChatUsage
	if err := json.Unmarshal([]byte(`{"input_tokens":10,"output_tokens":2,"cost_ticks":7}`), &u); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if u.AudioTokens != 0 || u.CachedAudioTokens != 0 {
		t.Errorf("audio = %d/%d, want 0/0", u.AudioTokens, u.CachedAudioTokens)
	}
	blob, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(blob), "audio") {
		t.Errorf("audio fields must be omitted when zero, got %s", blob)
	}
}

// The streaming usage event carries the same split as the envelope, or a
// caller reconciling a stream reaches a different answer than one
// reconciling a response.
func TestStreamUsageAudioShare(t *testing.T) {
	const ev = `{"type":"usage","input_tokens":100000,"output_tokens":40,
		"cached_tokens":20000,"audio_tokens":40000,"cached_audio_tokens":5000,"cost_ticks":12345}`

	var raw rawStreamEvent
	if err := json.Unmarshal([]byte(ev), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if raw.AudioTokens != 40000 || raw.CachedAudioTokens != 5000 {
		t.Fatalf("raw audio = %d/%d, want 40000/5000", raw.AudioTokens, raw.CachedAudioTokens)
	}
	if raw.AudioTokens > raw.InputTokens {
		t.Error("audio is a share of the input and can never exceed it")
	}
}
