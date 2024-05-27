package forza

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/vitoraguila/forza/tools"
)

type openAiCredentials struct {
	openAiApiKey        string
	azureOpenAiApiKey   string
	azureOpenAiEndpoint string
}

type openAI struct {
	Config        *llmConfig
	Functions     []openai.FunctionDefinition
	FnExecutable  *map[string]func(param string) (string, error)
	tools         *map[string]bool
	systemPrompts *[]agentPrompts
	userPrompt    *string
}

func NewOpenAI(c *llmConfig, a *agent) llmAgent {
	var fnExecutable = make(map[string]func(param string) (string, error))
	var tools = make(map[string]bool)

	systemPrompts := []agentPrompts{
		{
			Role:    agentRoleSystem,
			Context: fmt.Sprintf("As a %s, %s", a.Role, a.Backstory),
		},
		{
			Role:    agentRoleSystem,
			Context: fmt.Sprintf("Your goal is %s", a.Goal),
		},
	}

	return &openAI{
		Config:        c,
		FnExecutable:  &fnExecutable,
		systemPrompts: &systemPrompts,
		tools:         &tools,
	}
}

func (o *openAI) WithUserPrompt(prompt string) {
	o.userPrompt = &prompt
}

func (o *openAI) WithTools(tools ...tools.Tool) {
	for _, t := range tools {
		fmt.Println("Adding tool: ", t.Name())
		o.Functions = append(o.Functions, openai.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"properties": map[string]any{
					"input": map[string]string{"title": "input", "type": "string"},
				},
				"required": []string{"input"},
				"type":     "object",
			},
		})

		(*o.FnExecutable)[t.Name()] = t.Call
		(*o.tools)[t.Name()] = true
	}
}

func (o *openAI) AddCustomTools(name string, description string, params functionShape, fn func(param string) (string, error)) {
	jsonschema := generateSchema(params)

	fmt.Println("Adding tool: ", name)
	o.Functions = append(o.Functions, openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  jsonschema,
	})

	// store the function in a map
	(*o.FnExecutable)[name] = fn
}

func (o openAI) Completion(params ...string) string {
	if o.userPrompt == nil {
		panic("User prompt is required")
	}

	var userPrompt string = *o.userPrompt

	if len(params) > 1 {
		panic("Error: too many arguments. Only one optional argument(context) is allowed.")
	}

	if len(params) == 1 {
		var promptContext = params[0]
		userPrompt = userPrompt + "\n\n take in consideration the following context: " + promptContext
	}

	var fn []openai.Tool
	if len(o.Functions) > 0 {
		for _, f := range o.Functions {
			t := openai.Tool{
				Type:     openai.ToolTypeFunction,
				Function: &f,
			}
			fn = append(fn, t)
		}
	}

	var messages []openai.ChatCompletionMessage
	for _, p := range *o.systemPrompts {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    p.Role,
			Content: p.Context,
		})
	}

	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userPrompt,
	}
	messages = append(messages, user)

	var client *openai.Client

	if o.Config.provider == ProviderAzure {
		azureApiKey := o.Config.credentials.openAi.azureOpenAiApiKey
		azureEndpoint := o.Config.credentials.openAi.azureOpenAiEndpoint

		if azureApiKey == "" || azureEndpoint == "" {
			panic("AZURE_OPEN_AI_API_KEY or AZURE_OPEN_AI_ENDPOINT not provided")
		}

		config := openai.DefaultAzureConfig(azureApiKey, azureEndpoint)
		client = openai.NewClientWithConfig(config)
	} else {
		openAIApiKey := o.Config.credentials.openAi.openAiApiKey

		if openAIApiKey == "" {
			panic("OPENAI_API_KEY Key not provided")
		}

		client = openai.NewClient(openAIApiKey)
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:    o.Config.model,
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
		// fmt.Printf("OpenAI called us back wanting to invoke our function '%v' with params '%v'\n",
		// 	msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)

		for _, toolCall := range msg.ToolCalls {
			fn := (*o.FnExecutable)[toolCall.Function.Name]

			toolInputStr := toolCall.Function.Arguments

			toolInput := toolInputStr

			if (*o.tools)[toolCall.Function.Name] {
				toolInputMap := make(map[string]any, 0)
				err := json.Unmarshal([]byte(toolInputStr), &toolInputMap)
				if err != nil {
					// return nil, nil, err
				}

				if arg1, ok := toolInputMap["input"]; ok {
					toolInputCheck, ok := arg1.(string)
					if ok {
						toolInput = toolInputCheck
					}
				}
			}

			content, err := fn(toolInput)
			if err != nil {
				fmt.Printf("Error calling function: %v\n", err)
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    content,
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			})
		}

		openaiReq.Messages = messages

		resp, err = client.CreateChatCompletion(ctx, openaiReq)
		if err != nil || len(resp.Choices) != 1 {
			fmt.Printf("2nd completion error: err:%v len(choices):%v\n", err,
				len(resp.Choices))
		}
	}

	return resp.Choices[0].Message.Content
}
