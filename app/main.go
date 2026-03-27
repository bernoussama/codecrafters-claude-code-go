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
	handler func(map[string]any) string
}

func handleTools(msg openai.ChatCompletionMessage, tools map[string]toolDef) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion
	messages = append(messages, msg.ToParam())

	for _, toolCall := range msg.ToolCalls {
		name := toolCall.Function.Name
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse args: %v\n", err)
			messages = append(messages, openai.ToolMessage(fmt.Sprintf("error: %v", err), toolCall.ID))
			continue
		}

		t, ok := tools[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown tool: %s\n", name)
			messages = append(messages, openai.ToolMessage(fmt.Sprintf("unknown tool: %s", name), toolCall.ID))
			continue
		}

		result := t.handler(args)
		fmt.Print(result)
		messages = append(messages, openai.ToolMessage(result, toolCall.ID))
	}

	return messages
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
			handler: func(args map[string]any) string {
				path, _ := args["file_path"].(string)
				data, err := os.ReadFile(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
					return fmt.Sprintf("Error reading file: %v", err)
				}
				return string(data)
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

	var conversation []openai.ChatCompletionMessageParamUnion
	conversation = append(conversation, openai.UserMessage(prompt))

	for {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    "anthropic/claude-haiku-4.5",
				Tools:    toolList,
				Messages: conversation,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

		msg := resp.Choices[0].Message

		// handle tool calls
		var toolMessages []openai.ChatCompletionMessageParamUnion
		if len(msg.ToolCalls) != 0 {
			toolMessages = handleTools(msg, tools)
		} else {
			fmt.Print(msg.Content)
			break
		}

		conversation = append(conversation, toolMessages...)
	}
}
