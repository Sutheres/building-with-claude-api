package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

// chat sends a single-turn request, optionally with a system prompt.
// The system prompt sets instructions or a persona that persist for the
// whole conversation; it is passed via the top-level System field rather
// than as a message with a "system" role.
func chat(ctx context.Context, client anthropic.Client, system string, prompt string) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}
	if system != "" {
		params.System = []anthropic.TextBlockParam{{Text: system}}
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}

	var out string
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			out += text.Text
		}
	}
	return out, nil
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()
	ctx := context.Background()

	// This system prompt tightly constrains the output: code only, no prose.
	system := `You are a helpful assistant that writes Python code.
Your responses should be concise and to the point.
You should not include any other text than the code.
You should not include any explanations.`

	answer, err := chat(ctx, client, system,
		"Write a Python function that checks a string for duplicate characters.")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}
