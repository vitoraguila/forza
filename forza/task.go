package forza

type Task struct {
	agent  *Agent
	prompt string
}

type TaskService interface {
	SetPrompt(prompt string)
	Completion() string
	SetFunction(name string, description string, params interface{})
}

func NewTask(agent *Agent) TaskService {
	return &Task{
		agent: agent,
	}
}

func (t *Task) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *Task) Completion() string {
	if t.prompt == "" {
		panic("no prompt set")
	}
	return t.agent.adaptor.Completion(t.prompt, t.agent.prompts)
}

func (t *Task) SetFunction(name string, description string, params interface{}) {
	t.agent.adaptor.SetFunction(name, description, params)
}
