package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()

	fmt.Println("Streaming response:")
	fmt.Println("-------------------")

	stream := client.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"Write a short poem about Go programming language.",
			)),
		},
	})

	// Print tokens as they arrive
	for stream.Next() {
		event := stream.Current()
		if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if text, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok {
				fmt.Print(text.Text)
			}
		}
	}
	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}

	// Accumulate the final message for usage stats
	stream2 := client.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Say hello in three words.")),
		},
	})

	msg := anthropic.Message{}
	for stream2.Next() {
		msg.Accumulate(stream2.Current())
	}
	if err := stream2.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n\n--- accumulated message usage ---\ninput:  %d tokens\noutput: %d tokens\n",
		msg.Usage.InputTokens, msg.Usage.OutputTokens)
}
