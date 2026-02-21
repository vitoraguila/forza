<a href="https://github.com/vitoraguila/forza">
  <h1 align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://github.com/vitoraguila/forza/blob/master/assets/forza_logo_new.png?raw=true">
      <img alt="Forza" src="https://github.com/vitoraguila/forza/blob/master/assets/forza_logo_new.png?raw=true">
    </picture>
  </h1>
</a>

### Agents framework for Golang

Build AI agents with multiple LLM providers using a unified, idiomatic Go API.

**Supported providers:**

| Provider | Models | Status |
|----------|--------|--------|
| OpenAI | GPT-4o, GPT-4o-mini, GPT-4, GPT-5, O1 | Stable |
| Azure OpenAI | Same as OpenAI | Stable |
| Anthropic | Claude 4 Opus, Claude 4 Sonnet, Claude 3.7/3.5 Sonnet, Claude 3 Haiku | Stable |
| Google Gemini | Gemini 2.5 Pro, Gemini 2.5 Flash, Gemini 2.0 Flash | Stable |
| Ollama (local) | Llama 3, Mistral, Mixtral, Phi3, Gemma2, any custom model | Stable |

**Features:**

- LLM agents with Role, Backstory, and Goal
- Task pipelines: concurrent, sequential, and chained execution
- Function calling / tool use (all providers)
- Built-in web scraper tool
- Proper error handling (no panics)
- 87%+ test coverage

## Installation

```
go get github.com/vitoraguila/forza
```

Requires Go 1.21 or later.

## Environment Variables

### OpenAI
```env
OPENAI_API_KEY=sk-...
```

### Azure OpenAI
```env
AZURE_OPEN_AI_API_KEY=...
AZURE_OPEN_AI_ENDPOINT=https://your-resource.openai.azure.com/
```

### Anthropic
```env
ANTHROPIC_API_KEY=sk-ant-...
```

### Google Gemini
```env
GEMINI_API_KEY=...
```

### Ollama
No API key needed. Just run `ollama serve` locally.

## Quick Start

### OpenAI

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vitoraguila/forza"
)

func main() {
	config := forza.NewLLMConfig().
		WithProvider(forza.ProviderOpenAi).
		WithModel(forza.OpenAIModels.GPT4oMini).
		WithOpenAiCredentials(os.Getenv("OPENAI_API_KEY"))

	agent := forza.NewAgent().
		WithRole("You are famous writer").
		WithBackstory("you know how to captivate your audience with your words").
		WithGoal("building a compelling narrative")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		log.Fatal(err)
	}
	task.WithUserPrompt("Write a story about Hercules and the Hydra")

	result, err := task.Completion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
```

### Anthropic (Claude)

```go
config := forza.NewLLMConfig().
	WithProvider(forza.ProviderAnthropic).
	WithModel(forza.AnthropicModels.Claude4Sonnet).
	WithAnthropicCredentials(os.Getenv("ANTHROPIC_API_KEY"))
```

### Google Gemini

```go
config := forza.NewLLMConfig().
	WithProvider(forza.ProviderGemini).
	WithModel(forza.GeminiModels.Gemini25Flash).
	WithGeminiCredentials(os.Getenv("GEMINI_API_KEY"))
```

### Ollama (Local LLMs)

```go
config := forza.NewLLMConfig().
	WithProvider(forza.ProviderOllama).
	WithModel(forza.OllamaModels.Llama31).
	WithOllamaCredentials("http://localhost:11434/v1")
```

## Usage

### Running tasks concurrently

```go
pipeline := forza.NewPipeline()
pipeline.AddTasks(task1.Completion, task2.Completion)

results, err := pipeline.RunConcurrently()
if err != nil {
	log.Fatal(err)
}
fmt.Println("Task 1:", results[0])
fmt.Println("Task 2:", results[1])
```

### Chaining tasks

Each task receives the previous task's output as context:

```go
pipeline := forza.NewPipeline()
chain := pipeline.CreateChain(researchTask.Completion, writerTask.Completion)

result, err := chain()
if err != nil {
	log.Fatal(err)
}
fmt.Println(result)
```

### Running tasks sequentially

```go
pipeline := forza.NewPipeline()
pipeline.AddTasks(task1.Completion, task2.Completion, task3.Completion)

results, err := pipeline.RunSequentially()
if err != nil {
	log.Fatal(err)
}
```

### Function calling / Tool use

```go
// Built-in scraper tool
scraper, _ := scraper.NewScraper()
task.WithTools(scraper)

// Custom tools
params := forza.NewFunction(
	forza.WithProperty("city", "city name", true),
)
task.AddCustomTools("get_weather", "get weather for a city", params, func(input string) (string, error) {
	// your logic here
	return "Sunny, 22C", nil
})
```

### Configuration options

```go
config := forza.NewLLMConfig().
	WithProvider(forza.ProviderOpenAi).
	WithModel(forza.OpenAIModels.GPT4oMini).
	WithTemperature(0.7).      // 0.0 - 2.0 (default: 0.3)
	WithMaxTokens(2048).       // max response tokens (default: 4096)
	WithOpenAiCredentials(key)
```

## Available Models

### OpenAI
| Constant | Model |
|----------|-------|
| `OpenAIModels.GPT4oMini` | gpt-4o-mini (recommended) |
| `OpenAIModels.GPT4o` | gpt-4o |
| `OpenAIModels.GPT4` | gpt-4 |
| `OpenAIModels.GPT4Turbo` | gpt-4-turbo |
| `OpenAIModels.GPT5` | gpt-5 |
| `OpenAIModels.O1` | o1 |
| `OpenAIModels.O1Mini` | o1-mini |
| `OpenAIModels.GPT35Turbo` | gpt-3.5-turbo (deprecated) |

### Anthropic
| Constant | Model |
|----------|-------|
| `AnthropicModels.Claude4Opus` | claude-opus-4-20250514 |
| `AnthropicModels.Claude4Sonnet` | claude-sonnet-4-20250514 |
| `AnthropicModels.Claude37Sonnet` | claude-3-7-sonnet-latest |
| `AnthropicModels.Claude35Sonnet` | claude-3-5-sonnet-latest |
| `AnthropicModels.Claude3Haiku` | claude-3-haiku-20240307 |

### Google Gemini
| Constant | Model |
|----------|-------|
| `GeminiModels.Gemini25Pro` | gemini-2.5-pro |
| `GeminiModels.Gemini25Flash` | gemini-2.5-flash |
| `GeminiModels.Gemini20Flash` | gemini-2.0-flash |

### Ollama
| Constant | Model |
|----------|-------|
| `OllamaModels.Llama31` | llama3.1 |
| `OllamaModels.Llama3` | llama3 |
| `OllamaModels.Mistral` | mistral |
| `OllamaModels.Mixtral` | mixtral |
| `OllamaModels.Phi3` | phi3 |
| `OllamaModels.Gemma2` | gemma2 |

Ollama also accepts any custom model string.

## Architecture

```
forza/
├── agent.go        # Agent + provider factory
├── llm.go          # LLMAgent interface + LLMConfig
├── common.go       # Provider constants + model registry
├── errors.go       # Error types
├── functions.go    # Function calling parameter builder
├── forza.go        # Pipeline: concurrent, sequential, chain
├── openai.go       # OpenAI / Azure provider
├── anthropic.go    # Anthropic (Claude) provider
├── gemini.go       # Google Gemini provider
├── ollama.go       # Ollama (local LLMs) provider
├── tools/
│   ├── tool.go     # Tool interface
│   └── scraper/    # Web scraper tool
└── examples/       # Usage examples per provider
```

## Development

```bash
make test       # Run tests
make cover      # Run tests with coverage
make lint       # Run golangci-lint
make build      # Build all packages
make check      # Run vet + lint + test
```

## Contributing

Contributions, suggestions, and feature requests are welcome.

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure `make check` passes
5. Submit a pull request

## License

[MIT](LICENSE)
