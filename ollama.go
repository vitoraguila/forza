package forza

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/vitoraguila/forza/tools"
)

const defaultOllamaEndpoint = "http://localhost:11434/v1"

type ollamaProvider struct {
	config        *LLMConfig
	functions     []openai.FunctionDefinition
	fnExecutable  map[string]func(param string) (string, error)
	builtinTools  map[string]bool
	systemPrompts []agentPrompts
	userPrompt    *string
}

func newOllama(c *LLMConfig, a *Agent) LLMAgent {
	fnExecutable := make(map[string]func(param string) (string, error))
	builtinTools := make(map[string]bool)

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

	return &ollamaProvider{
		config:        c,
		fnExecutable:  fnExecutable,
		systemPrompts: systemPrompts,
		builtinTools:  builtinTools,
	}
}

func (o *ollamaProvider) WithUserPrompt(prompt string) {
	o.userPrompt = &prompt
}

func (o *ollamaProvider) WithTools(t ...tools.Tool) {
	for _, tool := range t {
		o.functions = append(o.functions, openai.FunctionDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters: map[string]any{
				"properties": map[string]any{
					"input": map[string]string{"title": "input", "type": "string"},
				},
				"required": []string{"input"},
				"type":     "object",
			},
		})
		o.fnExecutable[tool.Name()] = tool.Call
		o.builtinTools[tool.Name()] = true
	}
}

func (o *ollamaProvider) AddCustomTools(name string, description string, params FunctionShape, fn func(param string) (string, error)) {
	properties := make(map[string]jsonschema.Definition)
	var required []string

	for fieldName, props := range params {
		properties[fieldName] = jsonschema.Definition{
			Type:        jsonschema.String,
			Description: props.Description,
		}
		if props.Required {
			required = append(required, fieldName)
		}
	}

	o.functions = append(o.functions, openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters: jsonschema.Definition{
			Type:       jsonschema.Object,
			Properties: properties,
			Required:   required,
		},
	})
	o.fnExecutable[name] = fn
}

func (o *ollamaProvider) Completion(params ...string) (string, error) {
	if o.userPrompt == nil {
		return "", ErrMissingPrompt
	}

	userPrompt := *o.userPrompt
	if len(params) > 1 {
		return "", ErrTooManyArgs
	}
	if len(params) == 1 {
		userPrompt = userPrompt + "\n\nTake in consideration the following context: " + params[0]
	}

	// Build tools
	var fn []openai.Tool
	for _, f := range o.functions {
		fn = append(fn, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &f,
		})
	}

	// Build messages
	var messages []openai.ChatCompletionMessage
	for _, p := range o.systemPrompts {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    p.Role,
			Content: p.Context,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userPrompt,
	})

	// Create client pointing to Ollama's OpenAI-compatible endpoint
	client, err := o.createClient()
	if err != nil {
		return "", err
	}

	req := openai.ChatCompletionRequest{
		Model:       o.config.model,
		Messages:    messages,
		Temperature: float32(o.config.temperature),
	}
	if len(fn) > 0 {
		req.Tools = fn
	}

	ctx := context.Background()
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCompletionFailed, err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrCompletionFailed)
	}

	msg := resp.Choices[0].Message

	// Handle tool calls
	if len(msg.ToolCalls) > 0 {
		messages = append(messages, msg)

		for _, toolCall := range msg.ToolCalls {
			fn, exists := o.fnExecutable[toolCall.Function.Name]
			if !exists {
				return "", fmt.Errorf("%w: unknown tool %q", ErrToolCallFailed, toolCall.Function.Name)
			}

			toolInput := toolCall.Function.Arguments
			if o.builtinTools[toolCall.Function.Name] {
				toolInputMap := make(map[string]any)
				if err := json.Unmarshal([]byte(toolInput), &toolInputMap); err == nil {
					if arg, ok := toolInputMap["input"]; ok {
						if s, ok := arg.(string); ok {
							toolInput = s
						}
					}
				}
			}

			content, err := fn(toolInput)
			if err != nil {
				return "", fmt.Errorf("%w: tool %q: %v", ErrToolCallFailed, toolCall.Function.Name, err)
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    content,
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			})
		}

		req.Messages = messages
		resp, err = client.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", fmt.Errorf("%w: follow-up after tool call: %v", ErrCompletionFailed, err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("%w: no choices in follow-up response", ErrCompletionFailed)
		}
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *ollamaProvider) createClient() (*openai.Client, error) {
	endpoint := o.config.credentials.endpoint
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}

	config := openai.DefaultConfig("ollama")
	config.BaseURL = endpoint
	return openai.NewClientWithConfig(config), nil
}
