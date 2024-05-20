package adaptors

import (
	"fmt"

	utils "github.com/vitoraguila/forza/internal"
)

type AdaptorService interface {
	WithOpenAI(model string)
	SetFunction(name string, description string, params interface{})
	Completion(prompt string, prompts *[]utils.AgentPrompts) string
}

type AdaptorFuncParams struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

type Adaptor struct {
	provider string
	Service  interface{}
}

func NewAdaptor() AdaptorService {
	return &Adaptor{}
}

func (a *Adaptor) WithOpenAI(model string) {
	a.provider = "openai"
	a.Service = NewOpenAI()
	a.Service.(OpenAIService).WithModel(model)
}

func (a *Adaptor) Completion(prompt string, prompts *[]utils.AgentPrompts) string {
	switch provider := a.provider; provider {
	case "openai":
		return a.Service.(OpenAIService).Completion(prompt, prompts)
	default:
		fmt.Printf("provider does not exist")
		return ""
	}
}

func (a *Adaptor) SetFunction(name string, description string, params interface{}) {
	if a.provider == "" {
		panic("no provider selected")
	}

	requiredFields := utils.GetRequiredFields(params)

	switch provider := a.provider; provider {
	case "openai":
		p := a.Service.(OpenAIService).CreateFuncParams(params, requiredFields)
		a.Service.(OpenAIService).AddFunction(name, description, p)
	default:
		fmt.Printf("provider does not exist")
	}
}
