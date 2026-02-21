package forza

import (
	"errors"
	"testing"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent()
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.Role != "" || agent.Backstory != "" || agent.Goal != "" {
		t.Error("expected empty fields on new agent")
	}
}

func TestAgent_BuilderChain(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("A great storyteller").
		WithGoal("Write stories")

	if agent.Role != "Writer" {
		t.Errorf("expected role 'Writer', got %q", agent.Role)
	}
	if agent.Backstory != "A great storyteller" {
		t.Errorf("expected backstory, got %q", agent.Backstory)
	}
	if agent.Goal != "Write stories" {
		t.Errorf("expected goal, got %q", agent.Goal)
	}
}

func TestAgent_NewLLMTask_MissingRole(t *testing.T) {
	agent := NewAgent().
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("key")

	_, err := agent.NewLLMTask(config)
	if err == nil {
		t.Fatal("expected error for missing role")
	}
	if !errors.Is(err, ErrMissingRole) {
		t.Errorf("expected ErrMissingRole, got %v", err)
	}
}

func TestAgent_NewLLMTask_MissingBackstory(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("key")

	_, err := agent.NewLLMTask(config)
	if err == nil {
		t.Fatal("expected error for missing backstory")
	}
	if !errors.Is(err, ErrMissingBackstory) {
		t.Errorf("expected ErrMissingBackstory, got %v", err)
	}
}

func TestAgent_NewLLMTask_MissingGoal(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory")

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("key")

	_, err := agent.NewLLMTask(config)
	if err == nil {
		t.Fatal("expected error for missing goal")
	}
	if !errors.Is(err, ErrMissingGoal) {
		t.Errorf("expected ErrMissingGoal, got %v", err)
	}
}

func TestAgent_NewLLMTask_InvalidProvider(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider("nonexistent").
		WithModel("some-model").
		WithOpenAiCredentials("key")

	_, err := agent.NewLLMTask(config)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
}

func TestAgent_NewLLMTask_InvalidModel(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel("nonexistent-model").
		WithOpenAiCredentials("key")

	_, err := agent.NewLLMTask(config)
	if err == nil {
		t.Fatal("expected error for invalid model")
	}
	if !errors.Is(err, ErrModelNotFound) {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}

func TestAgent_NewLLMTask_Success_OpenAI(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("key")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
}

func TestAgent_NewLLMTask_Success_Anthropic(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials("key")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
}

func TestAgent_NewLLMTask_Success_Gemini(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash).
		WithGeminiCredentials("key")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
}

func TestAgent_NewLLMTask_Success_Ollama(t *testing.T) {
	agent := NewAgent().
		WithRole("Writer").
		WithBackstory("backstory").
		WithGoal("goal")

	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel(OllamaModels.Llama31).
		WithOllamaCredentials("http://localhost:11434/v1")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
}

func TestProviderFactory_AllProvidersRegistered(t *testing.T) {
	providers := []string{
		ProviderOpenAi,
		ProviderAzure,
		ProviderAnthropic,
		ProviderGemini,
		ProviderOllama,
	}

	for _, p := range providers {
		if _, ok := providerFactory[p]; !ok {
			t.Errorf("provider %q not registered in providerFactory", p)
		}
	}
}
