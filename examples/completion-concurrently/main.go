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

	marketAnalystAgent := forza.NewAgent().
		WithRole("Lead Market Analyst at a premier digital marketing firm").
		WithBackstory("you specialize in dissecting online business landscapes. Conduct amazing analysis of the products and competitors").
		WithGoal("providing in-depth insights to guide marketing strategies")

	task1, err := marketAnalystAgent.NewLLMTask(config)
	if err != nil {
		log.Fatal(err)
	}
	task1.WithUserPrompt("Give me a full report about the market of electric cars in the US.")

	contentCreatorAgent := forza.NewAgent().
		WithRole("Creative Content Creator at a top-tier digital marketing agency").
		WithBackstory("you excel in crafting narratives that resonate with audiences on social media. Your expertise lies in turning marketing strategies into engaging stories and visual content that capture attention and inspire action").
		WithGoal("Generate a creative social media post for a new line of eco-friendly products")

	task2, err := contentCreatorAgent.NewLLMTask(config)
	if err != nil {
		log.Fatal(err)
	}
	task2.WithUserPrompt("Generate a creative social media post for a new line of eco-friendly products.")

	// RUNNING ALL CONCURRENTLY
	f := forza.NewPipeline()

	f.AddTasks(task1.Completion, task2.Completion)
	result, err := f.RunConcurrently(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("result TASK1: ", result[0])
	fmt.Println("-----------------")
	fmt.Println("result TASK2: ", result[1])
}
