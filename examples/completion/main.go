package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	agentsConfig := forza.NewAgentConfig(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

	agentWriter, err := forza.NewAgent(
		forza.WithRole("You are famous writer"),
		forza.WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination."),
		forza.WithGoal("building a compelling narrative"),
		forza.WithConfig(agentsConfig),
	)
	forza.IsError(err)

	task := forza.NewTask(agentWriter)
	task.Instruction("the character is Hercules")

	result := task.Completion()
	fmt.Println("result TASK: ", result)
}
