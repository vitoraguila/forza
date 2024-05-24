package forza

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type openAI struct {
	Config        *llmCOnfig
	Functions     []openai.FunctionDefinition
	FnExecutable  *map[string]func(param string) string
	systemPrompts *[]agentPrompts
	userPrompt    *string
}

func NewOpenAI(c *llmCOnfig, t *task) llmService {
	var fnExecutable = make(map[string]func(param string) string)

	if t.agent.Role == "" || t.agent.Backstory == "" || t.agent.Goal == "" {
		panic("Agent Role(WithRole()), Backstory(WithBackstory()) and Goal(WithGoal) are required")
	}

	systemPrompts := []agentPrompts{
		{
			Role:    agentRoleSystem,
			Context: fmt.Sprintf("As a %s, %s", t.agent.Role, t.agent.Backstory),
		},
		{
			Role:    agentRoleSystem,
			Context: fmt.Sprintf("Your goal is %s", t.agent.Goal),
		},
	}

	return &openAI{
		Config:        c,
		FnExecutable:  &fnExecutable,
		systemPrompts: &systemPrompts,
		userPrompt:    &t.prompt,
	}
}

func (o *openAI) WithUserPrompt(prompt string) {
	o.userPrompt = &prompt
}

func (o *openAI) AddFunctions(name string, description string, params functionShape, fn func(param string) string) {
	jsonschema := generateSchema(params)

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

		fn := (*o.FnExecutable)[msg.ToolCalls[0].Function.Name]

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
