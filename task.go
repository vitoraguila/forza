package forza

import "fmt"

type task struct {
	agent  *agent
	prompt string
}

func NewTask(agent *agent) *task {
	return &task{
		agent: agent,
	}
}

func (t *task) Instruction(prompt string) {
	t.prompt = prompt
}

func (t *task) checkAgentConfiguration() {
	if !(t.agent.IsConfigured) {
		panic("agent is not configured. Missing provider and model")
	}
}

func (t *task) Completion(params ...string) string {
	var context string

	if len(params) > 1 {
		panic("Error: too many arguments. Only one optional argument(context) is allowed.")
	}

	if len(params) == 1 {
		fmt.Println("params: ", params)
		context = params[0]
	}

	t.checkAgentConfiguration()
	if t.prompt == "" {
		panic("no prompt set")
	}

	var prompt string = t.prompt
	if context != "" {
		prompt = prompt + "\n\n take in consideration the following context: " + context
	}

	return t.agent.adaptor.completion(prompt, t.agent.prompts)
}

func (t *task) SetFunction(name string, description string, params FunctionShape, fn func(param string) string) {
	t.agent.adaptor.setFunction(name, description, params, fn)
}
