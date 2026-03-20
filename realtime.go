package qai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// RealtimeConfig configures a realtime voice session.
type RealtimeConfig struct {
	// Voice to use (e.g. "Sal", "Eve", "Vesper"). Default: "Sal".
	Voice string `json:"voice,omitempty"`
	// System instructions for the AI.
	Instructions string `json:"instructions,omitempty"`
	// PCM sample rate in Hz. Default: 24000.
	SampleRate int `json:"sample_rate,omitempty"`
	// Tool definitions (xAI Realtime API format).
	Tools []json.RawMessage `json:"tools,omitempty"`
}

// RealtimeEvent is a parsed incoming event from the realtime API.
type RealtimeEvent struct {
	Type string `json:"type"`

	// AudioDelta
	Delta string `json:"delta,omitempty"`
	// TranscriptDelta / TranscriptDone
	Transcript string `json:"transcript,omitempty"`
	Source     string `json:"source,omitempty"` // "input" or "output"
	// FunctionCall
	Name      string `json:"name,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	// Error
	Message string `json:"message,omitempty"`
	// Unknown
	Raw json.RawMessage `json:"raw,omitempty"`
}

// Realtime event type constants.
const (
	EventSessionReady   = "session_ready"
	EventAudioDelta     = "audio_delta"
	EventTranscriptDelta = "transcript_delta"
	EventTranscriptDone = "transcript_done"
	EventSpeechStarted  = "speech_started"
	EventSpeechStopped  = "speech_stopped"
	EventFunctionCall   = "function_call"
	EventResponseDone   = "response_done"
	EventError          = "error"
	EventUnknown        = "unknown"
)

// RealtimeSender is the write half of a realtime session.
type RealtimeSender struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// RealtimeReceiver is the read half of a realtime session.
type RealtimeReceiver struct {
	conn *websocket.Conn
}

// RealtimeSession is the response from POST /qai/v1/realtime/session.
type RealtimeSession struct {
	EphemeralToken string `json:"ephemeral_token"`
	URL            string `json:"url"`
	SessionID      string `json:"session_id"`
}

// RealtimeSession requests an ephemeral token from the QAI proxy for direct xAI voice connection.
// Call this before RealtimeConnectDirect to get a scoped token.
func (c *Client) RealtimeSession(ctx context.Context) (*RealtimeSession, error) {
	var resp RealtimeSession
	_, err := c.doJSON(ctx, "POST", "/qai/v1/realtime/session", map[string]any{}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// RealtimeEnd finalizes billing for a direct voice session.
// Call after disconnecting from the direct xAI WebSocket.
func (c *Client) RealtimeEnd(ctx context.Context, sessionID string, durationSeconds int) error {
	var resp json.RawMessage
	_, err := c.doJSON(ctx, "POST", "/qai/v1/realtime/end", map[string]any{
		"session_id":       sessionID,
		"duration_seconds": durationSeconds,
	}, &resp)
	return err
}

// RealtimeRefresh refreshes an ephemeral token for long sessions (>4 min).
// Returns a new ephemeral token string.
func (c *Client) RealtimeRefresh(ctx context.Context, sessionID string) (string, error) {
	var resp struct {
		EphemeralToken string `json:"ephemeral_token"`
	}
	_, err := c.doJSON(ctx, "POST", "/qai/v1/realtime/refresh", map[string]any{
		"session_id": sessionID,
	}, &resp)
	if err != nil {
		return "", err
	}
	return resp.EphemeralToken, nil
}

// RealtimeConnectDirect connects directly to xAI's realtime API with an ephemeral token.
// Much lower latency than the proxy path. Use RealtimeSession() first to get the token.
func (c *Client) RealtimeConnectDirect(ctx context.Context, ephemeralToken string, config *RealtimeConfig) (*RealtimeSender, *RealtimeReceiver, error) {
	return RealtimeConnectDirectTo(ctx, "wss://api.x.ai/v1/realtime", ephemeralToken, config)
}

// RealtimeConnectDirectTo connects to a specific WebSocket URL with a token.
func RealtimeConnectDirectTo(ctx context.Context, wsURL, token string, config *RealtimeConfig) (*RealtimeSender, *RealtimeReceiver, error) {
	if config == nil {
		config = &RealtimeConfig{}
	}
	if config.Voice == "" {
		config.Voice = "Sal"
	}
	if config.SampleRate == 0 {
		config.SampleRate = 24000
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(connectCtx, wsURL, &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("qai: realtime direct connect: %w", err)
	}

	conn.SetReadLimit(16 * 1024 * 1024)

	sender := &RealtimeSender{conn: conn}
	receiver := &RealtimeReceiver{conn: conn}

	if err := sendSessionUpdate(ctx, sender, config); err != nil {
		conn.Close(websocket.StatusNormalClosure, "")
		return nil, nil, err
	}

	return sender, receiver, nil
}

// RealtimeConnect opens a realtime voice session via the QAI proxy WebSocket.
// Returns (sender, receiver) for bidirectional communication.
func (c *Client) RealtimeConnect(ctx context.Context, config *RealtimeConfig) (*RealtimeSender, *RealtimeReceiver, error) {
	if config == nil {
		config = &RealtimeConfig{}
	}
	if config.Voice == "" {
		config.Voice = "Sal"
	}
	if config.SampleRate == 0 {
		config.SampleRate = 24000
	}

	// Convert https:// → wss://, http:// → ws://
	wsBase := strings.Replace(c.baseURL, "https://", "wss://", 1)
	wsBase = strings.Replace(wsBase, "http://", "ws://", 1)
	url := wsBase + "/qai/v1/realtime"

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+c.apiKey)
	headers.Set("X-API-Key", c.apiKey)

	connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(connectCtx, url, &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("qai: realtime connect: %w", err)
	}

	// Increase read limit for audio data
	conn.SetReadLimit(16 * 1024 * 1024) // 16 MB

	sender := &RealtimeSender{conn: conn}
	receiver := &RealtimeReceiver{conn: conn}

	if err := sendSessionUpdate(ctx, sender, config); err != nil {
		conn.Close(websocket.StatusNormalClosure, "")
		return nil, nil, err
	}

	return sender, receiver, nil
}

func sendSessionUpdate(ctx context.Context, sender *RealtimeSender, config *RealtimeConfig) error {
	tools := config.Tools
	if tools == nil {
		tools = []json.RawMessage{}
	}
	return sender.sendJSON(ctx, map[string]any{
		"type": "session.update",
		"session": map[string]any{
			"voice":              config.Voice,
			"instructions":       config.Instructions,
			"input_audio_format": "pcm16",
			"output_audio_format": "pcm16",
			"input_audio_transcription": map[string]any{
				"model": "grok-2-audio",
			},
			"turn_detection": map[string]any{
				"type": "server_vad",
			},
			"tools": tools,
		},
	})
}

// SendAudio sends a base64-encoded PCM audio chunk.
func (s *RealtimeSender) SendAudio(ctx context.Context, base64PCM string) error {
	return s.sendJSON(ctx, map[string]any{
		"type":  "input_audio_buffer.append",
		"audio": base64PCM,
	})
}

// SendText sends a text message and requests a response.
func (s *RealtimeSender) SendText(ctx context.Context, text string) error {
	if err := s.sendJSON(ctx, map[string]any{
		"type": "conversation.item.create",
		"item": map[string]any{
			"type": "message",
			"role": "user",
			"content": []map[string]any{
				{"type": "input_text", "text": text},
			},
		},
	}); err != nil {
		return err
	}
	return s.sendJSON(ctx, map[string]any{
		"type":     "response.create",
		"response": map[string]any{"modalities": []string{"text", "audio"}},
	})
}

// SendFunctionResult sends a tool call result back to the model.
func (s *RealtimeSender) SendFunctionResult(ctx context.Context, callID, output string) error {
	if err := s.sendJSON(ctx, map[string]any{
		"type": "conversation.item.create",
		"item": map[string]any{
			"type":    "function_call_output",
			"call_id": callID,
			"output":  output,
		},
	}); err != nil {
		return err
	}
	return s.sendJSON(ctx, map[string]any{"type": "response.create"})
}

// CancelResponse cancels the current response (interrupt).
func (s *RealtimeSender) CancelResponse(ctx context.Context) error {
	return s.sendJSON(ctx, map[string]any{"type": "response.cancel"})
}

// Close closes the WebSocket connection gracefully.
func (s *RealtimeSender) Close() error {
	return s.conn.Close(websocket.StatusNormalClosure, "")
}

func (s *RealtimeSender) sendJSON(ctx context.Context, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("qai: realtime marshal: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.Write(ctx, websocket.MessageText, data)
}

// Recv receives the next event. Returns an error when the connection closes.
func (r *RealtimeReceiver) Recv(ctx context.Context) (*RealtimeEvent, error) {
	for {
		_, data, err := r.conn.Read(ctx)
		if err != nil {
			return nil, err
		}

		event := parseRealtimeEvent(data)
		return event, nil
	}
}

func parseRealtimeEvent(data []byte) *RealtimeEvent {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return &RealtimeEvent{Type: EventUnknown, Raw: data}
	}

	var eventType string
	if t, ok := raw["type"]; ok {
		json.Unmarshal(t, &eventType)
	}

	str := func(key string) string {
		if v, ok := raw[key]; ok {
			var s string
			json.Unmarshal(v, &s)
			return s
		}
		return ""
	}

	switch eventType {
	case "session.updated":
		return &RealtimeEvent{Type: EventSessionReady}

	case "response.audio.delta", "response.output_audio.delta":
		return &RealtimeEvent{Type: EventAudioDelta, Delta: str("delta")}

	case "response.audio_transcript.delta", "response.output_audio_transcript.delta":
		return &RealtimeEvent{Type: EventTranscriptDelta, Delta: str("delta"), Source: "output"}

	case "response.audio_transcript.done", "response.output_audio_transcript.done":
		return &RealtimeEvent{Type: EventTranscriptDone, Transcript: str("transcript"), Source: "output"}

	case "conversation.item.input_audio_transcription.completed":
		return &RealtimeEvent{Type: EventTranscriptDone, Transcript: str("transcript"), Source: "input"}

	case "input_audio_buffer.speech_started":
		return &RealtimeEvent{Type: EventSpeechStarted}

	case "input_audio_buffer.speech_stopped":
		return &RealtimeEvent{Type: EventSpeechStopped}

	case "response.function_call_arguments.done":
		return &RealtimeEvent{
			Type:      EventFunctionCall,
			Name:      str("name"),
			CallID:    str("call_id"),
			Arguments: str("arguments"),
		}

	case "response.done":
		return &RealtimeEvent{Type: EventResponseDone}

	case "error":
		// Parse nested error.message
		msg := ""
		if errRaw, ok := raw["error"]; ok {
			var errObj map[string]json.RawMessage
			if json.Unmarshal(errRaw, &errObj) == nil {
				if m, ok := errObj["message"]; ok {
					json.Unmarshal(m, &msg)
				}
			}
		}
		if msg == "" {
			msg = str("message")
		}
		if msg == "" {
			msg = "unknown error"
		}
		return &RealtimeEvent{Type: EventError, Message: msg}

	default:
		return &RealtimeEvent{Type: EventUnknown, Raw: data}
	}
}
