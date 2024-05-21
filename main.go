package main

import (
	"encoding/json"
	"fmt"

	"github.com/vitoraguila/forza/forza"
	utils "github.com/vitoraguila/forza/internal"
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

	return fmt.Sprintf("The user id is %s", UserParams.UserId)
}

func main() {
	// FIRST EXAMPLE
	ag1 := forza.NewAgent("you are a historian, give the perspective history", "")
	ag1.Configure(forza.ProviderOpenAi, utils.OpenAIModels.Gpt35turbo)
	task1 := forza.NewTask(ag1)
	task1.SetPrompt("who is barack obama?")

	// SECOND EXAMPLE
	ag2 := forza.NewAgent("you are a storyteller, write as a fairy tale for kids", "")
	ag2.Configure(forza.ProviderOpenAi, utils.OpenAIModels.Gpt35turbo)
	task2 := forza.NewTask(ag2)
	task2.SetPrompt("who is Hercules?")

	// THRID EXAMPLE
	func3Params := utils.FunctionShape{
		"userId": utils.FunctionProps{
			Description: "user id description",
			Required:    true,
		},
	}

	// func3Params := &Properties{UserId: "user id description"}
	ag3 := forza.NewAgent("You are a specialist to identify userId in the text", "")
	ag3.Configure(forza.ProviderOpenAi, utils.OpenAIModels.Gpt35turbo)
	task3 := forza.NewTask(ag3)
	task3.SetPrompt("My name is robert and my user id is 3434")
	task3.SetFunction("get_user_id", "user will provide an userId, identify and get this userId", func3Params, getUserId)

	// RUNNING INDIVIDUALLY
	// result := task3.Completion()
	// fmt.Println(result, "result TASK3")

	// RUNNING ALL CONCURRENTLY

	f := forza.NewForza()
	f.AddTask(task1.Completion)
	f.AddTask(task2.Completion)
	f.AddTask(task3.Completion)
	result := f.RunConcurrently()

	fmt.Println(result[0], "result TASK1")
	fmt.Println("-----------------")
	fmt.Println(result[1], "result TASK2")
	fmt.Println("-----------------")
	fmt.Println(result[2], "result TASKr")
}
