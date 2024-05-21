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

func NewAgent(context string) *agent {
	adaptor := newAdaptor()
	initialPrompt := &[]AgentPrompts{
		{
			Role:    AgentRoleSystem,
			Context: context,
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

func (a *agent) WithContext(context string) {
	(*a.prompts)[0].Context = context
}

func (a *agent) WithRole(role string) {
	(*a.prompts)[0].Role = role
}
