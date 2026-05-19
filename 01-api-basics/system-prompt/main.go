package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()

	// System prompts give Claude a persona or set of instructions that persist
	// across the entire conversation.

	// Example 1: pirate persona
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 256,
		System: []anthropic.TextBlockParam{
			{Text: "You are a helpful assistant who speaks like a pirate. Keep responses brief."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("What's the weather like today?")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Pirate persona ===")
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Println(text.Text)
		}
	}

	// Example 2: temperature controls creativity/randomness.
	// 0.0 = deterministic, 1.0 = very creative. Default is ~1.0.
	resp2, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 256,
		Temperature: anthropic.Float(0.0),
		System: []anthropic.TextBlockParam{
			{Text: "You are a concise assistant. Answer in exactly one sentence."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Name a random color.")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Low temperature (deterministic) ===")
	for _, block := range resp2.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Println(text.Text)
		}
	}
}
