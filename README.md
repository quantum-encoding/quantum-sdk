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
- GPU/CPU compute rental
- Batch processing (50% discount)

## Examples

### Chat Completion

```go
client := qai.NewClient("qai_k_your_key_here")

resp, err := client.ChatRequest(ctx, &qai.ChatRequest{
    Model: "claude-sonnet-4-6",
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
    Model:    "claude-sonnet-4-6",
    Messages: []qai.ChatMessage{qai.UserMessage("Write a haiku about Go")},
}, func(event *qai.StreamEvent) {
    if event.DeltaText() != "" {
        fmt.Print(event.DeltaText())
    }
})
```

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
audio, err := client.Speak(ctx, "Welcome to Quantum AI!", "alloy", "mp3")
if err != nil {
    return err
}
fmt.Println(audio.AudioURL)
```

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
| Compute | 7 | GPU/CPU rental |
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
