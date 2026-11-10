package qai

// Wire contract for POST /qai/v1/video/photo-avatar. The gateway accepts only
// the talking-photo body (photo_base64 + script, optional voice_id and
// aspect_ratio) and answers anything else with a 400, so the request must
// carry exactly those keys.

import (
	"net/http"
	"sort"
	"strings"
	"testing"
)

func TestVideoPhotoAvatarSendsTalkingPhotoBody(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusAccepted, nil,
		`{"job_id":"qai_job_tp1","status":"pending","type":"video/talking-photo","request_id":"req_tp"}`, &cap)

	resp, err := client.VideoPhotoAvatar(t.Context(), &PhotoAvatarRequest{
		PhotoBase64: "aGVsbG8=",
		Script:      "Hello from a photo.",
		VoiceID:     "v_42",
		AspectRatio: "9:16",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/video/photo-avatar" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	body := cap.bodyJSON(t)
	want := map[string]string{
		"photo_base64": "aGVsbG8=",
		"script":       "Hello from a photo.",
		"voice_id":     "v_42",
		"aspect_ratio": "9:16",
	}
	if len(body) != len(want) {
		t.Fatalf("body keys = %v, want exactly %v", keys(body), want)
	}
	for k, v := range want {
		if body[k] != v {
			t.Fatalf("body[%q] = %v, want %q", k, body[k], v)
		}
	}

	if resp.JobID != "qai_job_tp1" || resp.Status != "pending" || resp.RequestID != "req_tp" {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestVideoPhotoAvatarOmitsUnsetOptionals(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusAccepted, nil,
		`{"job_id":"qai_job_tp2","status":"pending","type":"video/talking-photo","request_id":"req_tp"}`, &cap)

	_, err := client.VideoPhotoAvatar(t.Context(), &PhotoAvatarRequest{
		PhotoBase64: "aGVsbG8=",
		Script:      "Hi.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(keys(cap.bodyJSON(t)), ","); got != "photo_base64,script" {
		t.Fatalf("body keys = %s, want photo_base64,script", got)
	}
}

func TestVideoPhotoAvatarSurfacesGateway400(t *testing.T) {
	client := newTestClient(t, http.StatusBadRequest, nil,
		`{"error":{"message":"photo_base64 and script are required","type":"invalid_request","code":"invalid_request"}}`, nil)

	_, err := client.VideoPhotoAvatar(t.Context(), &PhotoAvatarRequest{PhotoBase64: "aGVsbG8="})
	if err == nil {
		t.Fatal("want an error for a 400 response")
	}
	if !strings.Contains(err.Error(), "photo_base64 and script are required") {
		t.Fatalf("err = %v, want the gateway's message", err)
	}
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
