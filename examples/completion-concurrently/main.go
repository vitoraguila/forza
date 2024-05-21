package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	marketAnalystAgent := forza.NewAgent("As the Lead Market Analyst at a premier " +
		"digital marketing firm, you specialize in dissecting " +
		"online business landscapes. Conduct amazing analysis of the products and " +
		"competitors, providing in-depth insights to guide " +
		"marketing strategies.")
	marketAnalystAgent.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task1 := forza.NewTask(marketAnalystAgent)
	task1.SetPrompt("Give me a full report about the market of electric cars in the US.")

	contentCreatorAgent := forza.NewAgent("As a Creative Content Creator at a top-tier " +
		"digital marketing agency, you excel in crafting narratives " +
		"that resonate with audiences on social media. " +
		"Your expertise lies in turning marketing strategies " +
		"into engaging stories and visual content that capture " +
		"attention and inspire action.")
	contentCreatorAgent.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task2 := forza.NewTask(contentCreatorAgent)
	task2.SetPrompt("Generate a creative social media post for a new line of eco-friendly products.")

	// RUNNING ALL CONCURRENTLY
	f := forza.NewPipeline()
	f.AddTasks(task1.Completion, task2.Completion)
	result := f.RunConcurrently()

	fmt.Println(result[0], "result TASK1")
	fmt.Println("-----------------")
	fmt.Println(result[1], "result TASK2")
}
