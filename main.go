package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

// Fill in your Azure credentials.
const baseUrl = "https://EXAMPLE.openai.azure.com"
const model = "gpt-35-turbo"
const apiKey = "AZURE_API_KEY"

func main() {
	config := openai.DefaultAzureConfig(apiKey, baseUrl, model)
	client := openai.NewClientWithConfig(config)
	ctx := context.Background()

	app := &cli.App{
		Name:        "how",
		Description: "Copilot for your terminal",
		Usage:       "how <question>",
		Version:     "1.0.1",
		Action: func(c *cli.Context) error {
			q := strings.Join(c.Args().Slice(), " ")
			return getAnswer(client, &ctx, q)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func getAnswer(client *openai.Client, ctx *context.Context, query string) error {
	if query == "" {
		return fmt.Errorf("no question provided")
	}

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 400,
		Messages:  startConversation(query),
		Stream:    true,
	}

	stream, err := client.CreateChatCompletionStream(*ctx, req)
	if err != nil {
		return err
	}

	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		ans := resp.Choices[0].Delta.Content
		fmt.Print(ans)
	}
}

func startConversation(query string) []openai.ChatCompletionMessage {
	var systemPrompt string
	if runtime.GOOS == "windows" {
		systemPrompt = "You are a proficient PowerShell user. Answer my question with PowerShell commands."
	} else {
		systemPrompt = "You are a proficient terminal user. Answer my question with terminal commands."
	}

	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: query,
		},
	}
}
