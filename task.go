package forza

type Task struct {
	agent *Agent
}

func NewTask(agent *Agent) *Task {
	return &Task{
		agent: agent,
	}
}

func (t *Task) SetFunction(ops ...interface{}) {
	t.agent.adaptor.SetFunction(ops...)
}
