package main

import (
	"encoding/json"
	"fmt"

	"github.com/vitoraguila/forza"
)

type UserParams struct {
	UserId string `json:"userId" required:"true"`
}

func getUserId(params string) string {
	fmt.Println("params: ", params)
	var UserParams UserParams
	err := json.Unmarshal([]byte(params), &UserParams)
	if err != nil {
		panic(err)
	}

	// place any logic here

	return fmt.Sprintf("Answer the exact phrase 'The user id is %s'", UserParams.UserId)
}

func main() {
	agentsConfig := forza.NewAgentConfig(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

	agentSpecialist, err := forza.NewAgent(
		forza.WithRole("Specialist"),
		forza.WithBackstory("you are a specialist"),
		forza.WithGoal("you are a specialist"),
		forza.WithConfig(agentsConfig),
	)
	forza.IsError(err)

	funcCallingParams := forza.FunctionShape{
		"userId": forza.FunctionProps{
			Description: "user id description",
			Required:    true,
		},
	}

	task := forza.NewTask(agentSpecialist)
	task.Instruction("My name is robert and my user id is 3434")
	task.SetFunction("get_user_id", "user will provide an userId, identify and get this userId", funcCallingParams, getUserId)

	result := task.Completion()
	fmt.Println("result TASK: ", result)
}
