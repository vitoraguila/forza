package adaptors

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	utils "github.com/vitoraguila/forza/internal"
)

type OpenAIParams struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

type OpenAIService interface {
	WithModel(model string)
	AddFunction(name string, description string, params *OpenAIParams)
	CreateFuncParams(params interface{}, required []string) *OpenAIParams
	Completion(prompt string, prompts *[]utils.AgentPrompts) string
}
type OpenAI struct {
	model     string
	Functions []openai.FunctionDefinition
}

func NewOpenAI() OpenAIService {
	return &OpenAI{}
}

func (oai *OpenAI) Completion(prompt string, prompts *[]utils.AgentPrompts) string {
	var messages []openai.ChatCompletionMessage
	for _, p := range *prompts {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    p.Role,
			Content: p.Backstory,
		})
	}

	user := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	messages = append(messages, user)

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return ""
	}
	return resp.Choices[0].Message.Content
}

func (oai *OpenAI) WithModel(model string) {
	oai.model = model
}

func (oai *OpenAI) AddFunction(name string, description string, params *OpenAIParams) {
	oai.Functions = append(oai.Functions, openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  *params,
	})
}

func (oai *OpenAI) CreateFuncParams(params interface{}, required []string) *OpenAIParams {
	return &OpenAIParams{
		Type:       "object",
		Properties: params,
		Required:   required,
	}
}
