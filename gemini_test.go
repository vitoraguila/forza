package forza

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestGeminiTask(serverURL string) LLMAgent {
	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash).
		WithGeminiCredentials("test-key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)

	// Override HTTP client to point to test server
	g := task.(*geminiProvider)
	g.httpClient = &http.Client{
		Transport: &geminiRewriteTransport{baseURL: serverURL},
	}

	return task
}

type geminiRewriteTransport struct {
	baseURL string
}

func (t *geminiRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the Gemini API URL to our test server
	req.URL.Scheme = "http"
	host := strings.TrimPrefix(t.baseURL, "http://")
	req.URL.Host = host
	return http.DefaultTransport.RoundTrip(req)
}

func TestGemini_Completion_MissingPrompt(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash).
		WithGeminiCredentials("key")

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

func TestGemini_Completion_MissingAPIKey(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash)

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

func TestGemini_Completion_TooManyArgs(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash).
		WithGeminiCredentials("key")

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

func TestGemini_Completion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content: geminiContent{
						Role: "model",
						Parts: []geminiPart{
							{Text: "Hello from Gemini mock!"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestGeminiTask(server.URL)
	task.WithUserPrompt("hello")

	result, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from Gemini mock!" {
		t.Errorf("expected mock response, got %q", result)
	}
}

func TestGemini_Completion_WithContext(t *testing.T) {
	var capturedBody geminiRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)

		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{Content: geminiContent{Parts: []geminiPart{{Text: "response"}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestGeminiTask(server.URL)
	task.WithUserPrompt("test prompt")

	_, err := task.Completion("previous result")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedBody.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(capturedBody.Contents))
	}
	if len(capturedBody.Contents[0].Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(capturedBody.Contents[0].Parts))
	}

	text := capturedBody.Contents[0].Parts[0].Text
	if !strings.Contains(text, "previous result") {
		t.Errorf("expected context in prompt, got %q", text)
	}
}

func TestGemini_Completion_SystemInstruction(t *testing.T) {
	var capturedBody geminiRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)

		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{Content: geminiContent{Parts: []geminiPart{{Text: "response"}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestGeminiTask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedBody.SystemInstruction == nil {
		t.Fatal("expected system instruction to be set")
	}
	if len(capturedBody.SystemInstruction.Parts) == 0 {
		t.Fatal("expected system instruction parts")
	}
}

func TestGemini_Completion_ToolCalling(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp geminiResponse

		if callCount == 1 {
			resp = geminiResponse{
				Candidates: []geminiCandidate{
					{
						Content: geminiContent{
							Role: "model",
							Parts: []geminiPart{
								{
									FunctionCall: &geminiFunctionCall{
										Name: "get_weather",
										Args: map[string]interface{}{"city": "Tokyo"},
									},
								},
							},
						},
					},
				},
			}
		} else {
			resp = geminiResponse{
				Candidates: []geminiCandidate{
					{Content: geminiContent{Parts: []geminiPart{{Text: "Weather in Tokyo is sunny."}}}},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestGeminiTask(server.URL)
	task.WithUserPrompt("What's the weather in Tokyo?")

	params := NewFunction(WithProperty("city", "city name", true))
	task.AddCustomTools("get_weather", "get weather", params, func(input string) (string, error) {
		return "Sunny, 25C", nil
	})

	result, err := task.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Weather in Tokyo is sunny." {
		t.Errorf("expected tool follow-up response, got %q", result)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestGemini_Completion_NoCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{Candidates: []geminiCandidate{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestGeminiTask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion()
	if err == nil {
		t.Fatal("expected error for no candidates")
	}
	if !errors.Is(err, ErrCompletionFailed) {
		t.Errorf("expected ErrCompletionFailed, got %v", err)
	}
}

func TestGemini_Completion_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := geminiResponse{
			Error: &geminiError{
				Code:    400,
				Message: "Invalid API key",
				Status:  "INVALID_ARGUMENT",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	task := newTestGeminiTask(server.URL)
	task.WithUserPrompt("hello")

	_, err := task.Completion()
	if err == nil {
		t.Fatal("expected error for API error")
	}
	if !errors.Is(err, ErrCompletionFailed) {
		t.Errorf("expected ErrCompletionFailed, got %v", err)
	}
}

func TestGemini_WithTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash).
		WithGeminiCredentials("key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	g := task.(*geminiProvider)

	tool := &mockTool{name: "scraper", desc: "scrapes web"}
	task.WithTools(tool)

	if len(g.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(g.functions))
	}
	if g.functions[0].Name != "scraper" {
		t.Errorf("expected name 'scraper', got %q", g.functions[0].Name)
	}
}

func TestGemini_AddCustomTools(t *testing.T) {
	config := NewLLMConfig().
		WithProvider(ProviderGemini).
		WithModel(GeminiModels.Gemini25Flash).
		WithGeminiCredentials("key")

	agent := NewAgent().
		WithRole("Tester").
		WithBackstory("backstory").
		WithGoal("goal")

	task, _ := agent.NewLLMTask(config)
	g := task.(*geminiProvider)

	params := NewFunction(WithProperty("query", "search query", true))
	task.AddCustomTools("search", "search web", params, func(input string) (string, error) {
		return "results", nil
	})

	if len(g.functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(g.functions))
	}
	if g.functions[0].Name != "search" {
		t.Errorf("expected name 'search', got %q", g.functions[0].Name)
	}
}
