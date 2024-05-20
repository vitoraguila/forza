package forza

import (
	utils "github.com/vitoraguila/forza/internal"
	"github.com/vitoraguila/forza/internal/adaptors"
)

type Agent struct {
	adaptor adaptors.AdaptorService
	prompts *[]utils.AgentPrompts
}

func NewAgent(provider string, backstory string, goal string, role string) *Agent {
	adaptor := adaptors.NewAdaptor()
	initialPrompt := &[]utils.AgentPrompts{
		{
			Role:      role,
			Goal:      goal,
			Backstory: backstory,
		},
	}

	switch provider {
	case "openai":
		adaptor.WithOpenAI("gpt-35-turbo")
		return &Agent{
			adaptor: adaptor,
			prompts: initialPrompt,
		}
	default:
		panic("provider does not exist")
	}
}

func (a *Agent) AddExtraPrompt(backstory string, goal string, role string) {
	*a.prompts = append(*a.prompts, utils.AgentPrompts{
		Role:      role,
		Goal:      goal,
		Backstory: backstory,
	})
}

func (a *Agent) WithBackstory(backstory string) {
	(*a.prompts)[0].Backstory = backstory
}

func (a *Agent) WithGoal(goal string) {
	(*a.prompts)[0].Goal = goal
}

func (a *Agent) WithRole(role string) {
	(*a.prompts)[0].Role = role
}
