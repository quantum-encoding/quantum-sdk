package qai

import (
	"context"
	"fmt"
	"net/url"
)

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

// MissionCreateRequest is the request body for creating a mission.
type MissionCreateRequest struct {
	// High-level task description.
	Goal string `json:"goal"`

	// Strategy: "wave" (default), "dag", "mapreduce", "refinement", "branch".
	Strategy string `json:"strategy,omitempty"`

	// Conductor model (default: claude-sonnet-4-6).
	ConductorModel string `json:"conductor_model,omitempty"`

	// Worker team configuration.
	Workers map[string]MissionWorkerConfig `json:"workers,omitempty"`

	// Maximum orchestration steps (default: 25).
	MaxSteps *int32 `json:"max_steps,omitempty"`

	// Custom system prompt for the conductor.
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Existing session ID for context continuity.
	SessionID string `json:"session_id,omitempty"`
}

// MissionChatRequest is the request body for chatting with a mission's architect.
type MissionChatRequest struct {
	// Message to send to the architect.
	Message string `json:"message"`

	// Enable streaming (not yet supported).
	Stream *bool `json:"stream,omitempty"`
}

// MissionPlanUpdate is the request body for updating a mission plan.
type MissionPlanUpdate struct {
	// Updated task list.
	Tasks []map[string]any `json:"tasks,omitempty"`

	// Updated worker configuration.
	Workers map[string]MissionWorkerConfig `json:"workers,omitempty"`

	// Additional system prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Updated max steps.
	MaxSteps *int32 `json:"max_steps,omitempty"`

	// Additional context to inject.
	Context string `json:"context,omitempty"`
}

// MissionConfirmStructure is the request body for confirming/rejecting a mission structure.
type MissionConfirmStructure struct {
	// Whether the structure is approved.
	Confirmed bool `json:"confirmed"`

	// Rejection reason or modification notes.
	Feedback string `json:"feedback,omitempty"`
}

// MissionApproveRequest is the request body for approving a completed mission.
type MissionApproveRequest struct {
	// Git commit SHA associated with the mission output.
	CommitSHA string `json:"commit_sha,omitempty"`

	// Approval comment.
	Comment string `json:"comment,omitempty"`
}

// MissionImportRequest is the request body for importing a plan as a new mission.
type MissionImportRequest struct {
	// Mission goal.
	Goal string `json:"goal"`

	// Strategy.
	Strategy string `json:"strategy,omitempty"`

	// Conductor model.
	ConductorModel string `json:"conductor_model,omitempty"`

	// Worker configuration.
	Workers map[string]MissionWorkerConfig `json:"workers,omitempty"`

	// Pre-defined tasks.
	Tasks []map[string]any `json:"tasks"`

	// System prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Maximum steps.
	MaxSteps *int32 `json:"max_steps,omitempty"`

	// Auto-execute after import.
	AutoExecute bool `json:"auto_execute,omitempty"`
}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

// MissionCreateResponse is the response from mission creation.
type MissionCreateResponse struct {
	// Mission identifier.
	MissionID string `json:"mission_id"`

	// Initial status.
	Status string `json:"status"`

	// Session ID for conversation context.
	SessionID string `json:"session_id,omitempty"`

	// Conductor model used.
	ConductorModel string `json:"conductor_model,omitempty"`

	// Strategy used.
	Strategy string `json:"strategy,omitempty"`

	// Worker configuration.
	Workers map[string]MissionWorkerConfig `json:"workers,omitempty"`

	// Creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// Request identifier.
	RequestID string `json:"request_id,omitempty"`
}

// MissionDetail is the full mission detail (from GET /missions/{id}).
type MissionDetail struct {
	// Mission identifier.
	ID string `json:"id,omitempty"`

	// User who created the mission.
	UserID string `json:"user_id,omitempty"`

	// Mission goal.
	Goal string `json:"goal,omitempty"`

	// Strategy.
	Strategy string `json:"strategy,omitempty"`

	// Conductor model.
	ConductorModel string `json:"conductor_model,omitempty"`

	// Current status.
	Status string `json:"status,omitempty"`

	// Creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`

	// Start timestamp.
	StartedAt string `json:"started_at,omitempty"`

	// Completion timestamp.
	CompletedAt string `json:"completed_at,omitempty"`

	// Error message if failed.
	Error string `json:"error,omitempty"`

	// Total cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// Number of steps executed.
	TotalSteps int32 `json:"total_steps"`

	// Session ID.
	SessionID string `json:"session_id,omitempty"`

	// Final result text.
	Result string `json:"result,omitempty"`

	// Tasks within the mission.
	Tasks []MissionTask `json:"tasks"`

	// Whether the mission was approved.
	Approved bool `json:"approved"`

	// Commit SHA (if approved).
	CommitSHA string `json:"commit_sha,omitempty"`
}

// MissionTask is a task within a mission.
type MissionTask struct {
	// Task identifier.
	ID string `json:"id,omitempty"`

	// Task name.
	Name string `json:"name,omitempty"`

	// Task description.
	Description string `json:"description,omitempty"`

	// Assigned worker name.
	Worker string `json:"worker,omitempty"`

	// Model used.
	Model string `json:"model,omitempty"`

	// Task status.
	Status string `json:"status,omitempty"`

	// Task result.
	Result string `json:"result,omitempty"`

	// Error message if failed.
	Error string `json:"error,omitempty"`

	// Step number.
	Step int32 `json:"step"`

	// Input tokens used.
	TokensIn int32 `json:"tokens_in"`

	// Output tokens used.
	TokensOut int32 `json:"tokens_out"`
}

// MissionListResponse is the response from listing missions.
type MissionListResponse struct {
	// List of missions.
	Missions []MissionDetail `json:"missions"`
}

// MissionChatResponse is the response from chatting with the architect.
type MissionChatResponse struct {
	// Mission identifier.
	MissionID string `json:"mission_id,omitempty"`

	// Architect's response content.
	Content string `json:"content,omitempty"`

	// Model used.
	Model string `json:"model,omitempty"`

	// Cost in ticks.
	CostTicks int64 `json:"cost_ticks"`

	// Token usage.
	Usage *MissionChatUsage `json:"usage,omitempty"`
}

// MissionChatUsage is the token usage for a mission chat response.
type MissionChatUsage struct {
	InputTokens  int32 `json:"input_tokens"`
	OutputTokens int32 `json:"output_tokens"`
}

// MissionCheckpoint is a git checkpoint within a mission.
type MissionCheckpoint struct {
	// Checkpoint identifier.
	ID string `json:"id,omitempty"`

	// Commit SHA.
	CommitSHA string `json:"commit_sha,omitempty"`

	// Checkpoint message.
	Message string `json:"message,omitempty"`

	// Creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`
}

// MissionCheckpointsResponse is the response from listing checkpoints.
type MissionCheckpointsResponse struct {
	MissionID   string              `json:"mission_id,omitempty"`
	Checkpoints []MissionCheckpoint `json:"checkpoints"`
}

// MissionStatusResponse is a generic status response for mission operations.
type MissionStatusResponse struct {
	MissionID string `json:"mission_id,omitempty"`
	Status    string `json:"status,omitempty"`
	Confirmed *bool  `json:"confirmed,omitempty"`
	Approved  *bool  `json:"approved,omitempty"`
	Deleted   *bool  `json:"deleted,omitempty"`
	Updated   *bool  `json:"updated,omitempty"`
	CommitSHA string `json:"commit_sha,omitempty"`
}

// ---------------------------------------------------------------------------
// Client methods
// ---------------------------------------------------------------------------

// MissionCreate creates and executes a mission asynchronously.
func (c *Client) MissionCreate(ctx context.Context, req *MissionCreateRequest) (*MissionCreateResponse, error) {
	var resp MissionCreateResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/missions/create", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionList lists missions for the authenticated user.
// Pass an empty string for status to list all missions.
func (c *Client) MissionList(ctx context.Context, status string) (*MissionListResponse, error) {
	path := "/qai/v1/missions/list"
	if status != "" {
		path = fmt.Sprintf("/qai/v1/missions/list?status=%s", url.QueryEscape(status))
	}
	var resp MissionListResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionGet returns mission details including tasks.
func (c *Client) MissionGet(ctx context.Context, missionID string) (*MissionDetail, error) {
	var resp MissionDetail
	_, err := c.doJSON(ctx, "GET", fmt.Sprintf("/qai/v1/missions/%s", missionID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionDelete deletes a mission.
func (c *Client) MissionDelete(ctx context.Context, missionID string) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/qai/v1/missions/%s", missionID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionCancel cancels a running mission.
func (c *Client) MissionCancel(ctx context.Context, missionID string) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/cancel", missionID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionPause pauses a running mission.
func (c *Client) MissionPause(ctx context.Context, missionID string) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/pause", missionID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionResume resumes a paused mission.
func (c *Client) MissionResume(ctx context.Context, missionID string) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/resume", missionID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionChat chats with the mission's architect.
func (c *Client) MissionChat(ctx context.Context, missionID string, req *MissionChatRequest) (*MissionChatResponse, error) {
	var resp MissionChatResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/chat", missionID), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionRetryTask retries a failed task.
func (c *Client) MissionRetryTask(ctx context.Context, missionID, taskID string) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/retry/%s", missionID, taskID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionApprove approves a completed mission.
func (c *Client) MissionApprove(ctx context.Context, missionID string, req *MissionApproveRequest) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/approve", missionID), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionUpdatePlan updates the mission plan.
func (c *Client) MissionUpdatePlan(ctx context.Context, missionID string, req *MissionPlanUpdate) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "PUT", fmt.Sprintf("/qai/v1/missions/%s/plan", missionID), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionConfirmStructure confirms or rejects the proposed execution structure.
func (c *Client) MissionConfirmStructure(ctx context.Context, missionID string, req *MissionConfirmStructure) (*MissionStatusResponse, error) {
	var resp MissionStatusResponse
	_, err := c.doJSON(ctx, "POST", fmt.Sprintf("/qai/v1/missions/%s/confirm-structure", missionID), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionCheckpoints lists git checkpoints for a mission.
func (c *Client) MissionCheckpoints(ctx context.Context, missionID string) (*MissionCheckpointsResponse, error) {
	var resp MissionCheckpointsResponse
	_, err := c.doJSON(ctx, "GET", fmt.Sprintf("/qai/v1/missions/%s/checkpoints", missionID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// MissionImport imports an existing plan as a new mission.
func (c *Client) MissionImport(ctx context.Context, req *MissionImportRequest) (*MissionCreateResponse, error) {
	var resp MissionCreateResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/missions/import", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
