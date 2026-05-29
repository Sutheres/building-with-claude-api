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

	// Prefilling: seed the start of the assistant's turn by adding an
	// assistant message yourself. Claude continues from where the prefill
	// leaves off, which lets you steer tone, format, or commit it to an
	// answer. Here we force the response to argue for coffee.
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Is tea or coffee better at breakfast?")),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("Coffee is better because")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// The response continues the prefill, so we print the prefill plus the
	// generated continuation to read as one sentence.
	fmt.Print("Coffee is better because")
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Print(text.Text)
		}
	}
	fmt.Println()
}
