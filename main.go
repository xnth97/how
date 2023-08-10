package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/mattn/go-shellwords"
	"github.com/urfave/cli/v2"
)

// Fill in your Azure credentials.
const baseUrl = "https://EXAMPLE.openai.azure.com"
const model = "gpt-35-turbo"
const apiKey = "AZURE_API_KEY"

func main() {
	keyCredential, err := azopenai.NewKeyCredential(apiKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := azopenai.NewClientWithKeyCredential(baseUrl, keyCredential, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()

	app := &cli.App{
		Name:        "how",
		Description: "Copilot for your terminal",
		Usage:       "how <question>",
		Version:     "1.0.3",
		Action: func(c *cli.Context) error {
			q := strings.Join(c.Args().Slice(), " ")
			return getAnswer(client, &ctx, q)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func getAnswer(client *azopenai.Client, ctx *context.Context, query string) error {
	if query == "" {
		return fmt.Errorf("no question provided")
	}

	req := azopenai.ChatCompletionsOptions{
		DeploymentID: model,
		MaxTokens:    to.Ptr(int32(400)),
		Messages:     startConversation(query),
		Temperature:  to.Ptr(float32(0)),
	}

	resp, err := client.GetChatCompletions(*ctx, req, nil)
	if err != nil {
		return err
	}

	ans := *resp.Choices[0].Message.Content
	var answer Answer
	if err := json.Unmarshal([]byte(ans), &answer); err != nil {
		return err
	}

	outputAnswer(answer)
	return nil
}

func startConversation(query string) []azopenai.ChatMessage {
	var systemPrompt string
	if runtime.GOOS == "windows" {
		systemPrompt = "You are a proficient PowerShell user."
	} else {
		systemPrompt = "You are a proficient terminal user."
	}

	return []azopenai.ChatMessage{
		{
			Role:    to.Ptr(azopenai.ChatRoleSystem),
			Content: &systemPrompt,
		},
		{
			Role:    to.Ptr(azopenai.ChatRoleUser),
			Content: makePrompt(query),
		},
	}
}

func makePrompt(query string) *string {
	prompt := `
	Use terminal command to complete the task delimited by triple quotes.
	Provide answer in JSON format. "command" key contains the command to run.
	"explanation" key contains the explanation of the command.
	If no such command exists, provide an empty string for "command" key.
	"""
	%s
	"""
	`

	p := fmt.Sprintf(prompt, query)
	return &p
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
