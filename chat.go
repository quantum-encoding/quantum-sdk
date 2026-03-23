package qai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ChatRequest is the request body for text generation.
type ChatRequest struct {
	// Model is the model ID that determines provider routing (e.g. "claude-sonnet-4-6", "grok-4-1-fast-non-reasoning").
	Model string `json:"model"`

	// Messages is the conversation history.
	Messages []ChatMessage `json:"messages"`

	// Tools defines functions the model can call.
	Tools []ChatTool `json:"tools,omitempty"`

	// Stream enables server-sent event streaming. Use ChatStream instead of Chat for streaming.
	Stream bool `json:"stream,omitempty"`

	// Temperature controls randomness (0.0–2.0).
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens limits the response length.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// ProviderOptions passes provider-specific settings (e.g. Anthropic thinking, xAI search).
	//
	// Example: map[string]any{"anthropic": map[string]any{"thinking": map[string]any{"budget_tokens": 10000}}}
	ProviderOptions map[string]any `json:"provider_options,omitempty"`
}

// ChatMessage is a single message in a conversation.
type ChatMessage struct {
	// Role is one of "system", "user", "assistant", or "tool".
	Role string `json:"role"`

	// Content is the text content of the message.
	Content string `json:"content,omitempty"`

	// ContentBlocks is structured content for assistant messages with tool calls.
	// When present, takes precedence over Content.
	ContentBlocks []ContentBlock `json:"content_blocks,omitempty"`

	// ToolCallID is required when Role is "tool" — it references the tool_use ID.
	ToolCallID string `json:"tool_call_id,omitempty"`

	// IsError indicates whether a tool result is an error.
	IsError bool `json:"is_error,omitempty"`
}

// ContentBlock is a single block in the response content array.
type ContentBlock struct {
	// Type is one of "text", "thinking", or "tool_use".
	Type string `json:"type"`

	// BlockType is an alias for Type (sdk-graph canonical name).
	BlockType string `json:"block_type,omitempty"`

	// Text holds the content for "text" and "thinking" blocks.
	Text string `json:"text,omitempty"`

	// ID is the tool call identifier for "tool_use" blocks.
	ID string `json:"id,omitempty"`

	// Name is the function name for "tool_use" blocks.
	Name string `json:"name,omitempty"`

	// Input is the function arguments for "tool_use" blocks.
	Input map[string]any `json:"input,omitempty"`

	// ThoughtSignature is the Gemini thought signature — must be echoed back with tool results.
	ThoughtSignature string `json:"thought_signature,omitempty"`
}

// ChatTool defines a function the model can call.
type ChatTool struct {
	// Name is the function name.
	Name string `json:"name"`

	// Description explains what the function does.
	Description string `json:"description"`

	// Parameters is the JSON Schema for the function's arguments.
	Parameters map[string]any `json:"parameters,omitempty"`
}

// ChatResponse is the response from a non-streaming chat request.
type ChatResponse struct {
	// ID is the unique request identifier.
	ID string `json:"id"`

	// Model is the model that generated the response.
	Model string `json:"model"`

	// Content is the list of content blocks (text, thinking, tool_use).
	Content []ContentBlock `json:"content"`

	// Usage contains token counts and cost.
	Usage *ChatUsage `json:"usage,omitempty"`

	// StopReason indicates why generation stopped ("end_turn", "tool_use", "max_tokens").
	StopReason string `json:"stop_reason"`

	// CostTicks is the total cost from the X-QAI-Cost-Ticks header.
	CostTicks int64 `json:"cost_ticks"`

	// RequestID is from the X-QAI-Request-Id header.
	RequestID string `json:"request_id"`
}

// ChatUsage contains token counts and cost for a chat response.
type ChatUsage struct {
	InputTokens  int   `json:"input_tokens"`
	OutputTokens int   `json:"output_tokens"`
	CostTicks    int64 `json:"cost_ticks"`
}

// Text returns the concatenated text content from the response, ignoring thinking and tool_use blocks.
func (r *ChatResponse) Text() string {
	var parts []string
	for _, block := range r.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "")
}

// Thinking returns the concatenated thinking content from the response.
func (r *ChatResponse) Thinking() string {
	var parts []string
	for _, block := range r.Content {
		if block.Type == "thinking" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "")
}

// ToolCalls returns all tool_use blocks from the response.
func (r *ChatResponse) ToolCalls() []ContentBlock {
	var calls []ContentBlock
	for _, block := range r.Content {
		if block.Type == "tool_use" {
			calls = append(calls, block)
		}
	}
	return calls
}

// Chat sends a non-streaming text generation request.
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	req.Stream = false

	var resp ChatResponse
	meta, err := c.doJSON(ctx, "POST", "/qai/v1/chat", req, &resp)
	if err != nil {
		return nil, err
	}

	resp.CostTicks = meta.CostTicks
	resp.RequestID = meta.RequestID
	if resp.Model == "" {
		resp.Model = meta.Model
	}

	return &resp, nil
}

// StreamEvent is a single event from an SSE chat stream.
type StreamEvent struct {
	// Type is the event type: "content_delta", "thinking_delta", "tool_use", "usage", "heartbeat", "error", "done".
	Type string `json:"type"`

	// EventType is an alias for Type (sdk-graph canonical name).
	EventType string `json:"event_type,omitempty"`

	// Delta contains the incremental text for content_delta and thinking_delta events.
	Delta *StreamDelta `json:"delta,omitempty"`

	// ToolUse is populated for tool_use events.
	ToolUse *StreamToolUse `json:"tool_use,omitempty"`

	// Usage is populated for usage events.
	Usage *ChatUsage `json:"usage,omitempty"`

	// Error is populated for error events.
	Error string `json:"error,omitempty"`

	// Done is true when the stream is complete.
	Done bool `json:"done,omitempty"`
}

// StreamDelta contains the incremental text in a streaming event.
type StreamDelta struct {
	Text string `json:"text"`
}

// StreamToolUse contains a tool call from a streaming event.
type StreamToolUse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// rawStreamEvent is the raw JSON from the SSE stream before parsing into typed fields.
type rawStreamEvent struct {
	Type         string         `json:"type"`
	Delta        *StreamDelta   `json:"delta,omitempty"`
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Input        map[string]any `json:"input,omitempty"`
	InputTokens  int            `json:"input_tokens,omitempty"`
	OutputTokens int            `json:"output_tokens,omitempty"`
	CostTicks    int64          `json:"cost_ticks,omitempty"`
	Message      string         `json:"message,omitempty"`
}

// ChatStream sends a streaming text generation request and returns a channel of events.
//
// The caller should range over the returned channel. The last event will have Done=true.
// Cancel the context to abort the stream early.
//
//	events, err := client.ChatStream(ctx, &qai.ChatRequest{
//	    Model:    "claude-sonnet-4-6",
//	    Messages: []qai.ChatMessage{{Role: "user", Content: "Hello!"}},
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for ev := range events {
//	    if ev.Error != "" {
//	        log.Fatal(ev.Error)
//	    }
//	    if ev.Delta != nil {
//	        fmt.Print(ev.Delta.Text)
//	    }
//	}
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	req.Stream = true

	resp, _, err := c.doStreamRaw(ctx, "/qai/v1/chat", req)
	if err != nil {
		return nil, err
	}

	ch := make(chan StreamEvent, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		// Allow up to 1MB per SSE line (tool call inputs can be large).
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			// SSE lines are "data: <json>" or "data: [DONE]".
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")

			if payload == "[DONE]" {
				select {
				case ch <- StreamEvent{Type: "done", Done: true}:
				case <-ctx.Done():
				}
				return
			}

			var raw rawStreamEvent
			if err := json.Unmarshal([]byte(payload), &raw); err != nil {
				select {
				case ch <- StreamEvent{Type: "error", Error: fmt.Sprintf("parse SSE: %v", err)}:
				case <-ctx.Done():
				}
				return
			}

			ev := StreamEvent{Type: raw.Type}

			switch raw.Type {
			case "content_delta", "thinking_delta":
				ev.Delta = raw.Delta
			case "tool_use":
				ev.ToolUse = &StreamToolUse{
					ID:    raw.ID,
					Name:  raw.Name,
					Input: raw.Input,
				}
			case "usage":
				ev.Usage = &ChatUsage{
					InputTokens:  raw.InputTokens,
					OutputTokens: raw.OutputTokens,
					CostTicks:    raw.CostTicks,
				}
			case "error":
				ev.Error = raw.Message
			case "heartbeat":
				// pass through
			}

			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
