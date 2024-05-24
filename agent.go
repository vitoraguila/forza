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
