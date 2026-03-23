package qai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// JobCreateRequest is the request body for creating an async job.
type JobCreateRequest struct {
	// Type is the job type (e.g. "video/generate", "audio/music").
	Type string `json:"type"`

	// JobType is an alias for Type (sdk-graph canonical name).
	JobType string `json:"job_type,omitempty"`

	// Params contains model-specific job parameters.
	Params json.RawMessage `json:"params"`
}

// JobCreateResponse is the response from job creation.
type JobCreateResponse struct {
	// JobID is the unique job identifier for polling.
	JobID string `json:"job_id"`

	// Status is the initial job status (e.g. "pending").
	Status string `json:"status"`
}

// JobStatusResponse is the response from a job status check.
type JobStatusResponse struct {
	// JobID is the unique job identifier.
	JobID string `json:"job_id"`

	// Status is the current job status ("pending", "processing", "completed", "failed").
	Status string `json:"status"`

	// Result contains the job output when completed (shape depends on job type).
	Result json.RawMessage `json:"result,omitempty"`

	// Error contains the error message if the job failed.
	Error string `json:"error,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`
}

// JobStreamEvent is a single event from an SSE job stream.
type JobStreamEvent struct {
	// Type is the event type ("progress", "complete", "error").
	Type string `json:"type"`

	// JobID is the job identifier.
	JobID string `json:"job_id,omitempty"`

	// Status is the current job status.
	Status string `json:"status,omitempty"`

	// Result is the job output (present on "complete" events).
	Result json.RawMessage `json:"result,omitempty"`

	// Error is the error message (present on "error" events).
	Error string `json:"error,omitempty"`

	// CostTicks is the total cost in ticks (present on "complete" events).
	CostTicks int64 `json:"cost_ticks,omitempty"`

	// CompletedAt is the completion timestamp.
	CompletedAt string `json:"completed_at,omitempty"`
}

// CreateJob submits an async job and returns the job ID for polling.
func (c *Client) CreateJob(ctx context.Context, req *JobCreateRequest) (*JobCreateResponse, error) {
	var resp JobCreateResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/jobs", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetJob checks the status of an async job.
func (c *Client) GetJob(ctx context.Context, jobID string) (*JobStatusResponse, error) {
	var resp JobStatusResponse
	_, err := c.doJSON(ctx, "GET", fmt.Sprintf("/qai/v1/jobs/%s", jobID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// PollJob polls a job until it reaches a terminal state ("completed" or "failed")
// or the maximum number of attempts is reached.
//
// It sleeps for the given interval between each poll. If maxAttempts is reached
// without a terminal state, a synthetic "timeout" response is returned.
func (c *Client) PollJob(ctx context.Context, jobID string, interval time.Duration, maxAttempts int) (*JobStatusResponse, error) {
	for i := 0; i < maxAttempts; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		status, err := c.GetJob(ctx, jobID)
		if err != nil {
			return nil, err
		}

		switch status.Status {
		case "completed", "failed":
			return status, nil
		}
	}

	return &JobStatusResponse{
		JobID:  jobID,
		Status: "timeout",
		Error:  fmt.Sprintf("Job polling timed out after %d attempts", maxAttempts),
	}, nil
}

// Generate3D submits a 3D model generation job via the async jobs system.
// Returns the job creation response -- use PollJob to wait for completion.
func (c *Client) Generate3D(ctx context.Context, model string, prompt string, imageURL string) (*JobCreateResponse, error) {
	params := map[string]any{"model": model}
	if prompt != "" {
		params["prompt"] = prompt
	}
	if imageURL != "" {
		params["image_url"] = imageURL
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("qai: marshal 3d params: %w", err)
	}
	return c.CreateJob(ctx, &JobCreateRequest{
		Type:   "3d/generate",
		Params: paramsJSON,
	})
}

// ChatJob submits a chat completion as an async job.
//
// Useful for long-running models (e.g. Opus) where synchronous /qai/v1/chat
// may time out. Params are the same shape as ChatRequest.
// Use StreamJob() or PollJob() to get the result.
func (c *Client) ChatJob(ctx context.Context, req *ChatRequest) (*JobCreateResponse, error) {
	paramsJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("qai: marshal chat request: %w", err)
	}
	return c.CreateJob(ctx, &JobCreateRequest{
		Type:   "chat",
		Params: paramsJSON,
	})
}

// StreamJob opens an SSE stream for a job, returning a channel of events.
//
// Events: "progress" (status update), "complete" (with result), "error".
// The channel is closed when the stream ends or the context is cancelled.
func (c *Client) StreamJob(ctx context.Context, jobID string) (<-chan JobStreamEvent, error) {
	path := fmt.Sprintf("/qai/v1/jobs/%s/stream", jobID)
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("qai: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	// Use a client without timeout for streaming — context controls cancellation.
	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qai: GET %s: %w", path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, parseAPIError(resp, resp.Header.Get("X-QAI-Request-Id"))
	}

	ch := make(chan JobStreamEvent, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")

			if payload == "[DONE]" {
				return
			}

			var event JobStreamEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				select {
				case ch <- JobStreamEvent{Type: "error", Error: fmt.Sprintf("parse SSE: %v", err)}:
				case <-ctx.Done():
				}
				return
			}

			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}

			if event.Type == "complete" || event.Type == "error" {
				return
			}
		}
	}()

	return ch, nil
}

// RemeshRequest describes a 3D remesh operation.
type RemeshRequest struct {
	InputTaskID       string   `json:"input_task_id,omitempty"`
	ModelURL          string   `json:"model_url,omitempty"`
	TargetFormats     []string `json:"target_formats,omitempty"`
	Topology          string   `json:"topology,omitempty"`
	TargetPolycount   int      `json:"target_polycount,omitempty"`
	ResizeHeight      float64  `json:"resize_height,omitempty"`
	OriginAt          string   `json:"origin_at,omitempty"`
	ConvertFormatOnly bool     `json:"convert_format_only,omitempty"`
}

// RigRequest describes a 3D rigging operation for a humanoid model.
type RigRequest struct {
	InputTaskID     string  `json:"input_task_id,omitempty"`
	ModelURL        string  `json:"model_url,omitempty"`
	HeightMeters    float64 `json:"height_meters,omitempty"`
	TextureImageURL string  `json:"texture_image_url,omitempty"`
}

// AnimateRequest describes an animation operation on a rigged character.
type AnimateRequest struct {
	RigTaskID   string       `json:"rig_task_id"`
	ActionID    int          `json:"action_id"`
	PostProcess *PostProcess `json:"post_process,omitempty"`
}

// PostProcess describes post-processing options for animation.
type PostProcess struct {
	OperationType string `json:"operation_type"`
	FPS           int    `json:"fps,omitempty"`
}

// Remesh submits a 3D remesh job and polls until completion.
func (c *Client) Remesh(ctx context.Context, req *RemeshRequest) (*JobStatusResponse, error) {
	params, _ := json.Marshal(req)
	job, err := c.CreateJob(ctx, &JobCreateRequest{Type: "3d/remesh", Params: params})
	if err != nil {
		return nil, err
	}
	return c.PollJob(ctx, job.JobID, 5*time.Second, 120)
}

// Rig submits a 3D rigging job and polls until completion.
func (c *Client) Rig(ctx context.Context, req *RigRequest) (*JobStatusResponse, error) {
	params, _ := json.Marshal(req)
	job, err := c.CreateJob(ctx, &JobCreateRequest{Type: "3d/rig", Params: params})
	if err != nil {
		return nil, err
	}
	return c.PollJob(ctx, job.JobID, 5*time.Second, 120)
}

// Animate submits a 3D animation job and polls until completion.
func (c *Client) Animate(ctx context.Context, req *AnimateRequest) (*JobStatusResponse, error) {
	params, _ := json.Marshal(req)
	job, err := c.CreateJob(ctx, &JobCreateRequest{Type: "3d/animate", Params: params})
	if err != nil {
		return nil, err
	}
	return c.PollJob(ctx, job.JobID, 5*time.Second, 120)
}

// JobResponse is the response from async video job submission (sdk-graph canonical name).
type JobResponse struct {
	// JobID is the job identifier for polling status.
	JobID string `json:"job_id"`

	// Status is the current status.
	Status string `json:"status"`

	// CostTicks is the total cost in ticks (may be 0 until job completes).
	CostTicks int64 `json:"cost_ticks"`

	// Extra contains additional response fields.
	Extra map[string]any `json:"extra,omitempty"`
}

// JobSummary is a summary of a job in the list response (sdk-graph canonical name).
type JobSummary struct {
	// JobID is the unique job identifier.
	JobID string `json:"job_id"`

	// Status is the job status.
	Status string `json:"status"`

	// JobType is the job type.
	JobType string `json:"type,omitempty"`

	// CreatedAt is the job creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// CompletedAt is when the job finished.
	CompletedAt string `json:"completed_at,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`
}

// ListJobsResponse is the response from listing jobs (sdk-graph canonical name).
type ListJobsResponse struct {
	Jobs []JobSummary `json:"jobs"`
}

// ModelUrls contains URLs for each exported format in a remesh result.
type ModelUrls struct {
	GLB   string `json:"glb"`
	FBX   string `json:"fbx"`
	OBJ   string `json:"obj"`
	USDZ  string `json:"usdz"`
	STL   string `json:"stl"`
	Blend string `json:"blend"`
}

// AnimationPostProcess describes post-processing options for animation export.
type AnimationPostProcess struct {
	// OperationType is the operation: "change_fps", "fbx2usdz", "extract_armature".
	OperationType string `json:"operation_type"`

	// FPS is the target FPS (for "change_fps"): 24, 25, 30, 60.
	FPS int `json:"fps,omitempty"`
}

// RetextureRequest describes a retexture operation on a 3D model.
type RetextureRequest struct {
	InputTaskID      string   `json:"input_task_id,omitempty"`
	ModelURL         string   `json:"model_url,omitempty"`
	Prompt           string   `json:"prompt,omitempty"`
	TextStylePrompt  string   `json:"text_style_prompt,omitempty"`
	ImageStyleURL    string   `json:"image_style_url,omitempty"`
	AIModel          string   `json:"ai_model,omitempty"`
	EnableOriginalUV *bool    `json:"enable_original_uv,omitempty"`
	EnablePBR        *bool    `json:"enable_pbr,omitempty"`
	RemoveLighting   *bool    `json:"remove_lighting,omitempty"`
	TargetFormats    []string `json:"target_formats,omitempty"`
}

// Retexture submits a retexture job and polls until completion.
func (c *Client) Retexture(ctx context.Context, req *RetextureRequest) (*JobStatusResponse, error) {
	params, _ := json.Marshal(req)
	job, err := c.CreateJob(ctx, &JobCreateRequest{Type: "3d/retexture", Params: params})
	if err != nil {
		return nil, err
	}
	return c.PollJob(ctx, job.JobID, 5*time.Second, 120)
}
