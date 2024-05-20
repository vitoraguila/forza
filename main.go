package main

import (
	"fmt"

	"github.com/vitoraguila/forza/forza"
)

type Properties struct {
	UserId string `json:"userId" required:"true"`
}

func main() {
	ag1 := forza.NewAgent("openai", "you are a historian, give the perspective history", "", "system")
	task1 := forza.NewTask(ag1)
	task1.SetPrompt("who is barack obama?")

	ag2 := forza.NewAgent("openai", "you are a storyteller, write as a fairy tale for kids", "", "system")
	task2 := forza.NewTask(ag2)
	task2.SetPrompt("who is Hercules?")

	// func1Params := &Properties{
	// 	UserId: "123",
	// }

	f := forza.NewForza()
	f.AddTask(task1.Completion)
	f.AddTask(task2.Completion)

	result := f.RunConcurrently()

	fmt.Println(result[0], "result TASK1")
	fmt.Println("-----------------")
	fmt.Println(result[1], "result TASK2")

	// task1.Completion()
	// task1.SetFunction("get_user_id", "user will provide an userId, identify and get this userId", func1Params)

}
