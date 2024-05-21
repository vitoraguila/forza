package adaptors

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	utils "github.com/vitoraguila/forza/internal"
)

type OpenAIParams struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

type OpenAIService interface {
	WithModel(model string)
	AddFunction(name string, description string, params jsonschema.Definition, fn func(param string) string)
	Completion(prompt string, prompts *[]utils.AgentPrompts) string
}

type OpenAI struct {
	model        string
	Functions    []openai.FunctionDefinition
	FnExecutable *map[string]func(param string) string
}

func NewOpenAI() OpenAIService {
	var fnExecutable = make(map[string]func(param string) string)
	return &OpenAI{
		FnExecutable: &fnExecutable,
	}
}

func (oai *OpenAI) Completion(prompt string, prompts *[]utils.AgentPrompts) string {
	var fn []openai.Tool
	if len(oai.Functions) > 0 {
		for _, f := range oai.Functions {
			t := openai.Tool{
				Type:     openai.ToolTypeFunction,
				Function: &f,
			}
			fn = append(fn, t)
		}
	}

	var messages []openai.ChatCompletionMessage
	for _, p := range *prompts {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    p.Role,
			Content: p.Backstory,
		})
	}

	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	messages = append(messages, user)

	var client *openai.Client

	if oai.model == utils.ProviderAzure {
		azureApiKey := os.Getenv("AZURE_OPEN_AI_API_KEY")
		azureEndpoint := os.Getenv("AZURE_OPEN_AI_ENDPOINT")

		if azureApiKey == "" || azureEndpoint == "" {
			panic("AZURE_OPEN_AI_API_KEY or AZURE_OPEN_AI_ENDPOINT not provided")
		}

		config := openai.DefaultAzureConfig(azureApiKey, azureEndpoint)
		client = openai.NewClientWithConfig(config)
	} else {
		openAIApiKey := os.Getenv("OPENAI_API_KEY")

		if openAIApiKey == "" {
			panic("OPENAI_API_KEY Key not provided")
		}

		client = openai.NewClient(openAIApiKey)
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	}

	if len(fn) > 0 {
		openaiReq.Tools = fn
	}
	ctx := context.Background()
	resp, err := client.CreateChatCompletion(
		ctx,
		openaiReq,
	)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return ""
	}

	msg := resp.Choices[0].Message

	if len(msg.ToolCalls) > 0 {
		messages = append(messages, msg)
		fmt.Printf("OpenAI called us back wanting to invoke our function '%v' with params '%v'\n",
			msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)

		fn := (*oai.FnExecutable)[msg.ToolCalls[0].Function.Name]

		messages = append(messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    fn(msg.ToolCalls[0].Function.Arguments),
			Name:       msg.ToolCalls[0].Function.Name,
			ToolCallID: msg.ToolCalls[0].ID,
		})

		openaiReq.Messages = messages

		resp, err = client.CreateChatCompletion(ctx, openaiReq)
		if err != nil || len(resp.Choices) != 1 {
			fmt.Printf("2nd completion error: err:%v len(choices):%v\n", err,
				len(resp.Choices))
		}
	}

	return resp.Choices[0].Message.Content
}

func (oai *OpenAI) WithModel(model string) {
	oai.model = model
}

func (oai *OpenAI) AddFunction(name string, description string, params jsonschema.Definition, fn func(param string) string) {
	oai.Functions = append(oai.Functions, openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  params,
	})

	// store the function in a map
	(*oai.FnExecutable)[name] = fn
}
