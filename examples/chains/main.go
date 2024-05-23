package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	marketAnalystAgent := forza.NewAgent(&forza.AgentPersona{
		Role:      "Lead Market Analyst at a premier digital marketing firm",
		Backstory: "you specialize in dissecting online business landscapes. Conduct amazing analysis of the products and competitors",
		Goal:      "providing in-depth insights to guide marketing strategies",
	})
	marketAnalystAgent.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task1 := forza.NewTask(marketAnalystAgent)
	task1.Instruction("Give me a full report about the market of electric cars in the US.")

	contentCreatorAgent := forza.NewAgent(&forza.AgentPersona{
		Role:      "Creative Content Creator at a top-tier digital marketing agency",
		Backstory: "you excel in crafting narratives that resonate with audiences on social media. Your expertise lies in turning marketing strategies into engaging stories and visual content that capture attention and inspire action",
		Goal:      "Generate a creative social media post for a new line of eco-friendly products"},
	)

	contentCreatorAgent.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task2 := forza.NewTask(contentCreatorAgent)
	task2.Instruction("Generate a creative social media post for a new line of eco-friendly products.")

	// RUNNING ALL CONCURRENTLY
	f := forza.NewPipeline()
	chain := f.CreateChain(*task1.WithCompletion(), *task2.WithCompletion())

	fmt.Println("Chain result: ", chain())
}
