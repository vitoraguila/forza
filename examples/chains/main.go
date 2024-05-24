package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	agentsConfig := forza.NewAgentConfig(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	marketAnalystAgent, err := forza.NewAgent(
		forza.WithRole("Lead Market Analyst at a premier digital marketing firm"),
		forza.WithBackstory("you specialize in dissecting online business landscapes. Conduct amazing analysis of the products and competitors"),
		forza.WithGoal("providing in-depth insights to guide marketing strategies"),
		forza.WithConfig(agentsConfig),
	)
	forza.IsError(err)

	task1 := forza.NewTask(marketAnalystAgent)
	task1.Instruction("Give me a full report about the market of electric cars in the US.")

	contentCreatorAgent, err := forza.NewAgent(
		forza.WithRole("Creative Content Creator at a top-tier digital marketing agency"),
		forza.WithBackstory("you excel in crafting narratives that resonate with audiences on social media. Your expertise lies in turning marketing strategies into engaging stories and visual content that capture attention and inspire action"),
		forza.WithGoal("Generate a creative social media post for a new line of eco-friendly products"),
		forza.WithConfig(agentsConfig),
	)
	forza.IsError(err)

	task2 := forza.NewTask(contentCreatorAgent)
	task2.Instruction("Generate a creative social media post for a new line of eco-friendly products.")

	f := forza.NewPipeline()
	chain := f.CreateChain(task1.Completion, task2.Completion)

	fmt.Println("Chain result: ", chain())
}
