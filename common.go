package forza

import "fmt"

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
}

func (m OpenAIModelList) ListModels() []string {
	return []string{m.GPT35Turbo, m.GPT4, m.GPT4o, m.GPT4Turbo, m.GPT4oMini, m.O1Mini, m.O1, m.GPT5}
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
}

// --- Anthropic Models ---

// AnthropicModelList holds available Anthropic model identifiers.
type AnthropicModelList struct {
	Claude3Haiku   string
	Claude35Sonnet string
	Claude37Sonnet string
	Claude4Sonnet  string
	Claude4Opus    string
}

func (m AnthropicModelList) ListModels() []string {
	return []string{m.Claude3Haiku, m.Claude35Sonnet, m.Claude37Sonnet, m.Claude4Sonnet, m.Claude4Opus}
}

// AnthropicModels contains the predefined Anthropic model strings.
var AnthropicModels = AnthropicModelList{
	Claude3Haiku:   "claude-3-haiku-20240307",
	Claude35Sonnet: "claude-3-5-sonnet-latest",
	Claude37Sonnet: "claude-3-7-sonnet-latest",
	Claude4Sonnet:  "claude-sonnet-4-20250514",
	Claude4Opus:    "claude-opus-4-20250514",
}

// --- Gemini Models ---

// GeminiModelList holds available Google Gemini model identifiers.
type GeminiModelList struct {
	Gemini20Flash    string
	Gemini20FlashExp string
	Gemini25Pro      string
	Gemini25Flash    string
}

func (m GeminiModelList) ListModels() []string {
	return []string{m.Gemini20Flash, m.Gemini20FlashExp, m.Gemini25Pro, m.Gemini25Flash}
}

// GeminiModels contains the predefined Gemini model strings.
var GeminiModels = GeminiModelList{
	Gemini20Flash:    "gemini-2.0-flash",
	Gemini20FlashExp: "gemini-2.0-flash-exp",
	Gemini25Pro:      "gemini-2.5-pro",
	Gemini25Flash:    "gemini-2.5-flash",
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
