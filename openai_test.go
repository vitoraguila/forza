package forza

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func newTestOpenAIAgent(serverURL string) LLMAgent {
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

	// Inject the test server URL by re-creating the OpenAI client config
	o := task.(*openAI)
	o.config.credentials.apiKey = "test-key"

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
	_, err = task.Completion()
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

	_, err := task.Completion("ctx1", "ctx2")
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

	_, err := task.Completion()
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

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithTemperature(0.5).
		WithMaxTokens(100).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("hello")

	// Override the client to use our test server
	o := task.(*openAI)
	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = server.URL + "/v1"
	o.overrideClient = openai.NewClientWithConfig(openaiConfig)

	result, err := task.Completion()
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

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("test prompt")

	o := task.(*openAI)
	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = server.URL + "/v1"
	o.overrideClient = openai.NewClientWithConfig(openaiConfig)

	_, err := task.Completion("previous context")
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

	config := NewLLMConfig().
		WithProvider(ProviderOpenAi).
		WithModel(OpenAIModels.GPT4oMini).
		WithOpenAiCredentials("test-key")

	agent := NewAgent().
		WithRole("Weather Expert").
		WithBackstory("weather forecaster").
		WithGoal("provide weather info")

	task, _ := agent.NewLLMTask(config)
	task.WithUserPrompt("What's the weather in London?")

	// Register a custom tool
	params := NewFunction(WithProperty("city", "city name", true))
	task.AddCustomTools("get_weather", "get weather for a city", params, func(input string) (string, error) {
		return "Sunny, 22C", nil
	})

	o := task.(*openAI)
	openaiConfig := openai.DefaultConfig("test-key")
	openaiConfig.BaseURL = server.URL + "/v1"
	o.overrideClient = openai.NewClientWithConfig(openaiConfig)

	result, err := task.Completion()
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
	o := task.(*openAI)

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
	o := task.(*openAI)

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

	_, err := task.Completion()
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Errorf("expected ErrMissingAPIKey for Azure, got %v", err)
	}
}

// mockTool implements the tools.Tool interface for testing
type mockTool struct {
	name string
	desc string
}

func (m *mockTool) Name() string                          { return m.name }
func (m *mockTool) Description() string                   { return m.desc }
func (m *mockTool) Call(input string) (string, error)     { return "mock result: " + input, nil }
