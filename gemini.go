package forza

import "github.com/vitoraguila/forza/tools"

type gemini struct {
	Config *llmConfig
}

func Newgemini(c *llmConfig, a *agent) llmAgent {
	return &gemini{
		Config: c,
	}
}

func (o *gemini) AddCustomTools(name string, description string, params functionShape, fn func(param string) (string, error)) {
	panic("gemini AddCustomTools not implemented")
}

func (g *gemini) Completion(params ...string) string {
	panic("gemini completion not implemented")
}

func (g *gemini) WithTools(tools ...tools.Tool) {
	panic("gemini withTools not implemented")
}

func (g *gemini) WithUserPrompt(prompt string) {
	panic("gemini WithUserPrompt not implemented")
}
