# Forza

### Agents framework for Golang

This project is in a early stage, so contributions, suggestions and feature request are pretty welcome.

We support: 

* LLM agents
* Tasks
* Run tasks concurrently
* OpenAI
* Azure OpenAI
* Function calling

## Installation

```
go get github.com/vitoraguila/forza
```
Currently, `forza` requires Go version 1.18 or greater.


## Usage

### Agents using completion:

```go
package main

import (
	"fmt"

	"github.com/vitoraguila/forza"
)

func main() {
	agentStoryTeller := forza.NewAgent(&forza.AgentPersona{
		Role:      "storyteller",
		Backstory: "you are a storyteller, write as a fairy tale for kids",
		Goal:      "write a fairy tale",
	})
	agentStoryTeller.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)
	task := forza.NewTask(agentStoryTeller)
	task.Instruction("who is Hercules?")

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
	"encoding/json"
	"fmt"

	"github.com/vitoraguila/forza"
)

type UserParams struct {
	UserId string `json:"userId" required:"true"`
}

func getUserId(params string) string {
	var UserParams UserParams
	err := json.Unmarshal([]byte(params), &UserParams)
	if err != nil {
		panic(err)
	}

	// place any logic here

	return fmt.Sprintf("The user id is %s", UserParams.UserId)
}

func main() {
	agentSpecialist := forza.NewAgent(&forza.AgentPersona{
		Role:      "specialist",
		Backstory: "You are a specialist to identify userId in the text",
		Goal:      "identify userId",
	})
	agentSpecialist.Configure(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

	funcCallingParams := forza.FunctionShape{
		"userId": forza.FunctionProps{
			Description: "user id description",
			Required:    true,
		},
	}

	task := forza.NewTask(agentSpecialist)
	task.Instruction("My name is robert and my user id is 3434")
	task.SetFunction("get_user_id", "user will provide an userId, identify and get this userId", funcCallingParams, getUserId)

	result := task.Completion()
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

```
</details>

### TODO:

- [ ] Add tests
- [ ] Add support for Gemini
- [ ] Add support for Llama
- [x] Implement chain of actions