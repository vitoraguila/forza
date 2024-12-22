package forza

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type agentPrompts struct {
	Role    string
	Context string
}

const (
	ProviderOpenAi = "openai"
	ProviderAzure  = "openai-azure"
)

const (
	agentRoleSystem = "system"
	agentRoleUser   = "user"
)

type models interface {
	ListModels() []string
}
type openAIModels struct {
	Gpt35turbo string
	Gpt4       string
	Gpt4o      string
	Gpt4turbo  string
	GPT4oMini  string
}

func (m openAIModels) ListModels() []string {
	return []string{m.Gpt35turbo, m.Gpt4, m.Gpt4o, m.Gpt4turbo, m.GPT4oMini}
}

var OpenAIModels = openAIModels{
	Gpt35turbo: openai.GPT3Dot5Turbo,
	Gpt4:       openai.GPT4,
	Gpt4o:      openai.GPT4o,
	Gpt4turbo:  openai.GPT4Turbo,
	GPT4oMini:  openai.GPT4oMini,
}

// Map of providers to their models
var availableModels = map[string]models{
	ProviderOpenAi: OpenAIModels,
	ProviderAzure:  OpenAIModels, // Assuming Azure has the same models
}

func checkModel(provider, modelName string) (bool, string) {
	models, exists := availableModels[provider]
	if !exists {
		return false, fmt.Sprintf("model %s does not exist. odels available for provider selected are: %s\n", modelName, models.ListModels())
	}
	for _, model := range models.ListModels() {
		if model == modelName {
			return true, "model exists"
		}
	}
	return false, fmt.Sprintf("model %s does not exist. odels available for provider selected are: %s\n", modelName, models.ListModels())
}

func generateSchema(shape functionShape) jsonschema.Definition {
	properties := make(map[string]jsonschema.Definition)
	var required []string

	// Populate the properties with the provided definitions
	for fieldName, props := range shape {
		properties[fieldName] = jsonschema.Definition{
			Type:        jsonschema.String,
			Description: props.Description,
		}

		if props.Required {
			required = append(required, fieldName)
		}
	}

	// Construct and return the schema definition
	return jsonschema.Definition{
		Type:       jsonschema.Object,
		Properties: properties,
		Required:   required,
	}
}
