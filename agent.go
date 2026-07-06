package qai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AgentRequest is the request body for the legacy server-side agent
// orchestration surface. The backend's /qai/v1/agent endpoint is NOT this
// shape — it is a stateless, non-streaming tool-call passthrough (see Agent
// below). AgentRequest is the missions-flavoured shape historically passed
// to AgentRun; AgentRun now translates it to a MissionRequest and runs it
// against /qai/v1/missions, which is the SSE conductor endpoint.
//
// Deprecated: prefer MissionRun (with an explicit MissionRequest) for new
// code — the mapping is direct and the shape is the contract source of
// truth. AgentRun is kept as a back-compat shim and may be removed in a
// future major bump.
type AgentRequest struct {
	// SessionID is the conversation session (optional — creates new if empty).
	SessionID string `json:"session_id,omitempty"`

	// Task is the user's request (required). Mapped to MissionRequest.Goal.
	Task string `json:"task"`

	// ConductorModel is the model that plans and delegates. Default: claude-sonnet-4-6.
	ConductorModel string `json:"conductor_model,omitempty"`

	// Workers defines the agent team. If empty, uses default team.
	// The slice is mapped to a map keyed by each worker's Name.
	Workers []AgentWorkerConfig `json:"workers,omitempty"`

	// MaxSteps for the orchestration loop (default 25 on /qai/v1/missions).
	MaxSteps int `json:"max_steps,omitempty"`

	// SystemPrompt for the conductor (optional).
	SystemPrompt string `json:"system_prompt,omitempty"`

	// ContextConfig for session management (optional, used when creating new session).
	ContextConfig *ContextConfig `json:"context_config,omitempty"`
}

// AgentWorker describes a worker agent in a multi-agent run.
type AgentWorker struct {
	// Name is the worker name (e.g. "reader", "coder"). Used as the map key
	// when AgentRequest.Workers is translated into MissionRequest.Workers.
	Name string `json:"name"`

	// Model is the model ID for this worker.
	Model string `json:"model,omitempty"`

	// Tier is the cost tier: "cheap", "mid", or "expensive".
	Tier string `json:"tier,omitempty"`

	// Description explains what this worker is good at.
	Description string `json:"description,omitempty"`
}

// AgentWorkerConfig is an alias for AgentWorker for backwards compatibility.
type AgentWorkerConfig = AgentWorker

// AgentStreamEvent is a single event from an agent or mission SSE stream.
type AgentStreamEvent struct {
	// EventType is the event type (e.g. "step", "thought", "tool_call", "tool_result", "message", "error", "done").
	EventType string `json:"type"`

	// Data contains the raw JSON payload for the caller to interpret.
	Data map[string]any `json:"data,omitempty"`
}

// AgentEvent is a single SSE event from an agent run stream.
// Events have different types: "agent_session", "agent_step", "agent_result",
// "agent_error", "usage", "done".
type AgentEvent = StreamEvent

// AgentRun starts a server-side agent orchestration and returns a stream of
// events. The conductor model plans the work and delegates to worker models.
// Events are streamed as the agent progresses. The last event has Done=true.
//
// NOTE: this method used to POST to /qai/v1/agent, which is a NON-STREAMING
// JSON endpoint requiring {model, messages, tools, capabilities,
// system_prompt} — the old shape 400'd with "model is required". AgentRun now
// translates the AgentRequest into a MissionRequest and runs it against
// /qai/v1/missions, the SSE conductor endpoint whose shape matches.
//
// BREAKING CHANGE: callers that relied on the (broken) /qai/v1/agent path
// must switch to the new Agent() method for the stateless tool-call
// passthrough, or to MissionRun for full multi-worker orchestration.
//
//	events, err := client.AgentRun(ctx, &qai.AgentRequest{
//	    Task: "Summarize the top 3 HN stories",
//	})
//	for ev := range events {
//	    fmt.Printf("[%s] %s\n", ev.Type, ev.Delta)
//	}
func (c *Client) AgentRun(ctx context.Context, req *AgentRequest) (<-chan AgentEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("qai: AgentRun: nil request")
	}

	// Translate the legacy AgentRequest into the real MissionRequest shape
	// that /qai/v1/missions expects (Goal + map[string]MissionWorkerConfig).
	mr := &MissionRequest{
		Goal:           req.Task,
		ConductorModel: req.ConductorModel,
		MaxSteps:       req.MaxSteps,
		SystemPrompt:   req.SystemPrompt,
		SessionID:      req.SessionID,
		ContextConfig:  req.ContextConfig,
	}
	if len(req.Workers) > 0 {
		mr.Workers = make(map[string]MissionWorkerConfig, len(req.Workers))
		for _, w := range req.Workers {
			key := w.Name
			if key == "" {
				// Fall back to the model name so every worker lands in the
				// map (missions keys on worker name, not model).
				key = w.Model
			}
			mr.Workers[key] = MissionWorkerConfig{
				Model:       w.Model,
				Tier:        w.Tier,
				Description: w.Description,
			}
		}
	}

	return c.MissionRun(ctx, mr)
}

// ── Stateless tool-call passthrough: POST /qai/v1/agent ─────────────

// AgentCallRequest is the body for the non-streaming /qai/v1/agent endpoint —
// a stateless tool-call passthrough. The server does NOT execute tools; it
// runs one provider generation with the given tools and returns any tool_use
// the model produced. The client executes tools locally and sends the
// results back in the next request's message history.
//
// Mirror of internal/server/routes_agent.go agentRequest. Field names are
// snake_case so the Go and Zig backends accept the same payload.
type AgentCallRequest struct {
	// Model is the model ID (required). Unknown models 400 with unknown_model.
	Model string `json:"model"`

	// Messages is the conversation history (required, non-empty).
	Messages []AgentMessage `json:"messages"`

	// Tools are the canonical, provider-agnostic tool definitions.
	Tools []AgentToolDef `json:"tools,omitempty"`

	// Capabilities filters which tools the model sees:
	//   nil  → all tools (no filtering)
	//   []   → no tools (Safe Mode)
	//   list → allowlist by tool name (only matching tools forwarded)
	// Use *[]string so JSON null distinguishes from [] from omitted.
	Capabilities *[]string `json:"capabilities,omitempty"`

	// SystemPrompt is an optional system prompt for the generation.
	SystemPrompt string `json:"system_prompt,omitempty"`

	// MaxTokens caps the response length.
	MaxTokens *int32 `json:"max_tokens,omitempty"`

	// Temperature controls randomness.
	Temperature *float32 `json:"temperature,omitempty"`

	// IdempotencyKey is sent as the Idempotency-Key header; auto-generated
	// if empty. /qai/v1/agent is billing-bearing, so retries dedupe.
	IdempotencyKey string `json:"-"`
}

// idempotencyKey returns the caller-set key, auto-generating one if empty.
func (r *AgentCallRequest) idempotencyKey() string {
	if r.IdempotencyKey == "" {
		r.IdempotencyKey = newIdempotencyKey()
	}
	return r.IdempotencyKey
}

// AgentMessage is one message in an /qai/v1/agent conversation.
type AgentMessage struct {
	// Role is "user", "assistant", "tool", or "system".
	Role string `json:"role"`

	// Content is the text content of the message.
	Content string `json:"content,omitempty"`

	// ToolCallID references the tool_use ID (for "tool" role results).
	ToolCallID string `json:"tool_call_id,omitempty"`

	// ToolUse carries the tool calls the model made on its previous turn
	// (for "assistant" role, sent back in history).
	ToolUse []AgentToolUseIn `json:"tool_use,omitempty"`

	// IsError marks a tool result as an error (for "tool" role).
	IsError bool `json:"is_error,omitempty"`
}

// AgentToolDef is a canonical tool definition. InputSchema is JSON; either a
// JSON-encoded string or an inline JSON object is accepted.
type AgentToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

// AgentToolUseIn is a tool call from a previous assistant turn, replayed in
// history. Input is JSON-encoded.
type AgentToolUseIn struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input,omitempty"`
}

// AgentCallResponse is the response from /qai/v1/agent. ToolUse is present
// (non-empty) only when StopReason == "tool_use"; each ToolUse.Input is the
// parsed JSON object (not a string).
type AgentCallResponse struct {
	// ID is the request identifier.
	ID string `json:"id"`

	// Model is the model that generated the response.
	Model string `json:"model"`

	// StopReason is "end_turn" or "tool_use" (tool_use iff tool calls exist).
	StopReason string `json:"stop_reason"`

	// Content is the list of text content parts.
	Content []AgentContentPart `json:"content"`

	// ToolUse is the tool calls the model is requesting (empty unless
	// StopReason == "tool_use").
	ToolUse []AgentToolUseOut `json:"tool_use,omitempty"`

	// Usage is the token usage for this generation. The agent endpoint's
	// usage only carries input_tokens/output_tokens (no reasoning_tokens or
	// cost_ticks in the body); CostTicks and BalanceAfter are populated
	// from response headers.
	Usage AgentUsage `json:"usage"`

	// CostTicks is the total cost from the X-QAI-Cost-Ticks header.
	CostTicks int64 `json:"-"`

	// BalanceAfter is the wallet balance after this call, from the
	// X-QAI-Balance-After header. Signed — a claw-back can make it negative.
	BalanceAfter int64 `json:"-"`

	// RequestID is from the X-QAI-Request-Id header.
	RequestID string `json:"-"`
}

// AgentContentPart is one part of the agent response content.
type AgentContentPart struct {
	Type string `json:"type"`           // "text"
	Text string `json:"text,omitempty"` // text payload
}

// AgentToolUseOut is a tool call the model produced in this turn.
type AgentToolUseOut struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// AgentUsage is the token usage from /qai/v1/agent.
type AgentUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Text returns the concatenated text content from the agent response.
func (r *AgentCallResponse) Text() string {
	var parts []string
	for _, p := range r.Content {
		if p.Type == "text" {
			parts = append(parts, p.Text)
		}
	}
	return strings.Join(parts, "")
}

// IsToolUse reports whether the model is requesting tool execution
// (StopReason == "tool_use").
func (r *AgentCallResponse) IsToolUse() bool { return r.StopReason == StopReasonToolUse }

// Agent runs one non-streaming tool-call passthrough against /qai/v1/agent.
//
// The server runs a single provider generation with the supplied tools and
// returns any tool_use the model produced — it does NOT execute tools. The
// caller executes tools locally and sends the results back in the next
// request's Messages history (role "tool", ToolCallID set, Content = result
// or IsError = true on failure).
//
// This is the contract-correct surface for /qai/v1/agent
// (routes_agent.go): {model, messages, tools, capabilities, system_prompt}
// → {id, model, stop_reason, content, tool_use, usage}. The stream field is
// accepted but ignored by the server; use MissionRun for SSE orchestration.
func (c *Client) Agent(ctx context.Context, req *AgentCallRequest) (*AgentCallResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("qai: Agent: nil request")
	}

	var resp AgentCallResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/agent", req, &resp)
	if err != nil {
		return nil, err
	}
	resp.RequestID = meta.RequestID
	resp.CostTicks = meta.CostTicks
	resp.BalanceAfter = meta.BalanceAfter
	if resp.Model == "" {
		resp.Model = meta.Model
	}
	return &resp, nil
}

// MissionRequest is the request body for full project execution via the missions endpoint.
type MissionRequest struct {
	// Goal is the high-level task description (required).
	Goal string `json:"goal"`

	// Strategy is the execution strategy: "wave" (default), "dag", "mapreduce", "refinement", "branch".
	Strategy string `json:"strategy,omitempty"`

	// ConductorModel plans and reviews. Default: claude-sonnet-4-6.
	ConductorModel string `json:"conductor_model,omitempty"`

	// ConductorTier overrides the conductor's cost tier. Default: "expensive".
	// Set to "cheap" when using a fast router (e.g. Grok 4.1) as conductor.
	ConductorTier string `json:"conductor_tier,omitempty"`

	// Workers defines the agent team. Keys are worker names, values are configs.
	// If empty, uses cost-optimized defaults.
	Workers map[string]MissionWorkerConfig `json:"workers,omitempty"`

	// MaxSteps for the orchestration loop (default 25).
	MaxSteps int `json:"max_steps,omitempty"`

	// SystemPrompt for the conductor (optional).
	SystemPrompt string `json:"system_prompt,omitempty"`

	// SessionID for conversation context (optional).
	SessionID string `json:"session_id,omitempty"`

	// DeploymentID routes workers through a deployed Vertex endpoint (optional).
	DeploymentID string `json:"deployment_id,omitempty"`

	// WorkerModel is the model for codegen worker nodes (optional).
	WorkerModel string `json:"worker_model,omitempty"`

	// BuildCommand is the build command for codegen verification (optional).
	BuildCommand string `json:"build_command,omitempty"`

	// WorkspacePath is the workspace directory for generated files (optional).
	WorkspacePath string `json:"workspace_path,omitempty"`

	// AutoPlan controls whether the conductor generates a plan before executing. Default: true.
	AutoPlan *bool `json:"auto_plan,omitempty"`

	// ContextConfig for session management (optional).
	ContextConfig *ContextConfig `json:"context_config,omitempty"`
}

// MissionWorker describes a named worker for a mission.
type MissionWorker struct {
	// Model is the model ID for this worker.
	Model string `json:"model,omitempty"`

	// Tier is the cost tier: "cheap", "mid", or "expensive".
	Tier string `json:"tier,omitempty"`

	// Description explains what this worker is good at.
	Description string `json:"description,omitempty"`

	// EscalateTo names a worker to fall back to when this worker fails.
	// E.g. a cheap coder escalates to an expensive coder after MaxRetries failures.
	EscalateTo string `json:"escalate_to,omitempty"`

	// MaxRetries before escalating (default 1 = escalate on first failure).
	MaxRetries int `json:"max_retries,omitempty"`
}

// MissionWorkerConfig is an alias for MissionWorker for backwards compatibility.
type MissionWorkerConfig = MissionWorker

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
