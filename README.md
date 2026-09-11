# quantum-sdk

Go client SDK for the [Quantum AI API](https://api.quantumencoding.ai).

```bash
go get github.com/quantum-encoding/quantum-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    qai "github.com/quantum-encoding/quantum-sdk"
)

func main() {
    client := qai.NewClient("qai_k_your_key_here")
    resp, err := client.Chat(context.Background(), "gemini-2.5-flash", "Hello! What is quantum computing?")
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Text())
}
```

## Features

- 110+ endpoints across 10 AI providers and 45+ models
- `context.Context` on every call for cancellation and timeouts
- Streaming via callback or channel-based APIs
- No external dependencies beyond the standard library
- Strongly typed request/response structs
- Agent orchestration with SSE event streams
- GPU/CPU compute rental (requires per-account admin approval)
- Batch processing (50% discount)

## Examples

### Chat Completion

```go
client := qai.NewClient("qai_k_your_key_here")

resp, err := client.ChatRequest(ctx, &qai.ChatRequest{
    Model: "claude-opus-4-8",
    Messages: []qai.ChatMessage{
        qai.SystemMessage("You are a helpful assistant."),
        qai.UserMessage("Explain goroutines in Go"),
    },
    Temperature: qai.Float64(0.7),
    MaxTokens:   qai.Int(1000),
})
if err != nil {
    return err
}
fmt.Println(resp.Text())
```

### Streaming

```go
err := client.ChatStream(ctx, &qai.ChatRequest{
    Model:    "claude-opus-4-8",
    Messages: []qai.ChatMessage{qai.UserMessage("Write a haiku about Go")},
}, func(event *qai.StreamEvent) {
    if event.DeltaText() != "" {
        fmt.Print(event.DeltaText())
    }
})
```

### Reasoning state across a tool loop

Reasoning models on the OpenAI and xAI lanes mint a `reasoning` content block
alongside their `tool_use` blocks. It is the provider's own state, opaque, and
it must go back **unchanged and in the same position** on the next turn's
assistant message — its place among the tool calls is how the provider learns
where the reasoning sat. Drop it and the reasoning tokens are re-billed on
every round of the loop.

The simplest correct thing is to hand the whole `Content` slice back:

```go
messages := []qai.ChatMessage{qai.UserMessage("What is the weather in Oslo?")}

resp, err := client.ChatRequest(ctx, &qai.ChatRequest{
    Model:    "gpt-5.6",
    Messages: messages,
    // One key per conversation, reused on every turn, so all of them land
    // on the same warm provider cache shard.
    PromptCacheKey: "conv-7f3a",
})
if err != nil {
    return err
}

// Verbatim, in order: reasoning blocks, tool_use blocks, text blocks.
messages = append(messages, qai.ChatMessage{
    Role:          "assistant",
    ContentBlocks: resp.Content,
})
// ... then append one tool-result message per tool_use block and loop.
```

`ContentBlock.Reasoning` is `json.RawMessage` on purpose: it re-encodes the
bytes that arrived. A `thinking` block is the human-readable summary of the
same turn — render that one, replay this one.

On Gemini 3 the equivalent state is `ContentBlock.ThoughtSignature`, and it now
rides the **text** block of a turn that ended in text as well as the `tool_use`
blocks. Streaming delivers it as a `thought_signature` event just before
`done`, on `StreamEvent.ThoughtSignature`.

### Provider options

`ProviderOptions` is an open map keyed by provider, so a key the gateway
documents but this SDK version does not name still rides through:

```go
resp, err := client.ChatRequest(ctx, &qai.ChatRequest{
    Model:    "gpt-5.6",
    Messages: messages,
    ProviderOptions: map[string]any{
        "openai": map[string]any{
            "reasoning_summary": "detailed",   // auto | concise | detailed | none
            "reasoning_mode":    "pro",        // standard | pro
            "verbosity":         "low",        // low | medium | high
            "text_format":       "json_object", // text | json_object
        },
        "xai": map[string]any{"native_files": true},
    },
})
```

`ReasoningEffort` accepts `none`, `low`, `medium`, `high`, `xhigh` and `max` on
every lane; each adapter folds a tier its model lacks onto the nearest one.

### Image Generation

```go
images, err := client.GenerateImage(ctx, "grok-imagine-image", "A cosmic duck in space")
if err != nil {
    return err
}
for _, img := range images.Images {
    fmt.Println(img.URL)
}
```

### Text-to-Speech

```go
audio, err := client.Speak(ctx, &qai.TTSRequest{
    Model:        "gpt-4o-mini-tts",
    Text:         "Welcome to Quantum AI!",
    Voice:        "alloy",
    OutputFormat: "mp3",
})
if err != nil {
    return err
}
// The audio arrives inline, base64-encoded — there is no URL to fetch.
fmt.Printf("%d bytes of %s\n", audio.SizeBytes, audio.Format)
```

### Steering a Gemini voice

The gateway's house voice is **Gemini 3.1 Flash TTS**
(`gemini-3.1-flash-tts-preview`) with the **Laomedeia** voice; both apply when
the request names neither, so `Text` alone is a complete request.

Gemini has no knobs for tone, accent or pace. You steer it in prose — with
`Instructions` for the whole read, and with inline tags inside `Text` for
moment-to-moment inflection.

```go
audio, err := client.Speak(ctx, &qai.TTSRequest{
    // No model: the gateway supplies Gemini 3.1 Flash TTS + Laomedeia.
    Text: "Hi, this is Lacey from CRG Direct. [warmly] How can I help today?",
    Instructions: "Read aloud as a friendly, professional customer-service " +
        "assistant with a natural British accent, at a natural easy pace",
    Language: "en-GB",
})
if err != nil {
    return err
}
fmt.Printf("%d bytes of %s\n", audio.SizeBytes, audio.Format)
```

`Instructions` carries tone and character ("like telling a friend about
something you love"), accent ("with a natural British accent" — pair it with
`Language` so the pronunciation family matches), and pace ("slow down on the
phone number"). Spell digits with separators — `0-1-2-3, 4-5-6` — to have them
read one at a time.

**Inline tags** go in the text itself: `[amazed] [crying] [curious] [excited]
[sighs] [gasp] [giggles] [laughs] [mischievously] [panicked] [sarcastic]
[serious] [shouting] [tired] [trembling] [whispers]`, plus free-form ones like
`[like a cartoon dog]`.

**Two-speaker dialogue** replaces `Voice` with `Speakers`. Exactly two — the
gateway rejects any other count with a 400 — and the text carries each
speaker's lines under the matching label:

```go
audio, err := client.Speak(ctx, &qai.TTSRequest{
    Text: "Lacey: Hi, this is Lacey from CRG Direct. How can I help?\n" +
        "Customer: [excited] Hi! I'm calling about Tuesday's installation.",
    Instructions: "Lacey is calm and professional; the customer is cheerful",
    Speakers: []qai.TTSSpeaker{
        {Name: "Lacey", Voice: "Laomedeia"},
        {Name: "Customer", Voice: "Puck"},
    },
})
```

All 30 Gemini prebuilt voices (Zephyr, Puck, Charon, Kore, Laomedeia,
Sulafat, …) work on every Gemini TTS model. `client.ListVoices(ctx)` returns
the catalogue with each voice's `Provider` and the `Model` to pass back for it,
so a picker never hardcodes the provider-to-model mapping.

Limits: 32k-token session context, two speakers maximum, and quality drifts
past a few minutes of audio — split long scripts.

`Speed`, `SampleRate` and `BitRate` are xAI-only; `VoiceSettings` is
ElevenLabs-only. On Gemini, ask for pace in `Instructions` instead.

### Web Search

```go
results, err := client.WebSearch(ctx, "latest Go releases 2026")
if err != nil {
    return err
}
for _, r := range results.Results {
    fmt.Printf("%s: %s\n", r.Title, r.URL)
}
```

### Agent Orchestration

```go
err := client.AgentRun(ctx, "Research quantum computing breakthroughs",
    func(event *qai.AgentEvent) {
        switch event.Type {
        case "content_delta":
            fmt.Print(event.Content)
        case "done":
            fmt.Println("\n--- Done ---")
        }
    },
)
```

## All Endpoints

| Category | Endpoints | Description |
|----------|-----------|-------------|
| Chat | 2 | Text generation + session chat |
| Agent | 2 | Multi-step orchestration + missions |
| Images | 2 | Generation + editing |
| Video | 7 | Generation, studio, translation, avatars |
| Audio | 13 | TTS, STT, music, dialogue, dubbing, voice design |
| Voices | 5 | Clone, list, delete, library, design |
| Embeddings | 1 | Text embeddings |
| RAG | 4 | Vertex AI + SurrealDB search |
| Documents | 3 | Extract, chunk, process |
| Search | 3 | Web search, context, answers |
| Scanner | 11 | Code scanning, type queries, diffs |
| Scraper | 2 | Doc scraping + screenshots |
| Jobs | 3 | Async job management |
| Compute | 7 | GPU/CPU rental (admin-approved accounts only) |
| Keys | 3 | API key management |
| Account | 3 | Balance, usage, summary |
| Credits | 6 | Packs, tiers, lifetime, purchase |
| Batch | 4 | 50% discount batch processing |
| Realtime | 3 | Voice sessions |
| Models | 2 | Model list + pricing |

## Authentication

Pass your API key when creating the client:

```go
client := qai.NewClient("qai_k_your_key_here")
```

The SDK sends it as the `X-API-Key` header. Both `qai_...` (primary) and `qai_k_...` (scoped) keys are supported. You can also use `Authorization: Bearer <key>`.

Get your API key at [cosmicduck.dev](https://cosmicduck.dev).

## Pricing

See [api.quantumencoding.ai/pricing](https://api.quantumencoding.ai/pricing) for current rates.

The **Lifetime tier** offers 0% margin at-cost pricing via a one-time payment.

## Other SDKs

All SDKs are at v0.4.0 with type parity verified by scanner.

| Language | Package | Install |
|----------|---------|---------|
| Rust | quantum-sdk | `cargo add quantum-sdk` |
| **Go** | quantum-sdk | `go get github.com/quantum-encoding/quantum-sdk` |
| TypeScript | @quantum-encoding/quantum-sdk | `npm i @quantum-encoding/quantum-sdk` |
| Python | quantum-sdk | `pip install quantum-sdk` |
| Swift | QuantumSDK | Swift Package Manager |
| Kotlin | quantum-sdk | Gradle dependency |

MCP server: `npx @quantum-encoding/ai-conductor-mcp`

## API Reference

- Interactive docs: [api.quantumencoding.ai/docs](https://api.quantumencoding.ai/docs)
- OpenAPI spec: [api.quantumencoding.ai/openapi.yaml](https://api.quantumencoding.ai/openapi.yaml)
- LLM context: [api.quantumencoding.ai/llms.txt](https://api.quantumencoding.ai/llms.txt)

## License

MIT
