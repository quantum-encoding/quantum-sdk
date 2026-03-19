package qai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// readSSEStream reads SSE events from an HTTP response and sends them on a channel.
// It handles "data: <json>" lines and "data: [DONE]" termination.
// The caller is responsible for ranging over the returned channel.
func readSSEStream(ctx context.Context, resp *http.Response) <-chan StreamEvent {
	ch := make(chan StreamEvent, 64)

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

	return ch
}
