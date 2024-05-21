package utils

import (
	"fmt"

	"github.com/sashabaranov/go-openai/jsonschema"
)

type AgentPrompts struct {
	Role      string
	Goal      string
	Backstory string
}

type FunctionProps struct {
	Description string
	Required    bool
}

const (
	ProviderOpenAi = "openai"
	ProviderAzure  = "openai-azure"
)

var ListProviders = []string{
	ProviderOpenAi,
	ProviderAzure,
}

type Models interface {
	ListModels() []string
}
type openAIModels struct {
	Gpt35turbo string
	Gpt4       string
	Gpt4o      string
}

func (m openAIModels) ListModels() []string {
	return []string{m.Gpt35turbo, m.Gpt4, m.Gpt4o}
}

var OpenAIModels = openAIModels{
	Gpt35turbo: "gpt-3.5-turbo",
	Gpt4:       "gpt-4",
	Gpt4o:      "gpt-4o",
}

// Map of providers to their models
var availableModels = map[string]Models{
	ProviderOpenAi: OpenAIModels,
	ProviderAzure:  OpenAIModels, // Assuming Azure has the same models
}

type FunctionShape map[string]FunctionProps

func CheckModel(provider, modelName string) (bool, string) {
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

func CheckProvider(provider string) bool {
	switch provider {
	case ProviderOpenAi, ProviderAzure:
		return true
	default:
		return false
	}
}

func GenerateSchema(shape FunctionShape) jsonschema.Definition {
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
