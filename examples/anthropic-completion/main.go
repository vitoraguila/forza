package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vitoraguila/forza"
)

func main() {
	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderAnthropic).
		WithModel(forza.AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials(os.Getenv("ANTHROPIC_API_KEY"))

	agentWriter := forza.NewAgent().
		WithRole("You are famous writer").
		WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination.").
		WithGoal("building a compelling narrative")

	task, err := agentWriter.NewLLMTask(config)
	if err != nil {
		log.Fatal(err)
	}
	task.WithUserPrompt("Write a story about Hercules and the Hydra")

	result, err := task.Completion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("result TASK: ", result)
}
