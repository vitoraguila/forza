package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vitoraguila/forza"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderOpenAi).
		WithModel(forza.OpenAIModels.GPT4oMini).
		WithOpenAiCredentials(os.Getenv("OPENAI_API_KEY"))

	agentWriter := forza.NewAgent().
		WithRole("You are famous writer").
		WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination.").
		WithGoal("building a compelling narrative")

	task, err := agentWriter.NewLLMTask(config)
	if err != nil {
		log.Fatal(err)
	}
	task.WithUserPrompt("Write a story about Hercules and the Hydra")

	result, err := task.Completion(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("result TASK: ", result)
}
