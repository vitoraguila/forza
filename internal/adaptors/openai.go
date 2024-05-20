package adaptors

import "github.com/sashabaranov/go-openai"

type Parameters struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

// type Properties[T] struct {
// 	UserId string `json:"userId"`
// }

type OpenAIService interface {
	WithModel(model string)
	AddFunction(name string, description string, params *Parameters)
}
type OpenAI struct {
	model     string
	Functions []openai.FunctionDefinition
}

func NewOpenAI() OpenAIService {
	return &OpenAI{}
}

func (oai *OpenAI) Completion() string {
	return ""
}

func (oai *OpenAI) WithModel(model string) {
	oai.model = model
}

func (oai *OpenAI) AddFunction(name string, description string, params *Parameters) {
	oai.Functions = append(oai.Functions, openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  *params,
	})
}

func CreateFuncParams(params interface{}, required []string) *Parameters {
	return &Parameters{
		Type:       "object",
		Properties: params,
		Required:   required,
	}
}
