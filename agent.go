package qai

import "context"

// AgentRequest is the request body for server-side agent orchestration.
// The server creates a conductor and workers harness and runs the full
// agent loop internally, streaming events back via SSE.
type AgentRequest struct {
	// SessionID is the conversation session (optional — creates new if empty).
	SessionID string `json:"session_id,omitempty"`

	// Task is the user's request (required).
	Task string `json:"task"`

	// ConductorModel is the model that plans and delegates. Default: claude-sonnet-4-6.
	ConductorModel string `json:"conductor_model,omitempty"`

	// Workers defines the agent team. If empty, uses default team.
	Workers []AgentWorkerConfig `json:"workers,omitempty"`

	// MaxSteps for the orchestration loop (default 10).
	MaxSteps int `json:"max_steps,omitempty"`

	// SystemPrompt for the conductor (optional).
	SystemPrompt string `json:"system_prompt,omitempty"`

	// ContextConfig for session management (optional, used when creating new session).
	ContextConfig *ContextConfig `json:"context_config,omitempty"`
}

// AgentWorkerConfig defines a worker in the agent team.
type AgentWorkerConfig struct {
	// Name is the worker name (e.g. "reader", "coder").
	Name string `json:"name"`

	// Model is the model ID for this worker.
	Model string `json:"model"`

	// Tier is the cost tier: "cheap", "mid", or "expensive".
	Tier string `json:"tier"`

	// Description explains what this worker is good at.
	Description string `json:"description"`
}

// AgentEvent is a single SSE event from an agent run stream.
// Events have different types: "agent_session", "agent_step", "agent_result",
// "agent_error", "usage", "done".
type AgentEvent = StreamEvent

// AgentRun starts a server-side agent orchestration and returns a stream of events.
//
// The conductor model plans the work and delegates to worker models. Events are
// streamed as the agent progresses. The last event has Done=true.
//
//	events, err := client.AgentRun(ctx, &qai.AgentRequest{
//	    Task: "Summarize the top 3 HN stories",
//	})
//	for ev := range events {
//	    fmt.Printf("[%s] %s\n", ev.Type, ev.Delta)
//	}
func (c *Client) AgentRun(ctx context.Context, req *AgentRequest) (<-chan AgentEvent, error) {
	resp, _, err := c.doStreamRaw(ctx, "/qai/v1/agent", req)
	if err != nil {
		return nil, err
	}

	return readSSEStream(ctx, resp), nil
}

// MissionRequest is the request body for full project execution via the missions endpoint.
type MissionRequest struct {
	// Goal is the high-level task description (required).
	Goal string `json:"goal"`

	// Strategy is the execution strategy: "wave" (default), "dag", "mapreduce", "refinement", "branch".
	Strategy string `json:"strategy,omitempty"`

	// ConductorModel plans and reviews. Default: claude-sonnet-4-6.
	ConductorModel string `json:"conductor_model,omitempty"`

	// Workers defines the agent team. Keys are worker names, values are configs.
	// If empty, uses cost-optimized defaults.
	Workers map[string]MissionWorkerConfig `json:"workers,omitempty"`

	// MaxSteps for the orchestration loop (default 25).
	MaxSteps int `json:"max_steps,omitempty"`

	// SystemPrompt for the conductor (optional).
	SystemPrompt string `json:"system_prompt,omitempty"`

	// SessionID for conversation context (optional).
	SessionID string `json:"session_id,omitempty"`

	// AutoPlan controls whether the conductor generates a plan before executing. Default: true.
	AutoPlan *bool `json:"auto_plan,omitempty"`

	// ContextConfig for session management (optional).
	ContextConfig *ContextConfig `json:"context_config,omitempty"`
}

// MissionWorkerConfig defines a worker in the mission team.
type MissionWorkerConfig struct {
	// Model is the model ID for this worker.
	Model string `json:"model"`

	// Tier is the cost tier: "cheap", "mid", or "expensive".
	Tier string `json:"tier"`

	// Description explains what this worker is good at.
	Description string `json:"description,omitempty"`
}

// MissionEvent is a single SSE event from a mission run stream.
// Events include: "mission_started", "step_detail", "mission_completed",
// "mission_failed", "usage", "done", and execution-level events.
type MissionEvent = StreamEvent

// MissionRun starts a full project execution and returns a stream of events.
//
// The conductor plans the work, delegates to workers, and streams progress.
// The last event has Done=true.
//
//	events, err := client.MissionRun(ctx, &qai.MissionRequest{
//	    Goal: "Build a REST API for todo items",
//	})
//	for ev := range events {
//	    fmt.Printf("[%s]\n", ev.Type)
//	}
func (c *Client) MissionRun(ctx context.Context, req *MissionRequest) (<-chan MissionEvent, error) {
	resp, _, err := c.doStreamRaw(ctx, "/qai/v1/missions", req)
	if err != nil {
		return nil, err
	}

	return readSSEStream(ctx, resp), nil
}
