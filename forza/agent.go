package forza

import (
	"fmt"
	"strings"

	utils "github.com/vitoraguila/forza/internal"
	"github.com/vitoraguila/forza/internal/adaptors"
)

type Agent struct {
	adaptor      adaptors.AdaptorService
	IsConfigured bool
	prompts      *[]utils.AgentPrompts
}

func NewAgent(backstory string, goal string) *Agent {
	adaptor := adaptors.NewAdaptor()
	initialPrompt := &[]utils.AgentPrompts{
		{
			Role:      AgentRoleSystem,
			Goal:      goal,
			Backstory: backstory,
		},
	}

	return &Agent{
		adaptor:      adaptor,
		prompts:      initialPrompt,
		IsConfigured: false,
	}
}

func (a *Agent) Configure(provider string, model string) {
	if !utils.CheckProvider(provider) {
		panic(fmt.Sprintf("Provider %s does not exist. Providers available are: %s\n", provider, strings.Join(utils.ListProviders, ", ")))
	}

	isModelExist, msg := utils.CheckModel(provider, model)
	if !isModelExist {
		panic(msg)
	}

	a.adaptor.Configure(provider, model)
	a.IsConfigured = true
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
