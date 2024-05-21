package forza

type Task struct {
	agent  *Agent
	prompt string
}

func NewTask(agent *Agent) *Task {
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

func (t *Task) SetFunction(name string, description string, params FunctionShape, fn func(param string) string) {
	t.agent.adaptor.SetFunction(name, description, params, fn)
}
