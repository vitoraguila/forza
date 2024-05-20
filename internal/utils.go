package utils

import (
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

type FunctionShape map[string]FunctionProps

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
