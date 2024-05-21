package forza

import "fmt"

type AdaptorService interface {
	Configure(provider string, model string)
	SetFunction(name string, description string, params FunctionShape, fn func(param string) string)
	Completion(prompt string, prompts *[]AgentPrompts) string
}

type AdaptorFuncParams struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

type Adaptor struct {
	provider string
	model    string
	Service  interface{}
}

func NewAdaptor() AdaptorService {
	return &Adaptor{}
}

func (a *Adaptor) Configure(provider string, model string) {
	switch provider {
	case ProviderOpenAi, ProviderAzure:
		a.provider = provider
		a.model = model
		a.Service = NewOpenAI()
		a.Service.(OpenAIService).WithModel(model)
	default:
		fmt.Printf("provider does not exist")
	}
}

func (a *Adaptor) Completion(prompt string, prompts *[]AgentPrompts) string {
	if a.provider == "" && a.Service == nil && a.model == "" {
		panic("Please configure the adaptor first")
	}

	switch provider := a.provider; provider {
	case ProviderOpenAi:
		return a.Service.(OpenAIService).Completion(prompt, prompts)
	default:
		fmt.Printf("provider does not exist")
		return ""
	}
}

func (a *Adaptor) SetFunction(name string, description string, params FunctionShape, fn func(param string) string) {
	if a.provider == "" {
		panic("no provider selected")
	}

	switch provider := a.provider; provider {
	case ProviderOpenAi:
		jsonschema := GenerateSchema(params)
		a.Service.(OpenAIService).AddFunction(name, description, jsonschema, fn)
	default:
		fmt.Printf("provider does not exist")
	}
}