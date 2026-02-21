package forza

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vitoraguila/forza/tools"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const anthropicAPIVersion = "2023-06-01"

type anthropicProvider struct {
	config        *LLMConfig
	functions     []anthropicToolDef
	fnExecutable  map[string]func(param string) (string, error)
	builtinTools  map[string]bool
	systemPrompts []agentPrompts
	userPrompt    *string
	httpClient    *http.Client
}

// --- Anthropic API types ---

type anthropicToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicRequest struct {
	Model       string                 `json:"model"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature,omitempty"`
	System      string                 `json:"system,omitempty"`
	Messages    []anthropicMessage     `json:"messages"`
	Tools       []anthropicToolDef     `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type anthropicContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type anthropicToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type anthropicResponse struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
	Model   string                  `json:"model"`
	Error   *anthropicError         `json:"error,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// --- Constructor ---

func newAnthropic(c *LLMConfig, a *Agent) LLMAgent {
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

	return &anthropicProvider{
		config:        c,
		fnExecutable:  fnExecutable,
		systemPrompts: systemPrompts,
		builtinTools:  builtinTools,
		httpClient:    &http.Client{},
	}
}

func (a *anthropicProvider) WithUserPrompt(prompt string) {
	a.userPrompt = &prompt
}

func (a *anthropicProvider) WithTools(t ...tools.Tool) {
	for _, tool := range t {
		a.functions = append(a.functions, anthropicToolDef{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]string{"type": "string", "description": "input"},
				},
				"required": []string{"input"},
			},
		})
		a.fnExecutable[tool.Name()] = tool.Call
		a.builtinTools[tool.Name()] = true
	}
}

func (a *anthropicProvider) AddCustomTools(name string, description string, params FunctionShape, fn func(param string) (string, error)) {
	properties := make(map[string]interface{})
	var required []string

	for fieldName, props := range params {
		properties[fieldName] = map[string]string{
			"type":        "string",
			"description": props.Description,
		}
		if props.Required {
			required = append(required, fieldName)
		}
	}

	a.functions = append(a.functions, anthropicToolDef{
		Name:        name,
		Description: description,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		},
	})
	a.fnExecutable[name] = fn
}

func (a *anthropicProvider) Completion(params ...string) (string, error) {
	if a.userPrompt == nil {
		return "", ErrMissingPrompt
	}

	userPrompt := *a.userPrompt
	if len(params) > 1 {
		return "", ErrTooManyArgs
	}
	if len(params) == 1 {
		userPrompt = userPrompt + "\n\nTake in consideration the following context: " + params[0]
	}

	apiKey := a.config.credentials.apiKey
	if apiKey == "" {
		return "", fmt.Errorf("%w: Anthropic API key", ErrMissingAPIKey)
	}

	// Build system prompt
	var systemPrompt string
	for _, p := range a.systemPrompts {
		if systemPrompt != "" {
			systemPrompt += "\n"
		}
		systemPrompt += p.Context
	}

	// Build messages
	messages := []anthropicMessage{
		{Role: "user", Content: userPrompt},
	}

	// Build request
	req := anthropicRequest{
		Model:       a.config.model,
		MaxTokens:   a.config.maxTokens,
		Temperature: a.config.temperature,
		System:      systemPrompt,
		Messages:    messages,
	}
	if len(a.functions) > 0 {
		req.Tools = a.functions
	}

	resp, err := a.doRequest(apiKey, req)
	if err != nil {
		return "", err
	}

	// Check for tool use
	var toolUseBlocks []anthropicContentBlock
	var textContent string
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			textContent += block.Text
		case "tool_use":
			toolUseBlocks = append(toolUseBlocks, block)
		}
	}

	if len(toolUseBlocks) > 0 {
		// Add assistant response to messages
		req.Messages = append(req.Messages, anthropicMessage{
			Role:    "assistant",
			Content: resp.Content,
		})

		// Execute tools and build result blocks
		var toolResults []anthropicToolResult
		for _, toolBlock := range toolUseBlocks {
			fn, exists := a.fnExecutable[toolBlock.Name]
			if !exists {
				return "", fmt.Errorf("%w: unknown tool %q", ErrToolCallFailed, toolBlock.Name)
			}

			toolInput := string(toolBlock.Input)

			// For builtin tools, extract "input" field
			if a.builtinTools[toolBlock.Name] {
				inputMap := make(map[string]interface{})
				if err := json.Unmarshal(toolBlock.Input, &inputMap); err == nil {
					if v, ok := inputMap["input"]; ok {
						if s, ok := v.(string); ok {
							toolInput = s
						}
					}
				}
			}

			content, err := fn(toolInput)
			if err != nil {
				return "", fmt.Errorf("%w: tool %q: %v", ErrToolCallFailed, toolBlock.Name, err)
			}

			toolResults = append(toolResults, anthropicToolResult{
				Type:      "tool_result",
				ToolUseID: toolBlock.ID,
				Content:   content,
			})
		}

		// Send tool results back
		req.Messages = append(req.Messages, anthropicMessage{
			Role:    "user",
			Content: toolResults,
		})

		resp, err = a.doRequest(apiKey, req)
		if err != nil {
			return "", fmt.Errorf("%w: follow-up after tool call: %v", ErrCompletionFailed, err)
		}

		textContent = ""
		for _, block := range resp.Content {
			if block.Type == "text" {
				textContent += block.Text
			}
		}
	}

	return textContent, nil
}

func (a *anthropicProvider) doRequest(apiKey string, reqBody anthropicRequest) (*anthropicResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal request: %v", ErrCompletionFailed, err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrCompletionFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrCompletionFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrCompletionFailed, err)
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrCompletionFailed, err)
	}

	if anthropicResp.Error != nil {
		return nil, fmt.Errorf("%w: API error [%s]: %s", ErrCompletionFailed, anthropicResp.Error.Type, anthropicResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrCompletionFailed, resp.StatusCode)
	}

	return &anthropicResp, nil
}
