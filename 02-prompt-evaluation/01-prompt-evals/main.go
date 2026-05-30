package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

type chatOption func(*anthropic.MessageNewParams)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()
	ctx := context.Background()

	dataset, err := generateDataset(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(dataset)

	data, err := json.MarshalIndent(dataset, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("dataset.json", data, 0o644); err != nil {
		log.Fatal(err)
	}
}

func addUserMessage(messages []anthropic.MessageParam, msg string) []anthropic.MessageParam {
	return append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg)))
}

func addAssistantMessage(messages []anthropic.MessageParam, msg string) []anthropic.MessageParam {
	return append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg)))
}

func withSystemPrompt(prompt string) chatOption {
	return func(params *anthropic.MessageNewParams) {
		params.System = []anthropic.TextBlockParam{{Text: prompt}}
	}
}

func withStopSequences(sequences []string) chatOption {
	return func(params *anthropic.MessageNewParams) {
		params.StopSequences = sequences
	}
}

func withTemperature(temp float64) chatOption {
	return func(params *anthropic.MessageNewParams) {
		params.Temperature = anthropic.Float(temp)
	}
}

func chat(ctx context.Context, client anthropic.Client, messages []anthropic.MessageParam, opts ...chatOption) (string, error) {
	params := anthropic.MessageNewParams{
		Model:       anthropic.ModelClaudeHaiku4_5,
		MaxTokens:   1000,
		Messages:    messages,
		Temperature: anthropic.Float(1.0),
	}

	for _, opt := range opts {
		opt(&params)
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}

	if block, ok := resp.Content[0].AsAny().(anthropic.TextBlock); ok {
		return block.Text, nil
	}

	return "", errors.New("no text blocks returned from anthropic")
}

type DataItem struct {
	Task string `json:"task"`
}

func generateDataset(ctx context.Context, client anthropic.Client) ([]DataItem, error) {
	prompt := `
Generate an evaluation dataset ...

Example output:

` + "```json" + `
[
  {
    "task": "Description of task",
  },
  ...additional
]
` + "```" + `

* Focus on tasks that can be solved by writing a single Go function, a single JSON object, or a single regex
* Focus on tasks that do not require writing much code

Please generate 3 objects.
`

	var messages []anthropic.MessageParam
	var dataset []DataItem

	messages = addUserMessage(messages, prompt)
	messages = addAssistantMessage(messages, "```json")

	text, err := chat(ctx, client, messages,
		withStopSequences([]string{"```"}),
	)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(text), &dataset); err != nil {
		return nil, err
	}

	return dataset, nil
}
