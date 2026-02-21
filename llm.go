package forza

import (
	"context"
	"fmt"
	"time"

	"github.com/vitoraguila/forza/tools"
)

// LLMAgent is the interface that all LLM provider implementations must satisfy.
type LLMAgent interface {
	// Completion sends the prompt to the LLM and returns the response.
	// An optional context string can be passed (used in chains).
	Completion(ctx context.Context, params ...string) (string, error)

	// AddCustomTools registers a custom function-calling tool.
	AddCustomTools(name string, description string, params FunctionShape, fn func(param string) (string, error))

	// WithUserPrompt sets the user prompt for the next completion.
	WithUserPrompt(prompt string)

	// WithTools registers pre-built tools (e.g. scraper).
	WithTools(tools ...tools.Tool)
}

// credentials holds provider authentication details.
type credentials struct {
	apiKey   string
	endpoint string
}

const (
	defaultTimeout    = 120 * time.Second
	defaultMaxRetries = 3
)

// LLMConfig holds the configuration for an LLM provider.
type LLMConfig struct {
	provider    string
	model       string
	credentials credentials
	temperature float64
	maxTokens   int
	timeout     time.Duration
	maxRetries  int
}

// NewLLMConfig creates a new LLMConfig with sensible defaults.
func NewLLMConfig() *LLMConfig {
	return &LLMConfig{
		temperature: 0.3,
		maxTokens:   4096,
		timeout:     defaultTimeout,
		maxRetries:  defaultMaxRetries,
	}
}

// Validate checks that the configuration is valid.
func (c *LLMConfig) Validate() error {
	if c.provider == "" {
		return fmt.Errorf("%w: provider must not be empty", ErrInvalidConfig)
	}
	if c.model == "" {
		return fmt.Errorf("%w: model must not be empty", ErrInvalidConfig)
	}
	if c.temperature < 0 || c.temperature > 2.0 {
		return fmt.Errorf("%w: temperature must be between 0 and 2.0, got %f", ErrInvalidConfig, c.temperature)
	}
	if c.maxTokens <= 0 {
		return fmt.Errorf("%w: maxTokens must be greater than 0, got %d", ErrInvalidConfig, c.maxTokens)
	}
	return nil
}

// WithTemperature sets the sampling temperature (0.0 - 2.0).
func (c *LLMConfig) WithTemperature(temperature float64) *LLMConfig {
	c.temperature = temperature
	return c
}

// WithMaxTokens sets the maximum number of tokens in the response.
func (c *LLMConfig) WithMaxTokens(maxTokens int) *LLMConfig {
	c.maxTokens = maxTokens
	return c
}

// WithProvider sets the LLM provider (e.g. ProviderOpenAi, ProviderAnthropic).
func (c *LLMConfig) WithProvider(provider string) *LLMConfig {
	c.provider = provider
	return c
}

// WithModel sets the model identifier.
func (c *LLMConfig) WithModel(model string) *LLMConfig {
	c.model = model
	return c
}

// WithTimeout sets the HTTP client timeout for provider requests.
func (c *LLMConfig) WithTimeout(d time.Duration) *LLMConfig {
	c.timeout = d
	return c
}

// WithMaxRetries sets the maximum number of retry attempts for transient errors.
func (c *LLMConfig) WithMaxRetries(n int) *LLMConfig {
	c.maxRetries = n
	return c
}

// WithOpenAiCredentials sets OpenAI API credentials.
func (c *LLMConfig) WithOpenAiCredentials(openAiApiKey string) *LLMConfig {
	c.credentials = credentials{
		apiKey: openAiApiKey,
	}
	return c
}

// WithAzureOpenAiCredentials sets Azure OpenAI credentials.
func (c *LLMConfig) WithAzureOpenAiCredentials(azureApiKey, azureEndpoint string) *LLMConfig {
	c.credentials = credentials{
		apiKey:   azureApiKey,
		endpoint: azureEndpoint,
	}
	return c
}

// WithAnthropicCredentials sets Anthropic API credentials.
func (c *LLMConfig) WithAnthropicCredentials(apiKey string) *LLMConfig {
	c.credentials = credentials{
		apiKey: apiKey,
	}
	return c
}

// WithGeminiCredentials sets Google Gemini API credentials.
func (c *LLMConfig) WithGeminiCredentials(apiKey string) *LLMConfig {
	c.credentials = credentials{
		apiKey: apiKey,
	}
	return c
}

// WithOllamaCredentials sets the Ollama endpoint (default: http://localhost:11434).
func (c *LLMConfig) WithOllamaCredentials(endpoint string) *LLMConfig {
	c.credentials = credentials{
		endpoint: endpoint,
	}
	return c
}
