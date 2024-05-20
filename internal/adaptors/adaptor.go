package adaptors

import "fmt"

type AdaptorService interface {
	WithOpenAI(model string)
	SetFunction(name string, description string, params *Parameters)
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

func (a *Adaptor) SetFunction(name string, description string, params *Parameters) {
	if a.provider == "" {
		panic("no provider selected")
	}

	switch provider := a.provider; provider {
	case "openai":
		a.Service.(OpenAIService).AddFunction(name, description, params)
	default:
		fmt.Printf("provider does not exist")
	}
}
