package forza

import (
	"github.com/vitoraguila/forza/internal/adaptors"
)

type Agent struct {
	role      string
	adaptor   adaptors.AdaptorService
	goal      string
	backstory string
}

func NewAgent(provider string) *Agent {
	adaptor := adaptors.NewAdaptor()
	switch provider {
	case "openai":
		adaptor.WithOpenAI("gpt-35-turbo")
		return &Agent{
			adaptor: adaptor,
		}
	default:
		panic("provider does not exist")
	}
}

func (a *Agent) WithBackstory(backstory string) {
	a.backstory = backstory
}

func (a *Agent) WithGoal(goal string) {
	a.goal = goal
}

func (a *Agent) WithRole(role string) {
	a.role = role
}

type Properties struct {
	UserId string `json:"userId"`
}

func ma() {
	ag1 := NewAgent("openai")
	task1 := NewTask(ag1)

	func1Params := adaptors.CreateFuncParams(&Properties{
		UserId: "123",
	}, []string{"userId"})

	task1.agent.adaptor.SetFunction("get_user_id", "user will provide an userId, identify and get this userId", func1Params)

}
