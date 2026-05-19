package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()

	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("What is the capital of France? Answer in one sentence.")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Println(text.Text)
		}
	}

	fmt.Printf("\n--- usage ---\ninput tokens:  %d\noutput tokens: %d\n",
		resp.Usage.InputTokens, resp.Usage.OutputTokens)
}
