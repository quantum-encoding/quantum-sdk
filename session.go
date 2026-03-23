package qai

import (
	"context"
	"encoding/json"
)

// SessionChatRequest is the request body for session-based chat.
// Instead of sending full history, the client sends a session_id and a new message.
type SessionChatRequest struct {
	// SessionID is the conversation session. If empty, a new session is created.
	SessionID string `json:"session_id,omitempty"`

	// Message is the new user message (required unless sending tool results).
	Message string `json:"message,omitempty"`

	// Model override (optional — uses session model if not set).
	Model string `json:"model,omitempty"`

	// Tools for this request (optional).
	Tools []ChatTool `json:"tools,omitempty"`

	// ToolResults from the previous response's tool calls (optional).
	ToolResults []SessionToolResult `json:"tool_results,omitempty"`

	// Stream enables SSE streaming.
	Stream bool `json:"stream,omitempty"`

	// SystemPrompt is the system prompt (used when creating a new session).
	SystemPrompt string `json:"system_prompt,omitempty"`

	// ContextConfig controls context management for the session.
	ContextConfig *ContextConfig `json:"context_config,omitempty"`

	// ProviderOptions passes provider-specific settings.
	ProviderOptions map[string]json.RawMessage `json:"provider_options,omitempty"`
}

// SessionToolResult is a tool execution result from the client.
type SessionToolResult struct {
	// ToolCallID references the tool_use ID from the previous response.
	ToolCallID string `json:"tool_call_id"`

	// Content is the tool execution result.
	Content string `json:"content"`

	// IsError indicates whether the tool execution failed.
	IsError bool `json:"is_error,omitempty"`
}

// ContextConfig controls context management for a session.
type ContextConfig struct {
	// MaxTokens is the maximum context window size.
	MaxTokens int `json:"max_tokens,omitempty"`

	// AutoCompact controls whether to enable automatic context compaction.
	AutoCompact *bool `json:"auto_compact,omitempty"`

	// CompactionThreshold is the percentage of max_tokens before compaction triggers.
	CompactionThreshold float64 `json:"compaction_threshold,omitempty"`
}

// ToolResult is a tool result to feed back into the session (sdk-graph canonical name).
type ToolResult struct {
	// ToolCallID is the tool_use ID this result corresponds to.
	ToolCallID string `json:"tool_call_id"`

	// Content is the result content.
	Content string `json:"content"`

	// IsError indicates whether this result is an error.
	IsError *bool `json:"is_error,omitempty"`
}

// SessionContext is context metadata returned with session responses (sdk-graph canonical name).
type SessionContext struct {
	// TurnCount is the number of conversation turns in the session.
	TurnCount int64 `json:"turn_count"`

	// EstimatedTokens is the estimated total tokens in the session context.
	EstimatedTokens int64 `json:"estimated_tokens"`

	// Compacted indicates whether context was compacted during this turn.
	Compacted bool `json:"compacted,omitempty"`

	// CompactionNote is a note about the compaction, if any.
	CompactionNote string `json:"compaction_note,omitempty"`
}

// SessionChatResponse is the response from session-based chat.
type SessionChatResponse struct {
	// SessionID is the conversation session identifier.
	SessionID string `json:"session_id"`

	// Response is the chat response from the model.
	Response *ChatResponse `json:"response"`

	// Context contains metadata about the conversation's context state.
	Context *ContextMetadata `json:"context,omitempty"`
}

// ContextMetadata provides information about the conversation's context state.
type ContextMetadata struct {
	// TurnCount is the total number of turns in the conversation.
	TurnCount int `json:"turn_count"`

	// EstimatedTokens is the estimated token count of the current context.
	EstimatedTokens int `json:"estimated_tokens"`

	// Compacted is true if the conversation was compacted in this request.
	Compacted bool `json:"compacted,omitempty"`

	// CompactionNote describes the compaction that occurred.
	CompactionNote string `json:"compaction_note,omitempty"`

	// ToolsCleared is the number of stale tool results that were cleared.
	ToolsCleared int `json:"tools_cleared,omitempty"`
}

// ChatSession sends a session-based chat request. The server manages conversation
// history and context automatically.
//
// For new sessions, leave SessionID empty and provide a Model and optional SystemPrompt.
// For existing sessions, provide the SessionID returned from the first call.
func (c *Client) ChatSession(ctx context.Context, req *SessionChatRequest) (*SessionChatResponse, error) {
	req.Stream = false

	var resp SessionChatResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/chat/session", req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Response != nil {
		if resp.Response.CostTicks == 0 {
			resp.Response.CostTicks = meta.CostTicks
		}
		if resp.Response.RequestID == "" {
			resp.Response.RequestID = meta.RequestID
		}
	}

	return &resp, nil
}

// ChatSessionStream sends a streaming session-based chat request.
// The returned channel emits StreamEvent values. The first event has type "session"
// with the session_id. Subsequent events follow the same pattern as ChatStream.
func (c *Client) ChatSessionStream(ctx context.Context, req *SessionChatRequest) (<-chan StreamEvent, error) {
	req.Stream = true

	resp, _, err := c.doStreamRaw(ctx, "/qai/v1/chat/session", req)
	if err != nil {
		return nil, err
	}

	return readSSEStream(ctx, resp), nil
}
