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
	agentsConfig := forza.NewAgentConfig(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

	agentWriter, err := forza.NewAgent(
		forza.WithRole("You are famous writer"),
		forza.WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination."),
		forza.WithGoal("building a compelling narrative"),
		forza.WithConfig(agentsConfig),
	)
	forza.IsError(err)

	task := forza.NewTask(agentWriter)
	task.Instruction("the character is Hercules")

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
	fmt.Println("params: ", params)
	var UserParams UserParams
	err := json.Unmarshal([]byte(params), &UserParams)
	if err != nil {
		panic(err)
	}

	// place any logic here

	return fmt.Sprintf("Answer the exact phrase 'The user id is %s'", UserParams.UserId)
}

func main() {
	agentsConfig := forza.NewAgentConfig(forza.ProviderOpenAi, forza.OpenAIModels.Gpt35turbo)

	agentSpecialist, err := forza.NewAgent(
		forza.WithRole("Specialist"),
		forza.WithBackstory("you are a specialist"),
		forza.WithGoal("you are a specialist"),
		forza.WithConfig(agentsConfig),
	)
	forza.IsError(err)

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
