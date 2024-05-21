package forza

type task struct {
	agent  *agent
	prompt string
}

func NewTask(agent *agent) *task {
	return &task{
		agent: agent,
	}
}

func (t *task) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *task) checkAgentConfiguration() {
	if !(t.agent.IsConfigured) {
		panic("agent is not configured. Missing provider and model")
	}
}

func (t *task) Completion() string {
	t.checkAgentConfiguration()
	if t.prompt == "" {
		panic("no prompt set")
	}

	return t.agent.adaptor.completion(t.prompt, t.agent.prompts)
}

func (t *task) SetFunction(name string, description string, params FunctionShape, fn func(param string) string) {
	t.agent.adaptor.setFunction(name, description, params, fn)
}
