package forza

type gemini struct {
	Config *llmConfig
}

func Newgemini(c *llmConfig, t *task) llmService {
	return &gemini{
		Config: c,
	}
}

func (o *gemini) AddFunctions(name string, description string, params functionShape, fn func(param string) string) {
	panic("gemini AddFunctions not implemented")
}

func (g *gemini) Completion(params ...string) string {
	panic("gemini completion not implemented")
}

func (g *gemini) WithUserPrompt(prompt string) {
	panic("gemini WithUserPrompt not implemented")
}
