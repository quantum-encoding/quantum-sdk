package qai

// Tests for the TTS request contract on POST /qai/v1/audio/tts, and the voice
// catalogue on GET /qai/v1/voices.
//
// Gemini exposes no parameters for tone, accent or pace — the steering is
// prose in Instructions plus inline tags inside Text — so these tests pin the
// field names the handler actually decodes. Wire shapes come from
// internal/server/routes_media.go (ttsRequest, ttsVoiceSettings, ttsSpeaker)
// and internal/server/routes_voice.go (voiceResponse).

import (
	"encoding/json"
	"testing"
)

func ttsBody(t *testing.T, req *TTSRequest) map[string]any {
	t.Helper()
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

// The house default: text alone is a complete request, and the gateway
// supplies gemini-3.1-flash-tts-preview + Laomedeia. Sending an empty model
// would pin the request to a model that does not exist.
func TestTTSRequest_TextAloneIsACompleteRequest(t *testing.T) {
	m := ttsBody(t, &TTSRequest{Text: "Hello"})

	if m["text"] != "Hello" {
		t.Errorf("text = %v", m["text"])
	}
	for _, k := range []string{"model", "voice", "speakers", "voice_settings", "instructions"} {
		if _, ok := m[k]; ok {
			t.Errorf("%s sent on a request that set none", k)
		}
	}
}

// Every steering field rides under the name the handler decodes.
func TestTTSRequest_SteeringFieldsUseTheWireNames(t *testing.T) {
	speed := 1.1
	m := ttsBody(t, &TTSRequest{
		Model:        "gemini-3.1-flash-tts-preview",
		Text:         "[excited] Hi! [whispers] can you keep a secret?",
		Voice:        "Laomedeia",
		OutputFormat: "wav",
		Speed:        &speed,
		Instructions: "Read aloud with a natural British accent",
		Language:     "en-GB",
		SampleRate:   24000,
		BitRate:      128000,
	})

	if m["model"] != "gemini-3.1-flash-tts-preview" {
		t.Errorf("model = %v", m["model"])
	}
	if m["voice"] != "Laomedeia" {
		t.Errorf("voice = %v", m["voice"])
	}
	// OutputFormat rides as "format" — the handler reads no other key.
	if m["format"] != "wav" {
		t.Errorf("format = %v, want wav", m["format"])
	}
	if _, ok := m["output_format"]; ok {
		t.Error("output_format sent; the handler does not read that key")
	}
	if m["speed"] != 1.1 {
		t.Errorf("speed = %v", m["speed"])
	}
	if m["instructions"] != "Read aloud with a natural British accent" {
		t.Errorf("instructions = %v", m["instructions"])
	}
	if m["language"] != "en-GB" {
		t.Errorf("language = %v", m["language"])
	}
	if m["sample_rate"] != float64(24000) {
		t.Errorf("sample_rate = %v", m["sample_rate"])
	}
	if m["bit_rate"] != float64(128000) {
		t.Errorf("bit_rate = %v", m["bit_rate"])
	}
	// IdempotencyKey is a header, never a body field.
	if _, ok := m["IdempotencyKey"]; ok {
		t.Error("IdempotencyKey leaked into the body")
	}
}

// Two speakers, labelled to match the lines the text carries.
func TestTTSRequest_TwoSpeakerDialogue(t *testing.T) {
	m := ttsBody(t, &TTSRequest{
		Text:         "Lacey: Hi there.\nCustomer: [excited] Hi!",
		Instructions: "Lacey is calm; the customer is cheerful",
		Speakers: []TTSSpeaker{
			{Name: "Lacey", Voice: "Laomedeia"},
			{Name: "Customer", Voice: "Puck"},
		},
	})

	speakers, ok := m["speakers"].([]any)
	if !ok {
		t.Fatalf("speakers = %T, want array", m["speakers"])
	}
	if len(speakers) != 2 {
		t.Fatalf("speakers = %d, want exactly 2 — the gateway 400s on any other count", len(speakers))
	}
	first := speakers[0].(map[string]any)
	if first["name"] != "Lacey" || first["voice"] != "Laomedeia" {
		t.Errorf("speakers[0] = %v", first)
	}
	second := speakers[1].(map[string]any)
	if second["name"] != "Customer" || second["voice"] != "Puck" {
		t.Errorf("speakers[1] = %v", second)
	}
}

// An unset ElevenLabs knob is absent, not zero — 0.0 stability is a real
// setting the provider honours, so a zeroed struct silently retunes the voice.
func TestTTSVoiceSettings_OmitWhatWasNotSet(t *testing.T) {
	stability := 0.4
	boost := true
	m := ttsBody(t, &TTSRequest{
		Text: "hi",
		VoiceSettings: &TTSVoiceSettings{
			Stability:       &stability,
			UseSpeakerBoost: &boost,
		},
	})

	vs, ok := m["voice_settings"].(map[string]any)
	if !ok {
		t.Fatalf("voice_settings = %T, want object", m["voice_settings"])
	}
	if vs["stability"] != 0.4 {
		t.Errorf("stability = %v", vs["stability"])
	}
	if vs["use_speaker_boost"] != true {
		t.Errorf("use_speaker_boost = %v", vs["use_speaker_boost"])
	}
	for _, k := range []string{"similarity_boost", "style"} {
		if _, ok := vs[k]; ok {
			t.Errorf("%s sent as zero though it was never set", k)
		}
	}
}

// The voice catalogue a picker is built from.
func TestVoiceListing_DecodesEveryDocumentedField(t *testing.T) {
	body := `{
		"voices": [
			{"voice_id":"Laomedeia","name":"Laomedeia","category":"premade",
			 "provider":"gemini","model":"gemini-3.1-flash-tts-preview","is_cloned":false},
			{"voice_id":"el_7f3","name":"Rachel","category":"cloned",
			 "provider":"elevenlabs","model":"eleven_multilingual_v2","is_cloned":true,
			 "description":"warm narrator","preview_url":"https://cdn/x.mp3"}
		],
		"request_id": "qai_req_1"
	}`

	var resp VoicesResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Voices) != 2 {
		t.Fatalf("voices = %d, want 2", len(resp.Voices))
	}

	gemini := resp.Voices[0]
	if gemini.Provider != "gemini" {
		t.Errorf("provider = %q, want gemini", gemini.Provider)
	}
	// The model to pass back to Speak for this voice, so a caller never
	// hardcodes the provider-to-model mapping.
	if gemini.Model != "gemini-3.1-flash-tts-preview" {
		t.Errorf("model = %q", gemini.Model)
	}
	if gemini.IsCloned {
		t.Error("a prebuilt voice reported itself as cloned")
	}

	el := resp.Voices[1]
	if !el.IsCloned || el.Category != "cloned" {
		t.Errorf("cloned voice = %+v", el)
	}
	if el.Description != "warm narrator" || el.PreviewURL != "https://cdn/x.mp3" {
		t.Errorf("elevenlabs extras dropped: %+v", el)
	}
}
