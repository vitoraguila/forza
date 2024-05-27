<a href="https://flutter.dev/">
  <h1 align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://github.com/vitoraguila/forza/blob/master/assets/forza_logo_new.png?raw=true">
      <img alt="Forza" src="https://github.com/vitoraguila/forza/blob/master/assets/forza_logo_new.png?raw=true">
    </picture>
  </h1>
</a>

### Agents framework for Golang

This project is in a early stage, so contributions, suggestions and feature request are pretty welcome.

We support: 

* LLM agents
* Tasks
* Run tasks concurrently
* Run tasks in chain
* OpenAI
* Azure OpenAI
* Function calling

## Installation

```
go get github.com/vitoraguila/forza
```
Currently, `forza` requires Go version 1.18 or greater.

## Environment variables

### OpenAI
```.env
OPENAI_API_KEY=
```

### Azure OpenAI
```.env
AZURE_OPEN_AI_API_KEY=
AZURE_OPEN_AI_ENDPOINT=
```

## Usage

### Agents using completion:

```go
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


```

### Getting an OpenAI API Key:

1. Visit the OpenAI website at [https://platform.openai.com/account/api-keys](https://platform.openai.com/account/api-keys).
2. If you don't have an account, click on "Sign Up" to create one. If you do, click "Log In".
3. Once logged in, navigate to your API key management page.
4. Click on "Create new secret key".
5. Enter a name for your new key, then click "Create secret key".
6. Your new API key will be displayed. Use this key to interact with the OpenAI API.

**Note:** Your API key is sensitive information. Do not share it with anyone.

### Other examples:

<details>
<summary>Agents running completion concurrently</summary>

```go
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

	marketAnalystAgent := forza.NewAgent().
		WithRole("Lead Market Analyst at a premier digital marketing firm").
		WithBackstory("you specialize in dissecting online business landscapes. Conduct amazing analysis of the products and competitors").
		WithGoal("providing in-depth insights to guide marketing strategies")

	task1 := marketAnalystAgent.NewLLMTask(config)
	task1.WithUserPrompt("Give me a full report about the market of electric cars in the US.")

	contentCreatorAgent := forza.NewAgent().
		WithRole("Creative Content Creator at a top-tier digital marketing agency").
		WithBackstory("you excel in crafting narratives that resonate with audiences on social media. Your expertise lies in turning marketing strategies into engaging stories and visual content that capture attention and inspire action").
		WithGoal("Generate a creative social media post for a new line of eco-friendly products")

	task2 := contentCreatorAgent.NewLLMTask(config)
	task2.WithUserPrompt("Generate a creative social media post for a new line of eco-friendly products.")

	// RUNNING ALL CONCURRENTLY
	f := forza.NewPipeline()

	f.AddTasks(task1.Completion, task2.Completion)
	result := f.RunConcurrently()

	fmt.Println("result TASK1: ", result[0])
	fmt.Println("-----------------")
	fmt.Println("result TASK2: ", result[1])
}

```
</details>

<details>
<summary>Agents using function calling</summary>

```go
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

```
</details>

<details>
<summary>Agents chains for tasks</summary>

```go
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

```
</details>

### TODO:

- [ ] Add tests
- [ ] Add support for Gemini
- [ ] Add support for Llama
- [x] Implement chain of actions
