package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	config := forza.NewLLMConfig(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

	agentWriter := forza.NewAgent().
		WithRole("You are famous writer").
		WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination.").
		WithGoal("building a compelling narrative")

	task := forza.NewTask(agentWriter).WithLLM(config)
	task.WithUserPrompt("Write a story about Hercules and the Hydra")

	result := task.Completion()
	fmt.Println("result TASK: ", result)
}
