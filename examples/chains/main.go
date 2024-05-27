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

	marketAnalystAgent := forza.NewAgent()
	marketAnalystAgent.
		WithRole("Lead Market Analyst at a premier digital marketing firm").
		WithBackstory("you specialize in dissecting online business landscapes. Conduct amazing analysis of the products and competitors").
		WithGoal("providing in-depth insights to guide marketing strategies")

	task1 := marketAnalystAgent.NewLLMTask(config)
	task1.WithUserPrompt("Give me a full report about the market of electric cars in the US.")

	contentCreatorAgent := forza.NewAgent()
	contentCreatorAgent.
		WithRole("Creative Content Creator at a top-tier digital marketing agency").
		WithBackstory("you excel in crafting narratives that resonate with audiences on social media. Your expertise lies in turning marketing strategies into engaging stories and visual content that capture attention and inspire action").
		WithGoal("Generate a creative social media post for a new line of eco-friendly products")

	task2 := contentCreatorAgent.NewLLMTask(config)
	task2.WithUserPrompt("Generate a creative social media post for a new line of eco-friendly products.")

	f := forza.NewPipeline()
	chain := f.CreateChain(task1.Completion, task2.Completion)

	fmt.Println("Chain result: ", chain())
}
