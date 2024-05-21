package forza

import (
	"fmt"
	"strings"
)

type Agent struct {
	adaptor      AdaptorService
	IsConfigured bool
	prompts      *[]AgentPrompts
}

func NewAgent(context string) *Agent {
	adaptor := NewAdaptor()
	initialPrompt := &[]AgentPrompts{
		{
			Role:    AgentRoleSystem,
			Context: context,
		},
	}

	return &Agent{
		adaptor:      adaptor,
		prompts:      initialPrompt,
		IsConfigured: false,
	}
}

func (a *Agent) Configure(provider string, model string) {
	if !CheckProvider(provider) {
		panic(fmt.Sprintf("Provider %s does not exist. Providers available are: %s\n", provider, strings.Join(ListProviders, ", ")))
	}

	isModelExist, msg := CheckModel(provider, model)
	if !isModelExist {
		panic(msg)
	}

	a.adaptor.Configure(provider, model)
	a.IsConfigured = true
}

func (a *Agent) AddExtraPrompt(context string, role string) {
	*a.prompts = append(*a.prompts, AgentPrompts{
		Role:    role,
		Context: context,
	})
}

func (a *Agent) WithContext(context string) {
	(*a.prompts)[0].Context = context
}

func (a *Agent) WithRole(role string) {
	(*a.prompts)[0].Role = role
}
