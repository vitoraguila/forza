package forza

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func newTestOllamaTask(serverURL string) LLMAgent {
	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel(OllamaModels.Llama31).
		WithOllamaCredentials(serverURL + "/v1")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	return task
}

func TestOllama_Completion_MissingPrompt(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel(OllamaModels.Llama31).
		WithOllamaCredentials("http://localhost:11434/v1")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	_, err := task.Completion()
	if !errors.Is(err, ErrMissingPrompt) {
		t.Errorf("expected ErrMissingPrompt, got %v", err)
	}
}

func TestOllama_Completion_TooManyArgs(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel(OllamaModels.Llama31).
		WithOllamaCredentials("http://localhost:11434/v1")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("hello")

	_, err := task.Completion("a", "b")
	if !errors.Is(err, ErrTooManyArgs) {
		t.Errorf("expected ErrTooManyArgs, got %v", err)
	}
}

func TestOllama_Completion_DefaultEndpoint(t *testing.T) {
	// When no endpoint is set, should use default
	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel(OllamaModels.Llama31)

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	o := task.(*ollamaProvider)

	client, err := o.createClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestOllama_Completion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "Hello from Ollama mock!",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOllamaTask(server.URL)
	task.WithUserPrompt("hello")

	result, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from Ollama mock!" {
		t.Errorf("expected mock response, got %q", result)
	}
}

func TestOllama_Completion_ToolCalling(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp openai.ChatCompletionResponse

		if callCount == 1 {
			resp = openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							ToolCalls: []openai.ToolCall{
								{
									ID:   "call_1",
									Type: openai.ToolTypeFunction,
									Function: openai.FunctionCall{
										Name:      "calc",
										Arguments: `{"expr": "2+2"}`,
									},
								},
							},
						},
					},
				},
			}
		} else {
			resp = openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "The answer is 4.",
						},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOllamaTask(server.URL)
	task.WithUserPrompt("What is 2+2?")

	params := NewFunction(WithProperty("expr", "math expression", true))
	task.AddCustomTools("calc", "calculate", params, func(input string) (string, error) {
		return "4", nil
	})

	result, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "The answer is 4." {
		t.Errorf("expected tool follow-up, got %q", result)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestOllama_WithTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel(OllamaModels.Llama31).
		WithOllamaCredentials("http://localhost:11434/v1")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	o := task.(*ollamaProvider)

	tool := &mockTool{name: "tool1", desc: "desc1"}
	task.WithTools(tool)

	if len(o.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(o.functions))
	}
}

func TestOllama_AcceptsCustomModel(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOllama).
		WithModel("my-custom-finetuned-model").
		WithOllamaCredentials("http://localhost:11434/v1")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		t.Fatalf("ollama should accept any model name, got error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
}
