package forza

type task struct {
	agent       *agent
	prompt      string
	chainAction taskFn
	chainOutput string
}

func NewTask(agent *agent) *task {
	return &task{
		agent:       agent,
		chainAction: nil,
	}
}

func (t *task) WithCompletion() *task {
	t.chainAction = t.Completion

	return t
}

func (t *task) Instruction(prompt string) {
	t.prompt = prompt
}

func (t *task) checkAgentConfiguration() {
	if !(t.agent.IsConfigured) {
		panic("agent is not configured. Missing provider and model")
	}
}

func (t *task) setChainOutput(o string) {
	t.chainOutput = o
}

func (t *task) Completion() string {
	t.checkAgentConfiguration()
	if t.prompt == "" {
		panic("no prompt set")
	}

	var prompt string = t.prompt
	if t.chainOutput != "" {
		prompt = prompt + "\n\n take in consideration the following info: " + t.chainOutput
	}

	return t.agent.adaptor.completion(prompt, t.agent.prompts)
}

func (t *task) SetFunction(name string, description string, params FunctionShape, fn func(param string) string) {
	t.agent.adaptor.setFunction(name, description, params, fn)
}
