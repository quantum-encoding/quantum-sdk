// Example: basic chat request with the Quantum AI SDK.
//
// Usage:
//
//	export QAI_API_KEY=your-api-key
//	go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	qai "github.com/quantum-encoding/quantum-sdk"
)

func main() {
	apiKey := os.Getenv("QAI_API_KEY")
	if apiKey == "" {
		log.Fatal("QAI_API_KEY environment variable is required")
	}

	client := qai.New(apiKey)
	ctx := context.Background()

	// --- Non-streaming example ---
	fmt.Println("=== Non-streaming Chat ===")

	resp, err := client.Chat(ctx, &qai.ChatRequest{
		Model: "claude-sonnet-4-6",
		Messages: []qai.ChatMessage{
			{Role: "user", Content: "What is quantum computing in one sentence?"},
		},
	})
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Response: %s\n", resp.Text())
	if resp.Usage != nil {
		fmt.Printf("Tokens: %d in / %d out (cost: %d ticks)\n",
			resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.CostTicks)
	}
	fmt.Printf("Request ID: %s\n\n", resp.RequestID)

	// --- Streaming example ---
	fmt.Println("=== Streaming Chat ===")

	events, err := client.ChatStream(ctx, &qai.ChatRequest{
		Model: "claude-sonnet-4-6",
		Messages: []qai.ChatMessage{
			{Role: "user", Content: "Count from 1 to 5, one number per line."},
		},
	})
	if err != nil {
		log.Fatalf("ChatStream failed: %v", err)
	}

	for ev := range events {
		switch ev.Type {
		case "content_delta":
			if ev.Delta != nil {
				fmt.Print(ev.Delta.Text)
			}
		case "usage":
			if ev.Usage != nil {
				fmt.Printf("\n[Cost: %d ticks]\n", ev.Usage.CostTicks)
			}
		case "error":
			log.Fatalf("Stream error: %s", ev.Error)
		case "done":
			fmt.Println("\n[Stream complete]")
		}
	}
}
