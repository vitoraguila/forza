package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	agentStoryTeller := forza.NewAgent(&forza.AgentPersona{
		Role:      "storyteller",
		Backstory: "you are a storyteller, write as a fairy tale for kids",
		Goal:      "write a fairy tale",
	})
	agentStoryTeller.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task := forza.NewTask(agentStoryTeller)
	task.Instruction("who is Hercules?")

	result := task.Completion()
	fmt.Println("result TASK: ", result)
}
