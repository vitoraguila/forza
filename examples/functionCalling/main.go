package main

import (
	"encoding/json"
	"fmt"
	"os"

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
	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderOpenAi).
		WithModel(forza.OpenAIModels.Gpt35turbo).
		WithOpenAiCredentials(os.Getenv("OPENAI_API_KEY"))

	agentSpecialist := forza.NewAgent().
		WithRole("Specialist").
		WithBackstory("you are a specialist").
		WithGoal("you are a specialist")

	funcCallingParams := forza.NewFunction(
		forza.WithProperty("userId", "user id description", true),
	)

	task := forza.NewTask(agentSpecialist).WithLLM(config)
	task.WithUserPrompt("My name is robert and my user id is 3434")
	task.AddFunctions("get_user_id", "user will provide an userId, identify and get this userId", funcCallingParams, getUserId)

	result := task.Completion()
	fmt.Println("result TASK: ", result)
}
