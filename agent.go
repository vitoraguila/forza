package forza

type agent struct {
	Role      string
	Backstory string
	Goal      string
}

type agentConfig struct {
	provider string
	model    string
}

func NewAgent() *agent {
	return &agent{}
}

func NewAgentConfig(provider, model string) *agentConfig {
	return &agentConfig{
		provider: provider,
		model:    model,
	}
}

func (a *agent) WithRole(role string) *agent {
	a.Role = role

	return a
}

func (a *agent) WithBackstory(backstory string) *agent {
	a.Backstory = backstory

	return a
}

func (a *agent) WithGoal(goal string) *agent {
	a.Goal = goal

	return a
}

func (a *agent) NewLLMTask(c *llmConfig) llmAgent {
	if a.Role == "" || a.Backstory == "" || a.Goal == "" {
		panic("Agent Role(WithRole()), Backstory(WithBackstory()) and Goal(WithGoal) are required")
	}

	isModelExist, msg := checkModel(c.provider, c.model)
	if !isModelExist {
		panic(msg)
	}
	switch c.provider {
	case ProviderOpenAi, ProviderAzure:
		return NewOpenAI(c, a)
	default:
		panic("provider does not exist")
	}
}
