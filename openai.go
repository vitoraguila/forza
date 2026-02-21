package forza

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/vitoraguila/forza/tools"
)

type openaiProvider struct {
	config        *LLMConfig
	functions     []openai.FunctionDefinition
	fnExecutable  map[string]func(param string) (string, error)
	builtinTools  map[string]bool
	systemPrompts []agentPrompts
	userPrompt    *string
	client        *openai.Client // cached client, also used for testing
}

func newOpenAI(c *LLMConfig, a *Agent) LLMAgent {
	fnExecutable := make(map[string]func(param string) (string, error))
	builtinTools := make(map[string]bool)

	return &openaiProvider{
		config:        c,
		fnExecutable:  fnExecutable,
		systemPrompts: buildSystemPrompts(a),
		builtinTools:  builtinTools,
	}
}

func (o *openaiProvider) WithUserPrompt(prompt string) {
	o.userPrompt = &prompt
}

func (o *openaiProvider) WithTools(t ...tools.Tool) {
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
		o.fnExecutable[tool.Name()] = func(t tools.Tool) func(string) (string, error) {
			return func(input string) (string, error) {
				return t.Call(context.Background(), input)
			}
		}(tool)
		o.builtinTools[tool.Name()] = true
	}
}

func (o *openaiProvider) AddCustomTools(name string, description string, params FunctionShape, fn func(param string) (string, error)) {
	schema := generateOpenAISchema(params)

	o.functions = append(o.functions, openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  schema,
	})
	o.fnExecutable[name] = fn
}

func (o *openaiProvider) Completion(ctx context.Context, params ...string) (string, error) {
	userPrompt, err := resolveUserPrompt(o.userPrompt, params)
	if err != nil {
		return "", err
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

	// Get or create client
	client, err := o.getClient()
	if err != nil {
		return "", err
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model:       o.config.model,
		Messages:    messages,
		Temperature: float32(o.config.temperature),
		MaxTokens:   o.config.maxTokens,
	}
	if len(fn) > 0 {
		req.Tools = fn
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCompletionFailed, err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrCompletionFailed)
	}

	msg := resp.Choices[0].Message

	// Handle tool calls with depth limit
	for round := 0; len(msg.ToolCalls) > 0 && round < defaultMaxToolRounds; round++ {
		messages = append(messages, msg)

		for _, toolCall := range msg.ToolCalls {
			toolFn, exists := o.fnExecutable[toolCall.Function.Name]
			if !exists {
				return "", fmt.Errorf("%w: unknown tool %q", ErrToolCallFailed, toolCall.Function.Name)
			}

			toolInput := toolCall.Function.Arguments
			if o.builtinTools[toolCall.Function.Name] {
				toolInput = extractBuiltinToolInput(toolInput)
			}

			content, err := toolFn(toolInput)
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
		msg = resp.Choices[0].Message
	}

	if len(msg.ToolCalls) > 0 {
		return "", fmt.Errorf("%w: exceeded %d rounds", ErrMaxToolRoundsExceeded, defaultMaxToolRounds)
	}

	return msg.Content, nil
}

func (o *openaiProvider) getClient() (*openai.Client, error) {
	if o.client != nil {
		return o.client, nil
	}
	client, err := o.createClient()
	if err != nil {
		return nil, err
	}
	o.client = client
	return client, nil
}

func (o *openaiProvider) createClient() (*openai.Client, error) {
	if o.config.provider == ProviderAzure {
		apiKey := o.config.credentials.apiKey
		endpoint := o.config.credentials.endpoint
		if apiKey == "" {
			return nil, fmt.Errorf("%w: Azure OpenAI API key", ErrMissingAPIKey)
		}
		if endpoint == "" {
			return nil, fmt.Errorf("%w: Azure OpenAI endpoint", ErrMissingEndpoint)
		}
		config := openai.DefaultAzureConfig(apiKey, endpoint)
		return openai.NewClientWithConfig(config), nil
	}

	apiKey := o.config.credentials.apiKey
	if apiKey == "" {
		return nil, fmt.Errorf("%w: OpenAI API key", ErrMissingAPIKey)
	}
	return openai.NewClient(apiKey), nil
}

// generateOpenAISchema converts a FunctionShape to an OpenAI-compatible JSON Schema.
func generateOpenAISchema(shape FunctionShape) jsonschema.Definition {
	properties := make(map[string]jsonschema.Definition)
	var required []string

	for fieldName, props := range shape {
		properties[fieldName] = jsonschema.Definition{
			Type:        jsonschema.String,
			Description: props.Description,
		}
		if props.Required {
			required = append(required, fieldName)
		}
	}

	return jsonschema.Definition{
		Type:       jsonschema.Object,
		Properties: properties,
		Required:   required,
	}
}
