package qai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// JobCreateRequest is the request body for creating an async job.
type JobCreateRequest struct {
	// Type is the job type (e.g. "video/generate", "audio/music").
	Type string `json:"type"`

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
