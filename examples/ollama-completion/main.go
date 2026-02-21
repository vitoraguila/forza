package main

import (
	"fmt"
	"log"

	"github.com/vitoraguila/forza"
)

func main() {
	// Ollama runs locally - no API key needed, just the endpoint.
	// Make sure Ollama is running: ollama serve
	// And you have pulled the model: ollama pull llama3.1
	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderOllama).
		WithModel(forza.OllamaModels.Llama31).
		WithOllamaCredentials("http://localhost:11434/v1")

	agentWriter := forza.NewAgent().
		WithRole("You are famous writer").
		WithBackstory("you know how to captivate your audience with your words. You have a gift for storytelling and creating magical worlds with your imagination. You are known for your enchanting tales that transport readers to far-off lands and spark their imagination.").
		WithGoal("building a compelling narrative")

	task, err := agentWriter.NewLLMTask(config)
	if err != nil {
		log.Fatal(err)
	}
	task.WithUserPrompt("Write a short story about Hercules and the Hydra")

	result, err := task.Completion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("result TASK: ", result)
}
