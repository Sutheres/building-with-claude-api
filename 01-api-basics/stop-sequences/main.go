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

	// Stop sequences cut generation off as soon as one of the given strings
	// is produced. The string itself is not included in the output. Here we
	// ask Claude to count to 10 but stop the moment it reaches ", 5".
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:         anthropic.ModelClaudeSonnet4_6,
		MaxTokens:     1024,
		StopSequences: []string{", 5"},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Count from 1 to 10")),
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

	// stop_reason tells you why generation ended; "stop_sequence" means a
	// stop string was hit, and stop_sequence reports which one.
	fmt.Printf("\nstop_reason:   %s\n", resp.StopReason)
	fmt.Printf("stop_sequence: %q\n", resp.StopSequence)
}
