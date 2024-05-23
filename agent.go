package forza

import (
	"fmt"
	"strings"
)

type agent struct {
	adaptor      adaptorService
	IsConfigured bool
	prompts      *[]AgentPrompts
}

type AgentPersona struct {
	Role      string
	Backstory string
	Goal      string
}

func NewAgent(persona *AgentPersona) *agent {
	adaptor := newAdaptor()
	initialPrompt := &[]AgentPrompts{
		{
			Role:    AgentRoleSystem,
			Context: fmt.Sprintf("As a %s, %s", persona.Role, persona.Backstory),
		},
		{
			Role:    AgentRoleSystem,
			Context: fmt.Sprintf("Your goal is %s", persona.Goal),
		},
	}

	return &agent{
		adaptor:      adaptor,
		prompts:      initialPrompt,
		IsConfigured: false,
	}
}

func (a *agent) Configure(provider string, model string) {
	if !checkProvider(provider) {
		panic(fmt.Sprintf("Provider %s does not exist. Providers available are: %s\n", provider, strings.Join(ListProviders, ", ")))
	}

	isModelExist, msg := checkModel(provider, model)
	if !isModelExist {
		panic(msg)
	}

	a.adaptor.configure(provider, model)
	a.IsConfigured = true
}

func (a *agent) AddExtraPrompt(context string, role string) {
	*a.prompts = append(*a.prompts, AgentPrompts{
		Role:    role,
		Context: context,
	})
}
