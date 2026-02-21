package forza

import "fmt"

// Agent represents an AI agent with a role, backstory, and goal.
type Agent struct {
	Role      string
	Backstory string
	Goal      string
}

// NewAgent creates a new empty Agent.
func NewAgent() *Agent {
	return &Agent{}
}

// WithRole sets the agent's role.
func (a *Agent) WithRole(role string) *Agent {
	a.Role = role
	return a
}

// WithBackstory sets the agent's backstory.
func (a *Agent) WithBackstory(backstory string) *Agent {
	a.Backstory = backstory
	return a
}

// WithGoal sets the agent's goal.
func (a *Agent) WithGoal(goal string) *Agent {
	a.Goal = goal
	return a
}

// providerFactory maps provider names to their constructor functions.
var providerFactory = map[string]func(*LLMConfig, *Agent) LLMAgent{
	ProviderOpenAi:    newOpenAI,
	ProviderAzure:     newOpenAI,
	ProviderAnthropic: newAnthropic,
	ProviderGemini:    newGemini,
	ProviderOllama:    newOllama,
}

// NewLLMTask creates an LLMAgent for this agent using the provided configuration.
// Returns an error if the agent is incomplete or the provider/model is invalid.
func (a *Agent) NewLLMTask(c *LLMConfig) (LLMAgent, error) {
	if a.Role == "" {
		return nil, ErrMissingRole
	}
	if a.Backstory == "" {
		return nil, ErrMissingBackstory
	}
	if a.Goal == "" {
		return nil, ErrMissingGoal
	}

	ok, msg := checkModel(c.provider, c.model)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrModelNotFound, msg)
	}

	factory, exists := providerFactory[c.provider]
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrProviderNotFound, c.provider)
	}

	return factory(c, a), nil
}
