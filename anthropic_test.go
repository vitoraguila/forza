package forza

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestAnthropicTask(serverURL string) LLMAgent {
	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)

	// Override the HTTP client to point to test server
	a := task.(*anthropicProvider)
	a.httpClient = &http.Client{
		Transport: &rewriteTransport{baseURL: serverURL},
	}

	return task
}

// rewriteTransport redirects all requests to a test server
type rewriteTransport struct {
	baseURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[len("http://"):]
	return http.DefaultTransport.RoundTrip(req)
}

func TestAnthropic_Completion_MissingPrompt(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials("key")

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

func TestAnthropic_Completion_MissingAPIKey(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet)

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

func TestAnthropic_Completion_TooManyArgs(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials("key")

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

func TestAnthropic_Completion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key header, got %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != anthropicAPIVersion {
			t.Errorf("expected anthropic-version header")
		}

		resp := anthropicResponse{
			ID:   "msg_test",
			Type: "message",
			Role: "assistant",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Hello from Claude mock!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestAnthropicTask(server.URL)
	task.WithUserPrompt("hello")

	result, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from Claude mock!" {
		t.Errorf("expected mock response, got %q", result)
	}
}

func TestAnthropic_Completion_WithContext(t *testing.T) {
	var capturedBody anthropicRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)

		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: "response"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestAnthropicTask(server.URL)
	task.WithUserPrompt("test prompt")

	_, err := task.Completion("previous result")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The user message should contain both the prompt and context
	if len(capturedBody.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(capturedBody.Messages))
	}

	// Message content is a string in this case
	content, ok := capturedBody.Messages[0].Content.(string)
	if !ok {
		t.Fatal("expected string content")
	}
	expected := "test prompt\n\nTake in consideration the following context: previous result"
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}

func TestAnthropic_Completion_SystemPrompt(t *testing.T) {
	var capturedBody anthropicRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)

		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: "response"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestAnthropicTask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify system prompt is set
	if capturedBody.System == "" {
		t.Error("expected system prompt to be set")
	}
}

func TestAnthropic_Completion_ToolCalling(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp anthropicResponse

		if callCount == 1 {
			resp = anthropicResponse{
				Content: []anthropicContentBlock{
					{Type: "text", Text: "I'll check the weather."},
					{
						Type:  "tool_use",
						ID:    "toolu_1",
						Name:  "get_weather",
						Input: json.RawMessage(`{"city": "London"}`),
					},
				},
			}
		} else {
			resp = anthropicResponse{
				Content: []anthropicContentBlock{
					{Type: "text", Text: "The weather in London is sunny."},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestAnthropicTask(server.URL)
	task.WithUserPrompt("What's the weather?")

	params := NewFunction(WithProperty("city", "city name", true))
	task.AddCustomTools("get_weather", "get weather", params, func(input string) (string, error) {
		return "Sunny, 22C", nil
	})

	result, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "The weather in London is sunny." {
		t.Errorf("expected tool follow-up response, got %q", result)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestAnthropic_Completion_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Error: &anthropicError{
				Type:    "invalid_request_error",
				Message: "Invalid API key",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestAnthropicTask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion()
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !errors.Is(err, ErrCompletionFailed) {
		t.Errorf("expected ErrCompletionFailed, got %v", err)
	}
}

func TestAnthropic_WithTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials("key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	a := task.(*anthropicProvider)

	tool := &mockTool{name: "scraper", desc: "scrapes web pages"}
	task.WithTools(tool)

	if len(a.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(a.functions))
	}
	if a.functions[0].Name != "scraper" {
		t.Errorf("expected function name 'scraper', got %q", a.functions[0].Name)
	}
	if !a.builtinTools["scraper"] {
		t.Error("expected scraper to be marked as builtin")
	}
}

func TestAnthropic_AddCustomTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderAnthropic).
		WithModel(AnthropicModels.Claude4Sonnet).
		WithAnthropicCredentials("key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	a := task.(*anthropicProvider)

	params := NewFunction(
		WithProperty("query", "search query", true),
		WithProperty("limit", "result limit", false),
	)
	task.AddCustomTools("search", "search the web", params, func(input string) (string, error) {
		return "results", nil
	})

	if len(a.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(a.functions))
	}
	if a.functions[0].Name != "search" {
		t.Errorf("expected function name 'search', got %q", a.functions[0].Name)
	}

	schema := a.functions[0].InputSchema
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties in schema")
	}
	if _, ok := props["query"]; !ok {
		t.Error("expected 'query' in properties")
	}
	if _, ok := props["limit"]; !ok {
		t.Error("expected 'limit' in properties")
	}
}
