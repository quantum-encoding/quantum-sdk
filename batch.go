package qai

import (
	"context"
	"encoding/json"
	"fmt"
)

// BatchJob is a single job in a batch submission.
type BatchJob struct {
	// Model to use for this job.
	Model string `json:"model"`

	// Prompt is the prompt text.
	Prompt string `json:"prompt"`

	// Title is an optional title for this job.
	Title string `json:"title,omitempty"`

	// SystemPrompt is an optional system prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`

	// MaxTokens is the optional maximum tokens to generate.
	MaxTokens int `json:"max_tokens,omitempty"`
}

// BatchSubmitResponse is the response from batch submission.
type BatchSubmitResponse struct {
	// JobIDs is the list of created job IDs.
	JobIDs []string `json:"job_ids"`

	// Status of the batch submission.
	Status string `json:"status"`
}

// BatchJsonlResponse is the response from JSONL batch submission.
type BatchJsonlResponse struct {
	// JobIDs is the list of created job IDs.
	JobIDs []string `json:"job_ids"`
}

// BatchJobInfo is a single job in the batch jobs list.
type BatchJobInfo struct {
	// JobID is the job identifier.
	JobID string `json:"job_id"`

	// Status is the current status (e.g. "pending", "running", "completed", "failed").
	Status string `json:"status"`

	// Model used for this job.
	Model string `json:"model,omitempty"`

	// Title is the job title.
	Title string `json:"title,omitempty"`

	// CreatedAt is when the job was created.
	CreatedAt string `json:"created_at,omitempty"`

	// CompletedAt is when the job completed.
	CompletedAt string `json:"completed_at,omitempty"`

	// Result is present when the job is completed.
	Result json.RawMessage `json:"result,omitempty"`

	// Error is present when the job failed.
	Error string `json:"error,omitempty"`

	// CostTicks is the cost in ticks.
	CostTicks int64 `json:"cost_ticks,omitempty"`
}

// BatchJobsResponse is the response from listing batch jobs.
type BatchJobsResponse struct {
	// Jobs is the list of batch jobs.
	Jobs []BatchJobInfo `json:"jobs"`
}

// BatchSubmit submits a batch of jobs for processing.
// Each job runs independently and can be polled via the Jobs API.
func (c *Client) BatchSubmit(ctx context.Context, jobs []BatchJob) (*BatchSubmitResponse, error) {
	body := map[string]any{"jobs": jobs}
	var resp BatchSubmitResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/batch", body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchSubmitJsonl submits a batch of jobs using JSONL format.
// Each line in the JSONL string is a JSON object with model, prompt, etc.
func (c *Client) BatchSubmitJsonl(ctx context.Context, jsonl string) (*BatchJsonlResponse, error) {
	body := map[string]any{"jsonl": jsonl}
	var resp BatchJsonlResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/batch/jsonl", body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchJobs lists all batch jobs for the account.
func (c *Client) BatchJobs(ctx context.Context) (*BatchJobsResponse, error) {
	var resp BatchJobsResponse
	_, err := c.doJSON(ctx, "GET", "/qai/v1/batch/jobs", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchJob gets the status and result of a single batch job.
func (c *Client) BatchJob(ctx context.Context, id string) (*BatchJobInfo, error) {
	path := fmt.Sprintf("/qai/v1/batch/jobs/%s", id)
	var resp BatchJobInfo
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
