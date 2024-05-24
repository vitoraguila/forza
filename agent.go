package forza

import (
	"fmt"
	"log"
	"strings"
)

type agent struct {
	adaptor      adaptorService
	IsConfigured bool
	prompts      *[]AgentPrompts
	Role         string
	Backstory    string
	Goal         string
}

type agentConfig struct {
	provider string
	model    string
}

type OptionsAgent func(*agent) error

func IsError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
}

func NewAgent(opts ...OptionsAgent) (*agent, error) {
	agent := &agent{}
	for _, opt := range opts {
		if err := opt(agent); err != nil {
			return nil, err
		}
	}

	if agent.Role == "" || agent.Backstory == "" || agent.Goal == "" {
		return nil, fmt.Errorf("missing required fields: Role(use WithRole() function), Backstory(use WithBackstory() function) and Goal(use WithGoal() function) are required")
	}

	if !agent.IsConfigured {
		return nil, fmt.Errorf("configure your agent with a provider and model before using it. (use WithConfig() function)")
	}

	initialPrompt := &[]AgentPrompts{
		{
			Role:    AgentRoleSystem,
			Context: fmt.Sprintf("As a %s, %s", agent.Role, agent.Backstory),
		},
		{
			Role:    AgentRoleSystem,
			Context: fmt.Sprintf("Your goal is %s", agent.Goal),
		},
	}

	agent.prompts = initialPrompt

	return agent, nil
}

func NewAgentConfig(provider, model string) *agentConfig {
	return &agentConfig{
		provider: provider,
		model:    model,
	}
}

func WithConfig(cfg *agentConfig) OptionsAgent {
	return func(a *agent) error {
		if !checkProvider(cfg.provider) {
			return fmt.Errorf("provider %s does not exist (Providers available are %s)", cfg.provider, strings.Join(ListProviders, ", "))
		}

		isModelExist, msg := checkModel(cfg.provider, cfg.model)
		if !isModelExist {
			return fmt.Errorf(msg)
		}

		adaptor := newAdaptor()
		adaptor.configure(cfg.provider, cfg.model)
		a.adaptor = adaptor

		a.IsConfigured = true

		return nil
	}
}

func WithRole(role string) OptionsAgent {
	return func(ap *agent) error {
		if role == "" {
			return fmt.Errorf("role cannot be empty")
		}
		ap.Role = role
		return nil
	}
}

func WithBackstory(backstory string) OptionsAgent {
	return func(ap *agent) error {
		if backstory == "" {
			return fmt.Errorf("backstory cannot be empty")
		}
		ap.Backstory = backstory
		return nil
	}
}

func WithGoal(goal string) OptionsAgent {
	return func(ap *agent) error {
		if goal == "" {
			return fmt.Errorf("goal cannot be empty")
		}
		ap.Goal = goal
		return nil
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
