package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

type testCase struct {
	Task string `json:"task"`
}

type testOutput struct {
	Output   string   `json:"output,omitempty"`
	Testcase testCase `json:"testcase"`
	Score    int      `json:"score,omitempty"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client := anthropic.NewClient()

	data, err := os.ReadFile("dataset.json")
	if err != nil {
		log.Fatal(err)
	}

	var dataset []testCase
	if err = json.Unmarshal(data, &dataset); err != nil {
		log.Fatal(err)
	}

	results, err := runEval(ctx, client, dataset)
	if err != nil {
		log.Fatal(err)
	}

	resultData, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(resultData))
}

func addUserMessage(messages []anthropic.MessageParam, msg string) []anthropic.MessageParam {
	return append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg)))
}

func chat(ctx context.Context, client anthropic.Client, messages []anthropic.MessageParam) (string, error) {
	params := anthropic.MessageNewParams{
		Model:       anthropic.ModelClaudeHaiku4_5,
		MaxTokens:   1000,
		Messages:    messages,
		Temperature: anthropic.Float(1.0),
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

// runPrompt Merges the prompt and test case input, then returns the result
func runPrompt(ctx context.Context, client anthropic.Client, tc testCase) (string, error) {
	prompt := fmt.Sprintf(`
Please solve the following task:

%s
`, tc.Task)

	var messages []anthropic.MessageParam

	messages = addUserMessage(messages, prompt)

	output, err := chat(ctx, client, messages)
	if err != nil {
		return "", err
	}

	return output, nil
}

// runTestCase calls runPrompt, then grades the result
func runTestCase(ctx context.Context, client anthropic.Client, tc testCase) (testOutput, error) {
	output, err := runPrompt(ctx, client, tc)
	if err != nil {
		return testOutput{}, err
	}

	score := 10

	return testOutput{
		Output:   output,
		Testcase: tc,
		Score:    score,
	}, nil
}

// runEval takes the dataset and calls runTestCase with each case
func runEval(ctx context.Context, client anthropic.Client, dataset []testCase) ([]testOutput, error) {
	var results []testOutput

	for _, tc := range dataset {
		result, err := runTestCase(ctx, client, tc)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}
