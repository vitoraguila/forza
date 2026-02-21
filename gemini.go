package forza

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vitoraguila/forza/tools"
)

const geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

type geminiProvider struct {
	config        *LLMConfig
	functions     []geminiFunctionDecl
	fnExecutable  map[string]func(param string) (string, error)
	builtinTools  map[string]bool
	systemPrompts []agentPrompts
	userPrompt    *string
	httpClient    *http.Client
}

// --- Gemini API types ---

type geminiFunctionDecl struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type geminiRequest struct {
	Contents          []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	Tools             []geminiToolDef         `json:"tools,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiToolDef struct {
	FunctionDeclarations []geminiFunctionDecl `json:"functionDeclarations"`
}

type geminiGenerationConfig struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// --- Constructor ---

func newGemini(c *LLMConfig, a *Agent) LLMAgent {
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

	return &geminiProvider{
		config:        c,
		fnExecutable:  fnExecutable,
		systemPrompts: systemPrompts,
		builtinTools:  builtinTools,
		httpClient:    &http.Client{},
	}
}

func (g *geminiProvider) WithUserPrompt(prompt string) {
	g.userPrompt = &prompt
}

func (g *geminiProvider) WithTools(t ...tools.Tool) {
	for _, tool := range t {
		g.functions = append(g.functions, geminiFunctionDecl{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]string{"type": "string", "description": "input"},
				},
				"required": []string{"input"},
			},
		})
		g.fnExecutable[tool.Name()] = tool.Call
		g.builtinTools[tool.Name()] = true
	}
}

func (g *geminiProvider) AddCustomTools(name string, description string, params FunctionShape, fn func(param string) (string, error)) {
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

	g.functions = append(g.functions, geminiFunctionDecl{
		Name:        name,
		Description: description,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		},
	})
	g.fnExecutable[name] = fn
}

func (g *geminiProvider) Completion(params ...string) (string, error) {
	if g.userPrompt == nil {
		return "", ErrMissingPrompt
	}

	userPrompt := *g.userPrompt
	if len(params) > 1 {
		return "", ErrTooManyArgs
	}
	if len(params) == 1 {
		userPrompt = userPrompt + "\n\nTake in consideration the following context: " + params[0]
	}

	apiKey := g.config.credentials.apiKey
	if apiKey == "" {
		return "", fmt.Errorf("%w: Gemini API key", ErrMissingAPIKey)
	}

	// Build system instruction
	var systemText string
	for _, p := range g.systemPrompts {
		if systemText != "" {
			systemText += "\n"
		}
		systemText += p.Context
	}

	// Build request
	req := geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userPrompt}},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: g.config.temperature,
			MaxTokens:   g.config.maxTokens,
		},
	}

	if systemText != "" {
		req.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: systemText}},
		}
	}

	if len(g.functions) > 0 {
		req.Tools = []geminiToolDef{
			{FunctionDeclarations: g.functions},
		}
	}

	resp, err := g.doRequest(apiKey, req)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("%w: no candidates returned", ErrCompletionFailed)
	}

	// Check for function calls
	var functionCalls []geminiFunctionCall
	var textContent string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.FunctionCall != nil {
			functionCalls = append(functionCalls, *part.FunctionCall)
		}
		if part.Text != "" {
			textContent += part.Text
		}
	}

	if len(functionCalls) > 0 {
		// Add model response to contents
		req.Contents = append(req.Contents, resp.Candidates[0].Content)

		// Execute functions and build response parts
		var responseParts []geminiPart
		for _, fc := range functionCalls {
			fn, exists := g.fnExecutable[fc.Name]
			if !exists {
				return "", fmt.Errorf("%w: unknown tool %q", ErrToolCallFailed, fc.Name)
			}

			// Extract input
			toolInput := ""
			if g.builtinTools[fc.Name] {
				if v, ok := fc.Args["input"]; ok {
					if s, ok := v.(string); ok {
						toolInput = s
					}
				}
			} else {
				// For custom tools, serialize args as JSON
				argsJSON, _ := json.Marshal(fc.Args)
				toolInput = string(argsJSON)
			}

			content, err := fn(toolInput)
			if err != nil {
				return "", fmt.Errorf("%w: tool %q: %v", ErrToolCallFailed, fc.Name, err)
			}

			responseParts = append(responseParts, geminiPart{
				FunctionResponse: &geminiFunctionResponse{
					Name: fc.Name,
					Response: map[string]interface{}{
						"result": content,
					},
				},
			})
		}

		req.Contents = append(req.Contents, geminiContent{
			Role:  "user",
			Parts: responseParts,
		})

		resp, err = g.doRequest(apiKey, req)
		if err != nil {
			return "", fmt.Errorf("%w: follow-up after tool call: %v", ErrCompletionFailed, err)
		}

		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("%w: no candidates in follow-up response", ErrCompletionFailed)
		}

		textContent = ""
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				textContent += part.Text
			}
		}
	}

	return textContent, nil
}

func (g *geminiProvider) doRequest(apiKey string, reqBody geminiRequest) (*geminiResponse, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIBaseURL, g.config.model, apiKey)

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal request: %v", ErrCompletionFailed, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrCompletionFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrCompletionFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrCompletionFailed, err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrCompletionFailed, err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("%w: API error [%s]: %s", ErrCompletionFailed, geminiResp.Error.Status, geminiResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d: %s", ErrCompletionFailed, resp.StatusCode, string(respBody))
	}

	return &geminiResp, nil
}
