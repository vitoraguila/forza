package forza

type llmService interface {
	Completion(params ...string) string
	AddFunctions(name string, description string, params functionShape, fn func(param string) string)
	WithUserPrompt(prompt string)
}

type credentials struct {
	openAi openAiCredentials
}

type llmConfig struct {
	provider    string
	model       string
	credentials credentials
	temperature float64
}

func NewLLMConfig() *llmConfig {
	return &llmConfig{
		temperature: 0.3,
	}
}

func (c *llmConfig) WithTempature(temperature float64) *llmConfig {
	c.temperature = temperature
	return c
}

func (c *llmConfig) WithProvider(provider string) *llmConfig {
	c.provider = provider
	return c
}

func (c *llmConfig) WithModel(model string) *llmConfig {
	c.model = model
	return c
}

func (c *llmConfig) WithOpenAiCredentials(openAiApiKey string) *llmConfig {
	c.credentials.openAi = openAiCredentials{
		openAiApiKey: openAiApiKey,
	}

	return c
}

func (c *llmConfig) WithAzureOpenAiCredentials(azureOpenAiApiKey, azureOpenAiEndpoint string) *llmConfig {
	c.credentials.openAi = openAiCredentials{
		azureOpenAiApiKey:   azureOpenAiApiKey,
		azureOpenAiEndpoint: azureOpenAiEndpoint,
	}

	return c
}
