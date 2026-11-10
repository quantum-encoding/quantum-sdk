package qai

// Wire contract for POST /qai/v1/video/digital-twin, pinned to the gateway's
// handleDigitalTwinCreate: it creates a twin from training footage, taking
// JSON {name, video_url, avatar_group_id?} and answering with a synchronous
// record rather than a job envelope.

import (
	"net/http"
	"strings"
	"testing"
)

func TestVideoDigitalTwinSendsCreateBody(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, map[string]string{"X-QAI-Request-Id": "req_hdr"}, `{
		"name":"Rich","consent_url":"https://app.heygen.com/consent/1","model":"heygen-digital-twin",
		"request_id":"req_t","group_id":"grp_1","status":"pending","consent_status":"not_started",
		"avatar_id":"look_1"
	}`, &cap)

	resp, err := client.VideoDigitalTwin(t.Context(), &DigitalTwinCreateRequest{
		Name:          "Rich",
		VideoURL:      "https://example.com/train.mp4",
		AvatarGroupID: "grp_existing",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/video/digital-twin" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	if !strings.HasPrefix(cap.ContentType, "application/json") {
		t.Fatalf("content type = %q, want application/json", cap.ContentType)
	}
	body := cap.bodyJSON(t)
	if got := strings.Join(keys(body), ","); got != "avatar_group_id,name,video_url" {
		t.Fatalf("body keys = %s, want avatar_group_id,name,video_url", got)
	}
	if body["name"] != "Rich" || body["video_url"] != "https://example.com/train.mp4" || body["avatar_group_id"] != "grp_existing" {
		t.Fatalf("body = %v", body)
	}

	want := DigitalTwinCreateResponse{
		Name: "Rich", ConsentURL: "https://app.heygen.com/consent/1", Model: "heygen-digital-twin",
		GroupID: "grp_1", Status: "pending", ConsentStatus: "not_started", AvatarID: "look_1",
		RequestID: "req_t",
	}
	if *resp != want {
		t.Fatalf("resp = %+v, want %+v", *resp, want)
	}
}

// The gateway leaves out group_id, status, consent_status and avatar_id when
// HeyGen returned no group or look; request_id then comes from the header.
func TestVideoDigitalTwinWithoutGroupOrLook(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, map[string]string{"X-QAI-Request-Id": "req_hdr"},
		`{"name":"Rich","consent_url":"","model":"heygen-digital-twin"}`, &cap)

	resp, err := client.VideoDigitalTwin(t.Context(), &DigitalTwinCreateRequest{
		Name:     "Rich",
		VideoURL: "https://example.com/train.mp4",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(keys(cap.bodyJSON(t)), ","); got != "name,video_url" {
		t.Fatalf("body keys = %s, want name,video_url", got)
	}
	if resp.GroupID != "" || resp.AvatarID != "" || resp.RequestID != "req_hdr" {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestVideoDigitalTwinSurfacesGateway400(t *testing.T) {
	client := newTestClient(t, http.StatusBadRequest, nil,
		`{"error":{"message":"video_url is required (or send multipart/form-data with a \"training\" file)","type":"invalid_request","code":"invalid_request"}}`, nil)

	_, err := client.VideoDigitalTwin(t.Context(), &DigitalTwinCreateRequest{Name: "Rich"})
	if err == nil || !strings.Contains(err.Error(), "video_url is required") {
		t.Fatalf("err = %v, want the gateway's video_url message", err)
	}
}
