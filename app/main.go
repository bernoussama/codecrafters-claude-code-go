package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type toolDef struct {
	def     openai.ChatCompletionToolUnionParam
	handler func(map[string]any)
}

func main() {
	tools := map[string]toolDef{
		"Read": {
			def: openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
				Name:        "Read",
				Description: openai.String("Read and return the contents of a file"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"file_path": map[string]any{
							"type":        "string",
							"description": "The path to the file to read",
						},
					},
					"required": []string{"file_path"},
				},
			}),
			handler: func(args map[string]any) {
				path, _ := args["file_path"].(string)
				data, err := os.ReadFile(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
					return
				}
				fmt.Printf("%s", string(data))
			},
		},
	}

	toolList := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
	for _, t := range tools {
		toolList = append(toolList, t.def)
	}
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: "anthropic/claude-haiku-4.5",
			Tools: toolList,
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	// TODO: Uncomment the line below to pass the first stage
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) == 0 {
		fmt.Print(msg.Content)
	} else {
		for _, toolCall := range msg.ToolCalls {
			name := toolCall.Function.Name
			var args map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				fmt.Fprintf(os.Stderr, "failed to parse args: %v\n", err)
				continue
			}

			t, ok := tools[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown tool: %s\n", name)
				continue
			}
			t.handler(args)
		}
	}
}
