package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

// generate asks Claude for a one-sentence short-film idea at the given
// temperature. Temperature controls randomness: values near 0.0 are
// focused and (nearly) repeatable, while values near 1.0 are more varied
// and creative. Defaults to 1.0 when unset.
func generate(ctx context.Context, client anthropic.Client, temperature float64) (string, error) {
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:       anthropic.ModelClaudeSonnet4_6,
		MaxTokens:   1024,
		Temperature: anthropic.Float(temperature),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Generate a 1 sentence short film idea.")),
		},
	})
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

	// Run the same prompt twice at each temperature so the difference in
	// variability is visible: low temperature stays close to itself, high
	// temperature wanders.
	for _, temp := range []float64{0.0, 1.0} {
		fmt.Printf("=== temperature %.1f ===\n", temp)
		for i := 0; i < 2; i++ {
			idea, err := generate(ctx, client, temp)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%d. %s\n", i+1, idea)
		}
		fmt.Println()
	}
}
