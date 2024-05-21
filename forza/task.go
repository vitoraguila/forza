package forza

import utils "github.com/vitoraguila/forza/internal"

type Task struct {
	agent  *Agent
	prompt string
}

type TaskService interface {
	SetPrompt(prompt string)
	Completion() string
	checkAgentConfiguration()
	SetFunction(name string, description string, params utils.FunctionShape, fn func(param string) string)
}

func NewTask(agent *Agent) TaskService {
	return &Task{
		agent: agent,
	}
}

func (t *Task) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *Task) checkAgentConfiguration() {
	if !(t.agent.IsConfigured) {
		panic("agent is not configured. Missing provider and model")
	}
}

func (t *Task) Completion() string {
	t.checkAgentConfiguration()
	if t.prompt == "" {
		panic("no prompt set")
	}

	return t.agent.adaptor.Completion(t.prompt, t.agent.prompts)
}

func (t *Task) SetFunction(name string, description string, params utils.FunctionShape, fn func(param string) string) {
	t.agent.adaptor.SetFunction(name, description, params, fn)
}
