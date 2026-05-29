package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

// model is intentionally a small, fast model: dataset generation and the
// eval runs that follow make many calls, so cost and latency matter more
// than peak quality here.
const model = anthropic.ModelClaudeHaiku4_5

// testCase is one row of an evaluation dataset: a task to solve, the
// expected output format, and the criteria a good solution must meet.
type testCase struct {
	Task             string `json:"task"`
	Format           string `json:"format"`
	SolutionCriteria string `json:"solution_criteria"`
}

const datasetPrompt = `Generate an evaluation dataset for testing prompts that generate Go, JSON, or
Regex specifically for AWS-related tasks. Generate an array of JSON objects,
each representing a task that requires Go, JSON, or a Regex to complete.

Example output:
[
    {
        "task": "Description of task",
        "format": "json" | "go" | "regex",
        "solution_criteria": "Criteria for the solution to be correct"
    }
]

* Focus on tasks solvable by a single Go function, a single JSON object, or one regular expression.
* Focus on tasks that do not require writing much code.

Please generate 3 objects.`

func main() {
	out := flag.String("out", "02-prompt-evaluation/dataset.json", "path to write the generated dataset")
	flag.Parse()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()

	dataset, err := generateDataset(context.Background(), client)
	if err != nil {
		log.Fatal(err)
	}

	// Marshal with indentation so the dataset stays human-readable/diffable.
	data, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wrote %d test cases to %s\n", len(dataset), *out)
}

// generateDataset asks Claude for a fresh dataset. We prefill the assistant
// turn with a "```json" fence and stop at the closing fence so the response
// is bare JSON we can unmarshal directly (the technique from module 01's
// controlling-output example).
func generateDataset(ctx context.Context, client anthropic.Client) ([]testCase, error) {
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:         model,
		MaxTokens:     1000,
		StopSequences: []string{"```"},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(datasetPrompt)),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("```json")),
		},
	})
	if err != nil {
		return nil, err
	}

	var raw string
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			raw += text.Text
		}
	}

	var dataset []testCase
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &dataset); err != nil {
		return nil, fmt.Errorf("model did not return valid JSON: %w", err)
	}
	return dataset, nil
}
