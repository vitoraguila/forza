package forza

import "fmt"

type adaptorService interface {
	configure(provider string, model string)
	setFunction(name string, description string, params FunctionShape, fn func(param string) string)
	completion(prompt string, prompts *[]AgentPrompts) string
}
type adaptor struct {
	provider string
	model    string
	Service  interface{}
}

func newAdaptor() adaptorService {
	return &adaptor{}
}

func (a *adaptor) configure(provider string, model string) {
	switch provider {
	case ProviderOpenAi, ProviderAzure:
		a.provider = provider
		a.model = model
		a.Service = newOpenAI()
		a.Service.(openAIService).withModel(model)
	default:
		fmt.Printf("provider does not exist")
	}
}

func (a *adaptor) completion(prompt string, prompts *[]AgentPrompts) string {
	if a.provider == "" && a.Service == nil && a.model == "" {
		panic("Please configure the adaptor first")
	}

	switch provider := a.provider; provider {
	case ProviderOpenAi:
		return a.Service.(openAIService).completion(prompt, prompts)
	default:
		fmt.Printf("provider does not exist")
		return ""
	}
}

func (a *adaptor) setFunction(name string, description string, params FunctionShape, fn func(param string) string) {
	if a.provider == "" {
		panic("no provider selected")
	}

	switch provider := a.provider; provider {
	case ProviderOpenAi:
		jsonschema := generateSchema(params)
		a.Service.(openAIService).addFunction(name, description, jsonschema, fn)
	default:
		fmt.Printf("provider does not exist")
	}
}
