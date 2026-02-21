package forza

import (
	"testing"
)

func TestNewLLMConfig_Defaults(t *testing.T) {
	config := NewLLMConfig()

	if config.temperature != 0.3 {
		t.Errorf("expected default temperature 0.3, got %f", config.temperature)
	}
	if config.maxTokens != 4096 {
		t.Errorf("expected default maxTokens 4096, got %d", config.maxTokens)
	}
	if config.provider != "" {
		t.Errorf("expected empty provider, got %q", config.provider)
	}
	if config.model != "" {
		t.Errorf("expected empty model, got %q", config.model)
	}
}

func TestLLMConfig_BuilderChain(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithTemperature(0.7).
		WithMaxTokens(2048).
		WithOpenAiCredentials("test-key")

	if config.provider != ProviderOpenAi {
		t.Errorf("expected provider %q, got %q", ProviderOpenAi, config.provider)
	}
	if config.model != OpenAIModels.GPT4oMini {
		t.Errorf("expected model %q, got %q", OpenAIModels.GPT4oMini, config.model)
	}
	if config.temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", config.temperature)
	}
	if config.maxTokens != 2048 {
		t.Errorf("expected maxTokens 2048, got %d", config.maxTokens)
	}
	if config.credentials.apiKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %q", config.credentials.apiKey)
	}
}

func TestLLMConfig_AzureCredentials(t *testing.T) {
	config := NewLLMConfig().
		WithAzureOpenAiCredentials("azure-key", "https://my-endpoint.openai.azure.com/")

	if config.credentials.apiKey != "azure-key" {
		t.Errorf("expected apiKey 'azure-key', got %q", config.credentials.apiKey)
	}
	if config.credentials.endpoint != "https://my-endpoint.openai.azure.com/" {
		t.Errorf("expected endpoint, got %q", config.credentials.endpoint)
	}
}

func TestLLMConfig_AnthropicCredentials(t *testing.T) {
	config := NewLLMConfig().
		WithAnthropicCredentials("anthropic-key")

	if config.credentials.apiKey != "anthropic-key" {
		t.Errorf("expected apiKey 'anthropic-key', got %q", config.credentials.apiKey)
	}
}

func TestLLMConfig_GeminiCredentials(t *testing.T) {
	config := NewLLMConfig().
		WithGeminiCredentials("gemini-key")

	if config.credentials.apiKey != "gemini-key" {
		t.Errorf("expected apiKey 'gemini-key', got %q", config.credentials.apiKey)
	}
}

func TestLLMConfig_OllamaCredentials(t *testing.T) {
	config := NewLLMConfig().
		WithOllamaCredentials("http://localhost:11434/v1")

	if config.credentials.endpoint != "http://localhost:11434/v1" {
		t.Errorf("expected endpoint, got %q", config.credentials.endpoint)
	}
}

func TestLLMConfig_CredentialsOverwrite(t *testing.T) {
	// Setting credentials for one provider should overwrite the previous
	config := NewLLMConfig().
		WithOpenAiCredentials("openai-key").
		WithAnthropicCredentials("anthropic-key")

	if config.credentials.apiKey != "anthropic-key" {
		t.Errorf("expected credentials to be overwritten, got %q", config.credentials.apiKey)
	}
}
