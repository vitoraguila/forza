package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	agentStoryTeller := forza.NewAgent("you are a storyteller, write as a fairy tale for kids")
	agentStoryTeller.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task := forza.NewTask(agentStoryTeller)
	task.SetPrompt("who is Hercules?")

	result := task.Completion()
	fmt.Println(result, "result TASK")
}
