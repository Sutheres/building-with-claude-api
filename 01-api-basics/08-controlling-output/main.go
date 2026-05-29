package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()

	// Combining prefilling with a stop sequence is a lightweight way to get
	// clean structured output without tools. We prefill the assistant turn
	// with an opening ```json fence so Claude starts emitting JSON straight
	// away, and stop it at the closing ``` fence. What's left between them
	// is raw JSON we can parse directly.
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:         anthropic.ModelClaudeSonnet4_6,
		MaxTokens:     1024,
		StopSequences: []string{"```"},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Generate a very short EventBridge rule as json")),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("```json")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	var raw string
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			raw += text.Text
		}
	}
	raw = strings.TrimSpace(raw)

	fmt.Println("=== raw output ===")
	fmt.Println(raw)

	// Because we forced the fences, the output is parseable JSON.
	var rule map[string]any
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		log.Fatalf("output was not valid JSON: %v", err)
	}

	fmt.Println("\n=== parsed keys ===")
	for k := range rule {
		fmt.Printf("- %s\n", k)
	}
}
