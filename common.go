package forza

import (
	"encoding/json"
	"fmt"
)

// agentPrompts holds a role and context for system prompt construction.
type agentPrompts struct {
	Role    string
	Context string
}

// Provider constants.
const (
	ProviderOpenAi    = "openai"
	ProviderAzure     = "openai-azure"
	ProviderAnthropic = "anthropic"
	ProviderGemini    = "gemini"
	ProviderOllama    = "ollama"
)

// Role constants for prompt messages.
const (
	agentRoleSystem = "system"
	agentRoleUser   = "user"
)

// Shared limits.
const (
	maxResponseSize      = 10 << 20 // 10 MB
	defaultMaxToolRounds = 10
)

// Models interface allows listing available models for a provider.
type Models interface {
	ListModels() []string
}

// --- OpenAI Models ---

// OpenAIModelList holds available OpenAI model identifiers.
type OpenAIModelList struct {
	GPT35Turbo string // Deprecated: use GPT4oMini instead
	GPT4       string
	GPT4o      string
	GPT4Turbo  string
	GPT4oMini  string
	O1Mini     string
	O1         string
	GPT5       string
	Codex52    string
}

func (m OpenAIModelList) ListModels() []string {
	return []string{m.GPT35Turbo, m.GPT4, m.GPT4o, m.GPT4Turbo, m.GPT4oMini, m.O1Mini, m.O1, m.GPT5, m.Codex52}
}

// OpenAIModels contains the predefined OpenAI model strings.
var OpenAIModels = OpenAIModelList{
	GPT35Turbo: "gpt-3.5-turbo",
	GPT4:       "gpt-4",
	GPT4o:      "gpt-4o",
	GPT4Turbo:  "gpt-4-turbo",
	GPT4oMini:  "gpt-4o-mini",
	O1Mini:     "o1-mini",
	O1:         "o1",
	GPT5:       "gpt-5",
	Codex52:    "codex-5.2",
}

// --- Anthropic Models ---

// AnthropicModelList holds available Anthropic model identifiers.
type AnthropicModelList struct {
	Claude3Haiku   string
	Claude35Sonnet string
	Claude37Sonnet string
	Claude4Sonnet  string
	Claude4Opus    string
	Claude45Sonnet string
	Claude45Opus   string
	Claude46Sonnet string
	Claude46Opus   string
}

func (m AnthropicModelList) ListModels() []string {
	return []string{m.Claude3Haiku, m.Claude35Sonnet, m.Claude37Sonnet, m.Claude4Sonnet, m.Claude4Opus, m.Claude45Sonnet, m.Claude45Opus, m.Claude46Sonnet, m.Claude46Opus}
}

// AnthropicModels contains the predefined Anthropic model strings.
var AnthropicModels = AnthropicModelList{
	Claude3Haiku:   "claude-3-haiku-20240307",
	Claude35Sonnet: "claude-3-5-sonnet-latest",
	Claude37Sonnet: "claude-3-7-sonnet-latest",
	Claude4Sonnet:  "claude-sonnet-4-20250514",
	Claude4Opus:    "claude-opus-4-20250514",
	Claude45Sonnet: "claude-sonnet-4-5-20250620",
	Claude45Opus:   "claude-opus-4-5-20250620",
	Claude46Sonnet: "claude-sonnet-4-6-20250827",
	Claude46Opus:   "claude-opus-4-6-20250827",
}

// --- Gemini Models ---

// GeminiModelList holds available Google Gemini model identifiers.
type GeminiModelList struct {
	Gemini20Flash    string
	Gemini20FlashExp string
	Gemini25Pro      string
	Gemini25Flash    string
	Gemini3Flash     string
	Gemini3Pro       string
}

func (m GeminiModelList) ListModels() []string {
	return []string{m.Gemini20Flash, m.Gemini20FlashExp, m.Gemini25Pro, m.Gemini25Flash, m.Gemini3Flash, m.Gemini3Pro}
}

// GeminiModels contains the predefined Gemini model strings.
var GeminiModels = GeminiModelList{
	Gemini20Flash:    "gemini-2.0-flash",
	Gemini20FlashExp: "gemini-2.0-flash-exp",
	Gemini25Pro:      "gemini-2.5-pro",
	Gemini25Flash:    "gemini-2.5-flash",
	Gemini3Flash:     "gemini-3.0-flash",
	Gemini3Pro:       "gemini-3.0-pro",
}

// --- Ollama Models ---

// OllamaModelList holds common Ollama model identifiers.
type OllamaModelList struct {
	Llama3  string
	Llama31 string
	Mistral string
	Mixtral string
	Phi3    string
	Gemma2  string
}

func (m OllamaModelList) ListModels() []string {
	return []string{m.Llama3, m.Llama31, m.Mistral, m.Mixtral, m.Phi3, m.Gemma2}
}

// OllamaModels contains common Ollama model strings. Users can also pass any
// custom model name string that is available on their Ollama instance.
var OllamaModels = OllamaModelList{
	Llama3:  "llama3",
	Llama31: "llama3.1",
	Mistral: "mistral",
	Mixtral: "mixtral",
	Phi3:    "phi3",
	Gemma2:  "gemma2",
}

// availableModels maps providers to their model lists.
// Read-only after init. Do not modify at runtime.
var availableModels = map[string]Models{
	ProviderOpenAi:    OpenAIModels,
	ProviderAzure:     OpenAIModels,
	ProviderAnthropic: AnthropicModels,
	ProviderGemini:    GeminiModels,
	ProviderOllama:    OllamaModels,
}

// checkModel validates that the given model exists for the provider.
func checkModel(provider, modelName string) (bool, string) {
	models, exists := availableModels[provider]
	if !exists {
		return false, fmt.Sprintf("provider %q is not registered", provider)
	}

	// Ollama allows any model name since users can pull custom models.
	if provider == ProviderOllama {
		return true, "model accepted"
	}

	for _, model := range models.ListModels() {
		if model == modelName {
			return true, "model exists"
		}
	}
	return false, fmt.Sprintf("model %q does not exist. Available models for %s: %v", modelName, provider, models.ListModels())
}

// buildSystemPrompts creates the standard system prompt slice from an Agent.
func buildSystemPrompts(a *Agent) []agentPrompts {
	return []agentPrompts{
		{
			Role:    agentRoleSystem,
			Context: fmt.Sprintf("As a %s, %s", a.Role, a.Backstory),
		},
		{
			Role:    agentRoleSystem,
			Context: fmt.Sprintf("Your goal is %s", a.Goal),
		},
	}
}

// resolveUserPrompt validates and builds the final user prompt from the stored
// prompt and optional context parameters.
func resolveUserPrompt(userPrompt *string, params []string) (string, error) {
	if userPrompt == nil {
		return "", ErrMissingPrompt
	}
	prompt := *userPrompt
	if len(params) > 1 {
		return "", ErrTooManyArgs
	}
	if len(params) == 1 {
		prompt = prompt + "\n\nTake in consideration the following context: " + params[0]
	}
	return prompt, nil
}

// extractBuiltinToolInput extracts the "input" field from a JSON tool argument
// string, used for builtin tools that wrap their argument in {"input": "..."}.
func extractBuiltinToolInput(rawInput string) string {
	inputMap := make(map[string]any)
	if err := json.Unmarshal([]byte(rawInput), &inputMap); err == nil {
		if v, ok := inputMap["input"]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return rawInput
}
