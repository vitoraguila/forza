package forza

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// mockTool implements the tools.Tool interface for testing
type mockTool struct {
	name string
	desc string
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.desc }
func (m *mockTool) Call(ctx context.Context, input string) (string, error) {
	return "mock result: " + input, nil
}

func newTestOpenAITask(serverURL string) LLMAgent {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithTemperature(0.5).
		WithMaxTokens(100).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("a test agent").
		WithGoal("testing")

	task, _ := agent.NewLLMTask(config)

	// Inject test client
	o := task.(*openaiProvider)
	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = serverURL + "/v1"
	o.client = openai.NewClientWithConfig(openaiConfig)

	return task
}

func TestOpenAI_Completion_MissingPrompt(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, err := agent.NewLLMTask(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Don't set user prompt
	_, err = task.Completion(context.Background())
	if !errors.Is(err, ErrMissingPrompt) {
		t.Errorf("expected ErrMissingPrompt, got %v", err)
	}
}

func TestOpenAI_Completion_TooManyArgs(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("hello")

	_, err := task.Completion(context.Background(), "ctx1", "ctx2")
	if !errors.Is(err, ErrTooManyArgs) {
		t.Errorf("expected ErrTooManyArgs, got %v", err)
	}
}

func TestOpenAI_Completion_MissingAPIKey(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini)
	// No credentials set

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("hello")

	_, err := task.Completion(context.Background())
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Errorf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestOpenAI_Completion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "Hello from OpenAI mock!",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOpenAITask(server.URL)
	task.WithUserPrompt("hello")

	result, err := task.Completion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from OpenAI mock!" {
		t.Errorf("expected mock response, got %q", result)
	}
}

func TestOpenAI_Completion_WithContext(t *testing.T) {
	var capturedMessages []openai.ChatCompletionMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openai.ChatCompletionRequest
		json.NewDecoder(r.Body).Decode(&req)
		capturedMessages = req.Messages

		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{Content: "response"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOpenAITask(server.URL)
	task.WithUserPrompt("test prompt")

	_, err := task.Completion(context.Background(), "previous context")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the user message includes the context
	lastMsg := capturedMessages[len(capturedMessages)-1]
	if lastMsg.Role != "user" {
		t.Errorf("expected last message role 'user', got %q", lastMsg.Role)
	}
	if lastMsg.Content != "test prompt\n\nTake in consideration the following context: previous context" {
		t.Errorf("expected context in user prompt, got %q", lastMsg.Content)
	}
}

func TestOpenAI_Completion_ToolCalling(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp openai.ChatCompletionResponse

		if callCount == 1 {
			// First call: return a tool call
			resp = openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							ToolCalls: []openai.ToolCall{
								{
									ID:   "call_1",
									Type: openai.ToolTypeFunction,
									Function: openai.FunctionCall{
										Name:      "get_weather",
										Arguments: `{"city": "London"}`,
									},
								},
							},
						},
					},
				},
			}
		} else {
			// Second call: return final response after tool result
			resp = openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "The weather in London is sunny.",
						},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOpenAITask(server.URL)
	task.WithUserPrompt("What's the weather in London?")

	// Register a custom tool
	params := NewFunction(WithProperty("city", "city name", true))
	task.AddCustomTools("get_weather", "get weather for a city", params, func(input string) (string, error) {
		return "Sunny, 22C", nil
	})

	result, err := task.Completion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "The weather in London is sunny." {
		t.Errorf("expected tool-call follow-up response, got %q", result)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (tool call + follow-up), got %d", callCount)
	}
}

func TestOpenAI_WithTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	o := task.(*openaiProvider)

	// Create a mock tool
	tool := &mockTool{name: "test_tool", desc: "a test tool"}
	task.WithTools(tool)

	if len(o.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(o.functions))
	}
	if o.functions[0].Name != "test_tool" {
		t.Errorf("expected function name 'test_tool', got %q", o.functions[0].Name)
	}
	if !o.builtinTools["test_tool"] {
		t.Error("expected test_tool to be marked as builtin")
	}
}

func TestOpenAI_AddCustomTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	o := task.(*openaiProvider)

	params := NewFunction(WithProperty("query", "search query", true))
	task.AddCustomTools("search", "search the web", params, func(input string) (string, error) {
		return "results", nil
	})

	if len(o.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(o.functions))
	}
	if o.functions[0].Name != "search" {
		t.Errorf("expected function name 'search', got %q", o.functions[0].Name)
	}
	if _, exists := o.fnExecutable["search"]; !exists {
		t.Error("expected search function to be registered")
	}
}

func TestOpenAI_Azure_MissingCredentials(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderAzure).
		WithModel(OpenAIModels.GPT4oMini)
	// No credentials

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("hello")

	_, err := task.Completion(context.Background())
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Errorf("expected ErrMissingAPIKey for Azure, got %v", err)
	}
}

func TestOpenAI_Completion_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOpenAITask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion(context.Background())
	if err == nil {
		t.Fatal("expected error for empty response")
	}
	if !errors.Is(err, ErrCompletionFailed) {
		t.Errorf("expected ErrCompletionFailed, got %v", err)
	}
}

func TestOpenAI_Completion_UnknownToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						ToolCalls: []openai.ToolCall{
							{
								ID:   "call_1",
								Type: openai.ToolTypeFunction,
								Function: openai.FunctionCall{
									Name:      "nonexistent_tool",
									Arguments: `{}`,
								},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOpenAITask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion(context.Background())
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !errors.Is(err, ErrToolCallFailed) {
		t.Errorf("expected ErrToolCallFailed, got %v", err)
	}
}

func TestOpenAI_Completion_ToolReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						ToolCalls: []openai.ToolCall{
							{
								ID:   "call_1",
								Type: openai.ToolTypeFunction,
								Function: openai.FunctionCall{
									Name:      "failing_tool",
									Arguments: `{"query": "test"}`,
								},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestOpenAITask(server.URL)
	task.WithUserPrompt("hello")

	params := NewFunction(WithProperty("query", "q", true))
	task.AddCustomTools("failing_tool", "a tool that fails", params, func(input string) (string, error) {
		return "", errors.New("tool broke")
	})

	_, err := task.Completion(context.Background())
	if err == nil {
		t.Fatal("expected error from tool")
	}
	if !errors.Is(err, ErrToolCallFailed) {
		t.Errorf("expected ErrToolCallFailed, got %v", err)
	}
}
