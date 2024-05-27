package main

import (
	"fmt"
	"os"

	"github.com/vitoraguila/forza"
)

func main() {
	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderOpenAi).
		WithModel(forza.OpenAIModels.Gpt35turbo).
		WithOpenAiCredentials(os.Getenv("OPENAI_API_KEY"))

	agentWriter := forza.NewAgent().
		WithRole("You are famous writer").
		WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination.").
		WithGoal("building a compelling narrative")

	task := agentWriter.NewLLMTask(config)
	task.WithUserPrompt("Write a story about Hercules and the Hydra")

	result := task.Completion()
	fmt.Println("result TASK: ", result)
}
