package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

type MovieReview struct {
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	Rating   float64  `json:"rating"`
	Summary  string   `json:"summary"`
	Pros     []string `json:"pros"`
	Cons     []string `json:"cons"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()

	// Use tool_choice to force Claude to return structured JSON.
	// Defining a tool with a schema guarantees the response matches that shape.
	extractTool := anthropic.ToolParam{
		Name:        "extract_movie_review",
		Description: anthropic.String("Extract structured movie review data from the user's text."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "Movie title",
				},
				"year": map[string]any{
					"type":        "integer",
					"description": "Release year",
				},
				"rating": map[string]any{
					"type":        "number",
					"description": "Rating out of 10",
				},
				"summary": map[string]any{
					"type":        "string",
					"description": "One-sentence summary of the review",
				},
				"pros": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
					"description": "List of positives",
				},
				"cons": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
					"description": "List of negatives",
				},
			},
		},
	}

	reviewText := `I just watched Inception (2010) again. Still a masterpiece — 9/10.
The practical effects and Hans Zimmer's score are incredible, and the layered
dream logic keeps you thinking long after credits roll. Only gripe: some
exposition dumps slow the first act, and the ending is deliberately ambiguous
to the point of frustration.`

	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Tools:     []anthropic.ToolUnionParam{{OfTool: &extractTool}},
		// Force Claude to call our extraction tool
		ToolChoice: anthropic.ToolChoiceParamOfTool("extract_movie_review"),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(reviewText)),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Find the tool_use block and parse the JSON input
	for _, block := range resp.Content {
		if toolUse, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
			var review MovieReview
			if err := json.Unmarshal([]byte(toolUse.JSON.Input.Raw()), &review); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Title:   %s (%d)\n", review.Title, review.Year)
			fmt.Printf("Rating:  %.1f/10\n", review.Rating)
			fmt.Printf("Summary: %s\n", review.Summary)
			fmt.Println("\nPros:")
			for _, p := range review.Pros {
				fmt.Printf("  + %s\n", p)
			}
			fmt.Println("Cons:")
			for _, c := range review.Cons {
				fmt.Printf("  - %s\n", c)
			}
		}
	}
}
