package forza

import (
	"errors"
	"testing"
	"time"
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
	if config.timeout != defaultTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultTimeout, config.timeout)
	}
	if config.maxRetries != defaultMaxRetries {
		t.Errorf("expected default maxRetries %d, got %d", defaultMaxRetries, config.maxRetries)
	}
}

func TestLLMConfig_BuilderChain(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithTemperature(0.7).
		WithMaxTokens(2048).
		WithOpenAiCredentials("test-key").
		WithTimeout(60 * time.Second).
		WithMaxRetries(5)

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
	if config.timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", config.timeout)
	}
	if config.maxRetries != 5 {
		t.Errorf("expected maxRetries 5, got %d", config.maxRetries)
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

func TestLLMConfig_Validate_Valid(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini)

	if err := config.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestLLMConfig_Validate_EmptyProvider(t *testing.T) {
	config := NewLLMConfig().
		WithModel(OpenAIModels.GPT4oMini)

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for empty provider")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLLMConfig_Validate_EmptyModel(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi)

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for empty model")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLLMConfig_Validate_TemperatureOutOfRange(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithTemperature(3.0)

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for temperature > 2.0")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLLMConfig_Validate_NegativeTemperature(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithTemperature(-0.1)

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for negative temperature")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLLMConfig_Validate_ZeroMaxTokens(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithMaxTokens(0)

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for zero maxTokens")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}
