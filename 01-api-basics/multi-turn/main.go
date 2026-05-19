package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	client := anthropic.NewClient()
	history := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (type 'quit' to exit)")
	fmt.Println("--------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" {
			break
		}

		history = append(history, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

		resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_6,
			MaxTokens: 1024,
			Messages:  history,
		})
		if err != nil {
			log.Fatal(err)
		}

		var reply string
		for _, block := range resp.Content {
			if text, ok := block.AsAny().(anthropic.TextBlock); ok {
				reply += text.Text
			}
		}

		fmt.Printf("Claude: %s\n\n", reply)

		// append assistant response to keep conversation context
		history = append(history, resp.ToParam())
	}
}
