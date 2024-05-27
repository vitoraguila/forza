package main

import (
	"fmt"
	"os"

	"github.com/vitoraguila/forza"
	s "github.com/vitoraguila/forza/tools/scraper"
)

func getUserId(input string) (string, error) {
	fmt.Println("input: ", input)

	// place any logic here

	return fmt.Sprintf("Answer the exact phrase 'The user id is %s'", input), nil
}

func main() {
	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderOpenAi).
		WithModel(forza.OpenAIModels.Gpt35turbo).
		WithOpenAiCredentials(os.Getenv("OPENAI_API_KEY"))

	agentSpecialist := forza.NewAgent().
		WithRole("Expert to extract content").
		WithBackstory("You will extract two kind of content: a userId and an URL in format http or https. Be super precisely about those formats, userId has never https/http before").
		WithGoal("If the user provide a URL, extract the content from this URL, if the user provide a userId, extract the content from this userId.")

	scraper, _ := s.NewScraper()
	funcCallingParams := forza.NewFunction(
		forza.WithProperty("userId", "user id description", true),
	)

	tasks := agentSpecialist.NewLLMTask(config)
	tasks.WithUserPrompt("what this link is about https://blog.vitoraguila.com/clwaogts30003l808xrbgumu3 ?")
	tasks.AddCustomTools("get_user_id", "user will provide an userId, identify and get this userId", funcCallingParams, getUserId)
	tasks.WithTools(scraper)

	result := tasks.Completion()
	fmt.Println("result TASK: ", result)
}
