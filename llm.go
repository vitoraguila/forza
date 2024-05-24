package forza

type llmService interface {
	Completion(params ...string) string
	AddFunctions(name string, description string, params functionShape, fn func(param string) string)
	WithUserPrompt(prompt string)
}

type llmCOnfig struct {
	provider string
	model    string
}

func NewLLMConfig(provider, model string) *llmCOnfig {
	return &llmCOnfig{
		provider: provider,
		model:    model,
	}
}
