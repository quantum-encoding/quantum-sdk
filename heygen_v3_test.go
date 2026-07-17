package qai

// Tests for the HeyGen v3 gateway surface: avatar realtime sessions,
// sounds search, template detail/render, and batch videos.
//
// All tests run against a mock gateway (httptest) — production is never
// called. Wire shapes are pinned to the HeyGen v3 wire contract
// (routes_heygen_v3.go).

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient starts a mock gateway that records the request and replies
// with the given status/headers/body, returning a Client pointed at it.
func newTestClient(t *testing.T, status int, headers map[string]string, body string, capture *capturedRequest) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capture != nil {
			capture.Method = r.Method
			capture.Path = r.URL.Path
			capture.Query = r.URL.Query()
			capture.Auth = r.Header.Get("Authorization")
			capture.ContentType = r.Header.Get("Content-Type")
			b, _ := io.ReadAll(r.Body)
			capture.Body = b
		}
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return New("qai_test_key", WithBaseURL(srv.URL))
}

type capturedRequest struct {
	Method      string
	Path        string
	Query       map[string][]string
	Auth        string
	ContentType string
	Body        []byte
}

func (c *capturedRequest) bodyJSON(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(c.Body, &m); err != nil {
		t.Fatalf("request body is not JSON: %v (%s)", err, c.Body)
	}
	return m
}

// --- 1. POST /qai/v1/avatar/realtime -------------------------------------

func TestCreateAvatarRealtimeSession(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, map[string]string{
		"X-QAI-Cost-Ticks":    "345000000000",
		"X-QAI-Balance-After": "655000000000",
	}, `{"stream_id":"rt_9f2c1a","status":"pending","prepaid_seconds":300,"cost_ticks":345000000000,"request_id":"req_abc123"}`, &cap)

	resp, err := client.CreateAvatarRealtimeSession(t.Context(), &AvatarRealtimeRequest{
		SessionType:        "text_stream",
		AvatarID:           "Abigail_expressive_2024112501",
		VoiceID:            "73c0b6a2",
		Text:               "Hello!",
		MaxDurationSeconds: 300,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/avatar/realtime" {
		t.Fatalf("request = %s %s, want POST /qai/v1/avatar/realtime", cap.Method, cap.Path)
	}
	if cap.Auth != "Bearer qai_test_key" {
		t.Fatalf("Authorization = %q", cap.Auth)
	}
	body := cap.bodyJSON(t)
	if body["type"] != "text_stream" {
		t.Fatalf("body type = %v", body["type"])
	}
	if body["avatar_id"] != "Abigail_expressive_2024112501" {
		t.Fatalf("body avatar_id = %v", body["avatar_id"])
	}
	if body["voice_id"] != "73c0b6a2" {
		t.Fatalf("body voice_id = %v", body["voice_id"])
	}
	if body["text"] != "Hello!" {
		t.Fatalf("body text = %v", body["text"])
	}
	if body["max_duration_seconds"] != float64(300) {
		t.Fatalf("body max_duration_seconds = %v", body["max_duration_seconds"])
	}
	if _, present := body["audio"]; present {
		t.Fatal("audio must be omitted for text_stream sessions")
	}

	if resp.StreamID != "rt_9f2c1a" || resp.Status != "pending" {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.PrepaidSeconds != 300 {
		t.Fatalf("PrepaidSeconds = %d", resp.PrepaidSeconds)
	}
	if resp.CostTicks != 345000000000 {
		t.Fatalf("CostTicks = %d", resp.CostTicks)
	}
	if resp.BalanceAfter != 655000000000 {
		t.Fatalf("BalanceAfter (from X-QAI-Balance-After) = %d", resp.BalanceAfter)
	}
	if resp.RequestID != "req_abc123" {
		t.Fatalf("RequestID = %q", resp.RequestID)
	}
}

// TestCreateAvatarRealtimeSessionCostFromHeader pins the meta fallback:
// when the body carries no cost_ticks, the X-QAI-Cost-Ticks header fills it.
func TestCreateAvatarRealtimeSessionCostFromHeader(t *testing.T) {
	client := newTestClient(t, http.StatusOK, map[string]string{
		"X-QAI-Cost-Ticks":    "1000",
		"X-QAI-Balance-After": "2000",
		"X-QAI-Request-Id":    "req_hdr",
	}, `{"stream_id":"rt_1","status":"pending","prepaid_seconds":60}`, nil)

	resp, err := client.CreateAvatarRealtimeSession(t.Context(), &AvatarRealtimeRequest{
		SessionType: "tts", AvatarID: "av_1", VoiceID: "v_1", Text: "hi", MaxDurationSeconds: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CostTicks != 1000 {
		t.Fatalf("CostTicks = %d, want header fallback 1000", resp.CostTicks)
	}
	if resp.BalanceAfter != 2000 {
		t.Fatalf("BalanceAfter = %d, want header fallback 2000", resp.BalanceAfter)
	}
	if resp.RequestID != "req_hdr" {
		t.Fatalf("RequestID = %q, want header fallback req_hdr", resp.RequestID)
	}
}

// TestAvatarRealtimeRequestAudioShape pins the audio-session wire shape:
// audio union carried under "type", voice_id/text omitted entirely.
func TestAvatarRealtimeRequestAudioShape(t *testing.T) {
	data, err := json.Marshal(&AvatarRealtimeRequest{
		SessionType: "audio",
		AvatarID:    "av_1",
		Audio: &AvatarAudioInput{
			InputType: "base64",
			MediaType: "audio/mpeg",
			Data:      "AQID",
		},
		MaxDurationSeconds: 120,
	})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	if _, present := body["voice_id"]; present {
		t.Fatal("voice_id must be omitted for audio sessions")
	}
	if _, present := body["text"]; present {
		t.Fatal("text must be omitted for audio sessions")
	}
	audio, ok := body["audio"].(map[string]any)
	if !ok {
		t.Fatalf("audio = %v", body["audio"])
	}
	if audio["type"] != "base64" || audio["media_type"] != "audio/mpeg" || audio["data"] != "AQID" {
		t.Fatalf("audio union = %v", audio)
	}
	if _, present := audio["url"]; present {
		t.Fatal("audio.url must be omitted when unset")
	}
	if _, present := audio["asset_id"]; present {
		t.Fatal("audio.asset_id must be omitted when unset")
	}
}

func TestCreateAvatarRealtimeSessionInsufficientBalance(t *testing.T) {
	client := newTestClient(t, http.StatusPaymentRequired, nil,
		`{"error":{"message":"out of credits — top up to continue","type":"insufficient_balance","code":"INSUFFICIENT_BALANCE"}}`, nil)

	_, err := client.CreateAvatarRealtimeSession(t.Context(), &AvatarRealtimeRequest{
		SessionType: "tts", AvatarID: "av_1", VoiceID: "v_1", Text: "hi", MaxDurationSeconds: 60,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsInsufficientBalance(err) {
		t.Fatalf("IsInsufficientBalance(%v) = false", err)
	}
}

// --- 2. GET /qai/v1/avatar/realtime/{id} ---------------------------------

func TestGetAvatarRealtimeSession(t *testing.T) {
	cases := []struct {
		name string
		body string
		want AvatarRealtimeStatusResponse
	}{
		{
			name: "streaming",
			body: `{"stream_id":"rt_1","status":"streaming","hls_url":"https://cdn.heygen.com/rt_1/index.m3u8","request_id":"req_1"}`,
			want: AvatarRealtimeStatusResponse{StreamID: "rt_1", Status: "streaming", HLSURL: "https://cdn.heygen.com/rt_1/index.m3u8", RequestID: "req_1"},
		},
		{
			name: "completed with end_reason",
			body: `{"stream_id":"rt_1","status":"completed","end_reason":"idle_timeout","request_id":"req_2"}`,
			want: AvatarRealtimeStatusResponse{StreamID: "rt_1", Status: "completed", EndReason: "idle_timeout", RequestID: "req_2"},
		},
		{
			name: "error with message",
			body: `{"stream_id":"rt_1","status":"error","error_message":"avatar failed","request_id":"req_3"}`,
			want: AvatarRealtimeStatusResponse{StreamID: "rt_1", Status: "error", ErrorMessage: "avatar failed", RequestID: "req_3"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cap capturedRequest
			client := newTestClient(t, http.StatusOK, nil, tc.body, &cap)
			resp, err := client.GetAvatarRealtimeSession(t.Context(), "rt_1")
			if err != nil {
				t.Fatal(err)
			}
			if cap.Method != "GET" || cap.Path != "/qai/v1/avatar/realtime/rt_1" {
				t.Fatalf("request = %s %s", cap.Method, cap.Path)
			}
			if cap.Auth != "Bearer qai_test_key" {
				t.Fatalf("Authorization = %q", cap.Auth)
			}
			if *resp != tc.want {
				t.Fatalf("resp = %+v, want %+v", *resp, tc.want)
			}
		})
	}
}

func TestGetAvatarRealtimeSessionNotFound(t *testing.T) {
	client := newTestClient(t, http.StatusNotFound, nil,
		`{"error":{"message":"session rt_x not found","type":"not_found","code":"not_found"}}`, nil)
	_, err := client.GetAvatarRealtimeSession(t.Context(), "rt_x")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T (%v), want *APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "not_found" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

// --- 3. POST /qai/v1/avatar/realtime/{id}/text ---------------------------

func TestSendAvatarRealtimeText(t *testing.T) {
	cases := []struct {
		name      string
		req       AvatarRealtimeTextRequest
		wantDelta any  // nil = must be absent
		wantFinal bool
	}{
		{"delta append", AvatarRealtimeTextRequest{Delta: " more"}, " more", false},
		{"final with delta", AvatarRealtimeTextRequest{Delta: " end.", IsFinal: true}, " end.", true},
		{"final marker only", AvatarRealtimeTextRequest{IsFinal: true}, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cap capturedRequest
			client := newTestClient(t, http.StatusOK, nil,
				`{"ok":true,"buffered_bytes":512,"final":true,"request_id":"req_t"}`, &cap)
			resp, err := client.SendAvatarRealtimeText(t.Context(), "rt_1", &tc.req)
			if err != nil {
				t.Fatal(err)
			}
			if cap.Method != "POST" || cap.Path != "/qai/v1/avatar/realtime/rt_1/text" {
				t.Fatalf("request = %s %s", cap.Method, cap.Path)
			}
			body := cap.bodyJSON(t)
			delta, present := body["delta"]
			if tc.wantDelta == nil {
				if present {
					t.Fatalf("empty delta must be omitted, got %v", delta)
				}
			} else if delta != tc.wantDelta {
				t.Fatalf("body delta = %v, want %v", delta, tc.wantDelta)
			}
			// "final" is always on the wire (no omitempty), matching the Rust SDK.
			if body["final"] != tc.wantFinal {
				t.Fatalf("body final = %v, want %v", body["final"], tc.wantFinal)
			}
			if !resp.OK || resp.BufferedBytes != 512 || !resp.IsFinal || resp.RequestID != "req_t" {
				t.Fatalf("resp = %+v", resp)
			}
		})
	}
}

func TestSendAvatarRealtimeTextClosedStream(t *testing.T) {
	// Appending after close: upstream 410 passed through as provider_error.
	client := newTestClient(t, http.StatusGone, nil,
		`{"error":{"message":"text stream already closed","type":"provider_error","code":"provider_error"}}`, nil)
	_, err := client.SendAvatarRealtimeText(t.Context(), "rt_1", &AvatarRealtimeTextRequest{Delta: "late"})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T (%v), want *APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusGone || apiErr.Code != "provider_error" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

// --- 4. POST /qai/v1/avatar/realtime/{id}/cancel -------------------------

func TestCancelAvatarRealtimeSession(t *testing.T) {
	cases := []struct {
		name          string
		body          string
		wantCancelled bool
	}{
		{"initiated", `{"stream_id":"rt_1","cancelled":true,"request_id":"req_c"}`, true},
		{"already terminal", `{"stream_id":"rt_1","cancelled":false,"request_id":"req_c"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cap capturedRequest
			client := newTestClient(t, http.StatusOK, nil, tc.body, &cap)
			resp, err := client.CancelAvatarRealtimeSession(t.Context(), "rt_1")
			if err != nil {
				t.Fatal(err)
			}
			if cap.Method != "POST" || cap.Path != "/qai/v1/avatar/realtime/rt_1/cancel" {
				t.Fatalf("request = %s %s", cap.Method, cap.Path)
			}
			if len(cap.Body) != 0 {
				t.Fatalf("cancel must send no body, got %s", cap.Body)
			}
			if resp.StreamID != "rt_1" || resp.Cancelled != tc.wantCancelled {
				t.Fatalf("resp = %+v", resp)
			}
		})
	}
}

// --- 5. GET /qai/v1/audio/sounds -----------------------------------------

func TestSearchAudioSounds(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{
		"sounds":[{"id":"trk_1","name":"Uplifting Corporate","description":"Bright track","audio_url":"https://resource.heygen.ai/sounds/trk_1.wav","duration":94.5,"score":0.91,"type":"music"}],
		"has_more":true,"next_token":"tok_next","request_id":"req_s"}`, &cap)

	minScore := 0.7
	resp, err := client.SearchAudioSounds(t.Context(), &AudioSoundsQuery{
		Query:     "calm piano",
		SoundType: "music",
		Limit:     10,
		MinScore:  &minScore,
		Token:     "tok_prev",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "GET" || cap.Path != "/qai/v1/audio/sounds" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	if cap.Auth != "Bearer qai_test_key" {
		t.Fatalf("Authorization = %q", cap.Auth)
	}
	wantQuery := map[string]string{
		"query": "calm piano", "type": "music", "limit": "10", "min_score": "0.7", "token": "tok_prev",
	}
	for k, want := range wantQuery {
		if got := cap.Query[k]; len(got) != 1 || got[0] != want {
			t.Fatalf("query %s = %v, want %q", k, got, want)
		}
	}

	if len(resp.Sounds) != 1 {
		t.Fatalf("Sounds len = %d", len(resp.Sounds))
	}
	s := resp.Sounds[0]
	if s.ID != "trk_1" || s.Name != "Uplifting Corporate" || s.Description != "Bright track" {
		t.Fatalf("sound = %+v", s)
	}
	if s.AudioURL != "https://resource.heygen.ai/sounds/trk_1.wav" || s.Duration != 94.5 || s.Score != 0.91 || s.SoundType != "music" {
		t.Fatalf("sound = %+v", s)
	}
	if !resp.HasMore || resp.NextToken != "tok_next" || resp.RequestID != "req_s" {
		t.Fatalf("page fields = %+v", resp)
	}
}

func TestSearchAudioSoundsMinimalQueryAndEmptyPage(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil,
		`{"sounds":[],"has_more":false,"next_token":"","request_id":"req_e"}`, &cap)

	resp, err := client.SearchAudioSounds(t.Context(), &AudioSoundsQuery{Query: "whale noises"})
	if err != nil {
		t.Fatal(err)
	}
	if got := cap.Query["query"]; len(got) != 1 || got[0] != "whale noises" {
		t.Fatalf("query = %v", cap.Query)
	}
	for _, k := range []string{"type", "limit", "min_score", "token"} {
		if _, present := cap.Query[k]; present {
			t.Fatalf("unset param %s must be omitted, query = %v", k, cap.Query)
		}
	}
	if resp.Sounds == nil || len(resp.Sounds) != 0 {
		t.Fatalf("Sounds = %#v, want empty non-nil page decode", resp.Sounds)
	}
	if resp.HasMore || resp.NextToken != "" {
		t.Fatalf("page fields = %+v", resp)
	}
}

// --- 6. GET /qai/v1/video/template/{id} ----------------------------------

func TestVideoTemplateDetail(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{
		"template":{
			"id":"tmpl_5f0a","name":"Product Launch","aspect_ratio":"16:9",
			"variables":{
				"headline":{"type":"text","content":"Default headline"},
				"presenter":{"type":"character","character_id":"Abigail","character_type":"avatar"}
			},
			"scenes":[{"scene_id":"scene_1","script":"Introducing {{headline}}...","variables":[{"name":"headline","variable_type":"text"}]}]
		},
		"request_id":"req_d"}`, &cap)

	resp, err := client.VideoTemplateDetail(t.Context(), "tmpl_5f0a")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" || cap.Path != "/qai/v1/video/template/tmpl_5f0a" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}

	tmpl := resp.Template
	if tmpl.ID != "tmpl_5f0a" || tmpl.Name != "Product Launch" || tmpl.AspectRatio != "16:9" {
		t.Fatalf("template = %+v", tmpl)
	}
	headline, ok := tmpl.Variables["headline"].(map[string]any)
	if !ok || headline["type"] != "text" || headline["content"] != "Default headline" {
		t.Fatalf("headline variable = %v", tmpl.Variables["headline"])
	}
	presenter, ok := tmpl.Variables["presenter"].(map[string]any)
	if !ok || presenter["type"] != "character" || presenter["character_id"] != "Abigail" {
		t.Fatalf("presenter variable = %v", tmpl.Variables["presenter"])
	}
	if len(tmpl.Scenes) != 1 {
		t.Fatalf("scenes = %+v", tmpl.Scenes)
	}
	scene := tmpl.Scenes[0]
	if scene.SceneID != "scene_1" || scene.Script != "Introducing {{headline}}..." {
		t.Fatalf("scene = %+v", scene)
	}
	if len(scene.Variables) != 1 || scene.Variables[0].Name != "headline" || scene.Variables[0].VariableType != "text" {
		t.Fatalf("scene variables = %+v", scene.Variables)
	}
	if resp.RequestID != "req_d" {
		t.Fatalf("RequestID = %q", resp.RequestID)
	}
}

func TestVideoTemplateDetailUnknownTemplate(t *testing.T) {
	// Unknown template id: upstream 4xx passed through as provider_error.
	client := newTestClient(t, http.StatusNotFound, nil,
		`{"error":{"message":"template not found","type":"provider_error","code":"provider_error"}}`, nil)
	_, err := client.VideoTemplateDetail(t.Context(), "tmpl_nope")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T (%v), want *APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "provider_error" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

// --- 7. POST /qai/v1/video/template/{id} (async job) ---------------------

func TestVideoTemplateGenerate(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusAccepted, nil,
		`{"job_id":"qai_job_3def45c00112","status":"pending","type":"video/template-v3","request_id":"req_g"}`, &cap)

	fpsCaption := true
	reorder := false
	resp, err := client.VideoTemplateGenerate(t.Context(), "tmpl_5f0a", &VideoTemplateGenerateRequest{
		Variables: map[string]any{
			"headline": map[string]any{"type": "text", "content": "New headline"},
		},
		Title:        "Launch video",
		SceneIDs:     []string{"scene_1", "scene_1"},
		Dimension:    &VideoTemplateDimension{Width: 1920, Height: 1080},
		FPS:          30,
		Caption:      &fpsCaption,
		Subtitles:    &VideoTemplateSubtitles{PresetName: "classic"},
		ReorderMusic: &reorder,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/video/template/tmpl_5f0a" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	body := cap.bodyJSON(t)
	vars, ok := body["variables"].(map[string]any)
	if !ok {
		t.Fatalf("variables = %v", body["variables"])
	}
	headline, ok := vars["headline"].(map[string]any)
	if !ok || headline["type"] != "text" || headline["content"] != "New headline" {
		t.Fatalf("headline = %v", vars["headline"])
	}
	if body["title"] != "Launch video" {
		t.Fatalf("title = %v", body["title"])
	}
	sceneIDs, ok := body["scene_ids"].([]any)
	if !ok || len(sceneIDs) != 2 || sceneIDs[0] != "scene_1" {
		t.Fatalf("scene_ids = %v", body["scene_ids"])
	}
	dim, ok := body["dimension"].(map[string]any)
	if !ok || dim["width"] != float64(1920) || dim["height"] != float64(1080) {
		t.Fatalf("dimension = %v", body["dimension"])
	}
	if body["fps"] != float64(30) {
		t.Fatalf("fps = %v", body["fps"])
	}
	if body["caption"] != true {
		t.Fatalf("caption = %v", body["caption"])
	}
	subs, ok := body["subtitles"].(map[string]any)
	if !ok || subs["preset_name"] != "classic" {
		t.Fatalf("subtitles = %v", body["subtitles"])
	}
	if body["reorder_music"] != false {
		t.Fatalf("reorder_music = %v (a false pointer must survive to the wire)", body["reorder_music"])
	}
	// Optional fields left unset must be omitted entirely.
	for _, k := range []string{"keep_text_vertically_centered", "include_gif", "enable_sharing", "folder_id", "brand_voice_id"} {
		if _, present := body[k]; present {
			t.Fatalf("unset field %s must be omitted", k)
		}
	}

	if resp.JobID != "qai_job_3def45c00112" || resp.Status != "pending" || resp.Type != "video/template-v3" || resp.RequestID != "req_g" {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestVideoTemplateGenerateJobResultDecode(t *testing.T) {
	// The completed job's result carries the template render payload.
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{
		"job_id":"qai_job_1","type":"video/template-v3","status":"completed",
		"result":{"video_id":"vid_77aa01","video_url":"https://resource.heygen.ai/video/vid_77aa01.mp4","thumbnail_url":"https://resource.heygen.ai/thumb/vid_77aa01.jpg","duration_seconds":42.7,"model":"heygen-template","cost_ticks":22500000000,"request_id":"req_g"},
		"cost_ticks":22500000000}`, &cap)

	job, err := client.GetJob(t.Context(), "qai_job_1")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" || cap.Path != "/qai/v1/jobs/qai_job_1" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	if job.Status != "completed" || job.CostTicks != 22500000000 {
		t.Fatalf("job = %+v", job)
	}
	var result map[string]any
	if err := json.Unmarshal(job.Result, &result); err != nil {
		t.Fatal(err)
	}
	if result["video_url"] != "https://resource.heygen.ai/video/vid_77aa01.mp4" || result["model"] != "heygen-template" || result["duration_seconds"] != 42.7 {
		t.Fatalf("result = %v", result)
	}
}

// --- 8. POST /qai/v1/video/batch -----------------------------------------

func TestVideoBatchSubmit(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusAccepted, nil,
		`{"batch_id":"batch_66aa1c","status":"processing","total_items":2,"request_id":"req_b"}`, &cap)

	resp, err := client.VideoBatchSubmit(t.Context(), &VideoBatchSubmitRequest{
		Title: "Onboarding videos",
		Videos: []map[string]any{
			{"type": "avatar", "avatar_id": "Abigail", "voice_id": "v_1", "script": "Welcome!"},
			{"type": "cinematic_avatar", "avatar_id": []any{"look_1", "look_2"}, "script": "Billing 101"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/video/batch" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	body := cap.bodyJSON(t)
	if body["title"] != "Onboarding videos" {
		t.Fatalf("title = %v", body["title"])
	}
	videos, ok := body["videos"].([]any)
	if !ok || len(videos) != 2 {
		t.Fatalf("videos = %v", body["videos"])
	}
	first, ok := videos[0].(map[string]any)
	if !ok || first["type"] != "avatar" || first["avatar_id"] != "Abigail" || first["script"] != "Welcome!" {
		t.Fatalf("videos[0] = %v", videos[0])
	}
	// Polymorphic passthrough: cinematic_avatar's avatar_id is an array.
	second, ok := videos[1].(map[string]any)
	if !ok || second["type"] != "cinematic_avatar" {
		t.Fatalf("videos[1] = %v", videos[1])
	}
	if looks, ok := second["avatar_id"].([]any); !ok || len(looks) != 2 {
		t.Fatalf("videos[1].avatar_id = %v (array must pass through verbatim)", second["avatar_id"])
	}

	if resp.BatchID != "batch_66aa1c" || resp.Status != "processing" || resp.TotalItems != 2 || resp.RequestID != "req_b" {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestVideoBatchSubmitOmitsEmptyTitle(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusAccepted, nil,
		`{"batch_id":"batch_1","status":"processing","total_items":1,"request_id":"req_b"}`, &cap)
	_, err := client.VideoBatchSubmit(t.Context(), &VideoBatchSubmitRequest{
		Videos: []map[string]any{{"type": "avatar", "avatar_id": "a", "voice_id": "v", "script": "s"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, present := cap.bodyJSON(t)["title"]; present {
		t.Fatal("empty title must be omitted")
	}
}

// --- 9. GET /qai/v1/video/batch/{id} -------------------------------------

func TestVideoBatchStatus(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{
		"batch_id":"batch_66aa1c","title":"Onboarding videos","status":"completed","total_items":3,
		"counts_by_status":{"completed":2,"failed":1},
		"created_at":1752741600,
		"items":[
			{"item_index":0,"status":"completed","video_id":"vid_001","video_url":"https://resource.heygen.ai/video/vid_001.mp4"},
			{"item_index":1,"status":"completed","video_id":"vid_002","video_url":"https://resource.heygen.ai/video/vid_002.mp4"},
			{"item_index":2,"status":"failed","error":{"code":"avatar_not_found","message":"avatar id not found"}}
		],
		"has_more":false,"next_token":"","billing_status":"settled","cost_ticks":46000000000,"request_id":"req_bs"}`, &cap)

	resp, err := client.VideoBatchStatus(t.Context(), "batch_66aa1c", &VideoBatchStatusQuery{Limit: 100, Token: "tok_page2"})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "GET" || cap.Path != "/qai/v1/video/batch/batch_66aa1c" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	if got := cap.Query["limit"]; len(got) != 1 || got[0] != "100" {
		t.Fatalf("limit = %v", cap.Query)
	}
	if got := cap.Query["token"]; len(got) != 1 || got[0] != "tok_page2" {
		t.Fatalf("token = %v", cap.Query)
	}

	if resp.BatchID != "batch_66aa1c" || resp.Title != "Onboarding videos" || resp.Status != "completed" || resp.TotalItems != 3 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.CountsByStatus["completed"] != 2 || resp.CountsByStatus["failed"] != 1 {
		t.Fatalf("CountsByStatus = %v", resp.CountsByStatus)
	}
	if resp.CreatedAt != 1752741600 {
		t.Fatalf("CreatedAt = %d (unix seconds)", resp.CreatedAt)
	}
	if len(resp.Items) != 3 {
		t.Fatalf("Items len = %d", len(resp.Items))
	}
	if resp.Items[0].ItemIndex != 0 || resp.Items[0].Status != "completed" || resp.Items[0].VideoID != "vid_001" || resp.Items[0].VideoURL != "https://resource.heygen.ai/video/vid_001.mp4" {
		t.Fatalf("Items[0] = %+v", resp.Items[0])
	}
	failed := resp.Items[2]
	if failed.Status != "failed" || failed.VideoID != "" || failed.VideoURL != "" {
		t.Fatalf("Items[2] = %+v", failed)
	}
	if failed.Error == nil || failed.Error.Code != "avatar_not_found" || failed.Error.Message != "avatar id not found" {
		t.Fatalf("Items[2].Error = %+v", failed.Error)
	}
	if resp.HasMore || resp.NextToken != "" {
		t.Fatalf("page fields = %+v", resp)
	}
	if resp.BillingStatus != "settled" || resp.CostTicks != 46000000000 || resp.RequestID != "req_bs" {
		t.Fatalf("billing fields = %+v", resp)
	}
}

func TestVideoBatchStatusUnsettledWithholdsURLs(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{
		"batch_id":"batch_1","title":"","status":"completed","total_items":1,
		"counts_by_status":{"completed":1},"created_at":1752741600,
		"items":[{"item_index":0,"status":"completed","video_id":"vid_001"}],
		"has_more":false,"next_token":"","billing_status":"settlement_pending","cost_ticks":0,"request_id":"req_p"}`, &cap)

	resp, err := client.VideoBatchStatus(t.Context(), "batch_1", nil)
	if err != nil {
		t.Fatal(err)
	}
	// nil query: no query string at all.
	if len(cap.Query) != 0 {
		t.Fatalf("query = %v, want none", cap.Query)
	}
	if resp.BillingStatus != "settlement_pending" || resp.CostTicks != 0 {
		t.Fatalf("billing fields = %+v", resp)
	}
	if resp.Items[0].VideoURL != "" {
		t.Fatalf("VideoURL = %q, must be withheld until settled", resp.Items[0].VideoURL)
	}
}

func TestVideoBatchStatusNotFound(t *testing.T) {
	client := newTestClient(t, http.StatusNotFound, nil,
		`{"error":{"message":"batch batch_x not found","type":"not_found","code":"not_found"}}`, nil)
	_, err := client.VideoBatchStatus(t.Context(), "batch_x", nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T (%v), want *APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "not_found" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}
