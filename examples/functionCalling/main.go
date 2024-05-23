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
	var UserParams UserParams
	err := json.Unmarshal([]byte(params), &UserParams)
	if err != nil {
		panic(err)
	}

	// place any logic here

	return fmt.Sprintf("The user id is %s", UserParams.UserId)
}

func main() {
	agentSpecialist := forza.NewAgent(&forza.AgentPersona{
		Role:      "specialist",
		Backstory: "You are a specialist to identify userId in the text",
		Goal:      "identify userId",
	})
	agentSpecialist.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

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
