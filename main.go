package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/mattn/go-shellwords"
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
		Version:     "1.0.2",
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
		Model:       openai.GPT3Dot5Turbo,
		MaxTokens:   400,
		Messages:    startConversation(query),
		Stream:      false,
		Temperature: 0,
	}

	resp, err := client.CreateChatCompletion(*ctx, req)
	if err != nil {
		return err
	}

	ans := resp.Choices[0].Message.Content
	var answer Answer
	if err := json.Unmarshal([]byte(ans), &answer); err != nil {
		return err
	}

	outputAnswer(answer)
	return nil
}

func startConversation(query string) []openai.ChatCompletionMessage {
	var systemPrompt string
	if runtime.GOOS == "windows" {
		systemPrompt = "You are a proficient PowerShell user."
	} else {
		systemPrompt = "You are a proficient terminal user."
	}

	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: makePrompt(query),
		},
	}
}

func makePrompt(query string) string {
	prompt := `
	Use terminal command to complete the task delimited by triple quotes.
	Provide answer in JSON format. "command" key contains the command to run.
	"explanation" key contains the explanation of the command.
	If no such command exists, provide an empty string for "command" key.
	"""
	%s
	"""
	`

	return fmt.Sprintf(prompt, query)
}

func outputAnswer(answer Answer) {
	if answer.Command == "" {
		fmt.Println("No such command exists.")
		return
	}

	output := `
Command:
    
    %s

Explanation:

    %s

`
	fmt.Printf(output, answer.Command, answer.Explanation)

	sp := confirmation.New("Execute the command?", confirmation.Undecided)
	ready, err := sp.RunPrompt()
	if err != nil {
		fmt.Println(err)
		return
	}

	if ready {
		fmt.Println("")
		if err := executeCommand(answer.Command); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("OK!")
	}
}

func executeCommand(command string) error {
	args, err := shellwords.Parse(command)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

type Answer struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
}
