package qai

import (
	"context"
	"encoding/json"
	"time"
)

// JobListEntry is a single job in the job list response.
type JobListEntry struct {
	// JobID is the unique job identifier.
	JobID string `json:"job_id"`

	// Type is the job type (e.g. "video/generate", "audio/tts").
	Type string `json:"type"`

	// Status is the job status ("pending", "processing", "completed", "failed").
	Status string `json:"status"`

	// Result contains the job output when completed.
	Result json.RawMessage `json:"result,omitempty"`

	// Error contains the error message if the job failed.
	Error string `json:"error,omitempty"`

	// CostTicks is the total cost in ticks.
	CostTicks int64 `json:"cost_ticks,omitempty"`

	// CreatedAt is the job creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is when processing began.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when the job finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// RequestID is the originating request identifier.
	RequestID string `json:"request_id"`
}

// JobListResponse is the response from listing jobs.
type JobListResponse struct {
	// Jobs is the list of jobs.
	Jobs []JobListEntry `json:"jobs"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id"`
}

// ListJobs returns all jobs for the authenticated user.
func (c *Client) ListJobs(ctx context.Context) (*JobListResponse, error) {
	var resp JobListResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/jobs", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
