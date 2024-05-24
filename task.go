package forza

type task struct {
	agent  *agent
	prompt string
}

type taskService interface {
	WithLLM(c *llmConfig) llmService
}

func NewTask(agent *agent) taskService {
	return &task{
		agent: agent,
	}
}

func (t *task) WithLLM(c *llmConfig) llmService {
	isModelExist, msg := checkModel(c.provider, c.model)
	if !isModelExist {
		panic(msg)
	}
	switch c.provider {
	case ProviderOpenAi, ProviderAzure:
		return NewOpenAI(c, t)
	default:
		panic("provider does not exist")
	}
}

func (t *task) WithInstruction(prompt string) {
	t.prompt = prompt
}
