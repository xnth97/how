package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
)

// Fill in your Azure credentials.
const baseUrl = "https://EXAMPLE.openai.azure.com"
const model = "gpt-35-turbo"
const apiVersion = "2023-03-15-preview"
const apiKey = "AZURE_API_KEY"

func main() {
	app := &cli.App{
		Name:        "how",
		Description: "Copilot for your terminal",
		Usage:       "how <question>",
		Version:     "1.0.0",
		Action: func(c *cli.Context) error {
			q := strings.Join(c.Args().Slice(), " ")
			answer, err := getAnswer(q)
			if err != nil {
				return err
			}
			fmt.Println(answer)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func getAnswer(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("no question provided")
	}

	client := NewClient(baseUrl, model, apiVersion, apiKey)
	conv := startConversation()
	return client.Chat(conv, query)
}

func startConversation() *Conversation {
	var systemPrompt string
	if runtime.GOOS == "windows" {
		systemPrompt = "You are a proficient PowerShell user. Answer my question with PowerShell commands."
	} else {
		systemPrompt = "You are a proficient terminal user. Answer my question with terminal commands."
	}

	return &Conversation{
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
		},
	}
}
