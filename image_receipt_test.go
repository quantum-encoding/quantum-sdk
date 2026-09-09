package qai

// Tests for the image receipt fields: the prompt the picture was actually
// made from, and the token counts a token-priced charge was computed from.
// The sync /images/generate and /images/edit routes and the async image job
// results all emit the same envelope.
//
// Decoding is exercised directly against the response struct, since these
// fields are body-only — no header carries them. Wire shapes are pinned to
// routes_media.go (imageGenResponse, imageUsage, imageUsageOf).

import (
	"encoding/json"
	"testing"
)

// A token-priced generation: the usage is the whole audit, and the revised
// prompt is the text the picture was actually made from.
func TestImageReceiptCarriesTheRewriteAndTheUsage(t *testing.T) {
	body := `{
		"images": [{"base64":"iVBOR","format":"png","index":0}],
		"model": "gpt-image-2",
		"cost_ticks": 527000000,
		"balance_after": 91,
		"request_id": "qai_req_f372518d-447",
		"revised_prompt": "A golden rubber duck wearing a black silk top hat, studio lit",
		"usage": {"prompt_tokens": 120, "completion_tokens": 1580, "total_tokens": 1700}
	}`

	var resp ImageResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	const want = "A golden rubber duck wearing a black silk top hat, studio lit"
	if resp.RevisedPrompt != want {
		t.Errorf("RevisedPrompt = %q, want %q", resp.RevisedPrompt, want)
	}
	if resp.Usage == nil {
		t.Fatal("Usage is nil, want the reported buckets")
	}
	if got := *resp.Usage; got != (ImageUsage{
		PromptTokens:     120,
		CompletionTokens: 1580,
		TotalTokens:      1700,
	}) {
		t.Errorf("Usage = %+v, want the three reported buckets", got)
	}
}

// A flat-priced model rewrites nothing and reports no tokens, so the gateway
// sends neither field. A zeroed usage struct would assert a token basis the
// charge does not have. A gateway predating these fields must still decode.
func TestImageReceiptWithoutThemLeavesUsageNil(t *testing.T) {
	body := `{
		"images": [],
		"model": "grok-imagine-image-2.0",
		"cost_ticks": 500000000,
		"balance_after": 88,
		"request_id": "qai_req_9c03b229-d09"
	}`

	var resp ImageResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.RevisedPrompt != "" {
		t.Errorf("RevisedPrompt = %q, want empty — a rewrite was invented", resp.RevisedPrompt)
	}
	if resp.Usage != nil {
		t.Errorf("Usage = %+v, want nil — a usage block was invented", resp.Usage)
	}
}

// The async job result marshals the same envelope as the sync route, so a job
// caller can check the same bill. A usage the provider partly reported
// re-encodes to exactly the buckets it reported, rather than filling the rest
// with zeros.
func TestAsyncImageJobResultDecodesTheSameReceipt(t *testing.T) {
	body := `{
		"images": [{"base64":"iVBOR","format":"png","index":0}],
		"model": "gpt-image-2",
		"cost_ticks": 162800000,
		"request_id": "qai_req_4a1e77c0-2b8",
		"revised_prompt": "A golden rubber duck, softbox lit",
		"usage": {"prompt_tokens": 96, "total_tokens": 610}
	}`

	var resp ImageResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.RevisedPrompt != "A golden rubber duck, softbox lit" {
		t.Errorf("RevisedPrompt = %q", resp.RevisedPrompt)
	}

	out, err := json.Marshal(resp.Usage)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"prompt_tokens":96,"total_tokens":610}` {
		t.Errorf("usage re-encoded as %s, want only the reported buckets", out)
	}
}
