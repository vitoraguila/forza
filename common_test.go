package forza

import (
	"testing"
)

func TestCheckModel_ValidOpenAI(t *testing.T) {
	tests := []struct {
		provider string
		model    string
	}{
		{ProviderOpenAi, OpenAIModels.GPT4oMini},
		{ProviderOpenAi, OpenAIModels.GPT4o},
		{ProviderOpenAi, OpenAIModels.GPT4},
		{ProviderOpenAi, OpenAIModels.GPT4Turbo},
		{ProviderOpenAi, OpenAIModels.GPT35Turbo},
		{ProviderAzure, OpenAIModels.GPT4oMini},
	}

	for _, tt := range tests {
		ok, msg := checkModel(tt.provider, tt.model)
		if !ok {
			t.Errorf("checkModel(%q, %q) = false: %s", tt.provider, tt.model, msg)
		}
	}
}

func TestCheckModel_ValidAnthropic(t *testing.T) {
	tests := []struct {
		model string
	}{
		{AnthropicModels.Claude35Sonnet},
		{AnthropicModels.Claude37Sonnet},
		{AnthropicModels.Claude4Sonnet},
		{AnthropicModels.Claude4Opus},
		{AnthropicModels.Claude3Haiku},
	}

	for _, tt := range tests {
		ok, msg := checkModel(ProviderAnthropic, tt.model)
		if !ok {
			t.Errorf("checkModel(anthropic, %q) = false: %s", tt.model, msg)
		}
	}
}

func TestCheckModel_ValidGemini(t *testing.T) {
	tests := []struct {
		model string
	}{
		{GeminiModels.Gemini20Flash},
		{GeminiModels.Gemini25Pro},
		{GeminiModels.Gemini25Flash},
	}

	for _, tt := range tests {
		ok, msg := checkModel(ProviderGemini, tt.model)
		if !ok {
			t.Errorf("checkModel(gemini, %q) = false: %s", tt.model, msg)
		}
	}
}

func TestCheckModel_OllamaAcceptsAnyModel(t *testing.T) {
	// Ollama should accept any model string since users can pull custom models
	tests := []string{
		"llama3",
		"my-custom-model",
		"totally-made-up",
	}

	for _, model := range tests {
		ok, _ := checkModel(ProviderOllama, model)
		if !ok {
			t.Errorf("checkModel(ollama, %q) should accept any model", model)
		}
	}
}

func TestCheckModel_InvalidProvider(t *testing.T) {
	ok, msg := checkModel("nonexistent", "gpt-4")
	if ok {
		t.Error("expected checkModel to return false for invalid provider")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCheckModel_InvalidModel(t *testing.T) {
	ok, msg := checkModel(ProviderOpenAi, "nonexistent-model")
	if ok {
		t.Error("expected checkModel to return false for invalid model")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestOpenAIModels_ListModels(t *testing.T) {
	models := OpenAIModels.ListModels()
	if len(models) == 0 {
		t.Fatal("expected non-empty model list")
	}

	// Check that all known models are in the list
	expected := map[string]bool{
		"gpt-3.5-turbo": true,
		"gpt-4":         true,
		"gpt-4o":        true,
		"gpt-4-turbo":   true,
		"gpt-4o-mini":   true,
	}

	for _, m := range models {
		delete(expected, m)
	}

	for m := range expected {
		t.Errorf("expected model %q in list", m)
	}
}

func TestAnthropicModels_ListModels(t *testing.T) {
	models := AnthropicModels.ListModels()
	if len(models) == 0 {
		t.Fatal("expected non-empty model list")
	}
}

func TestGeminiModels_ListModels(t *testing.T) {
	models := GeminiModels.ListModels()
	if len(models) == 0 {
		t.Fatal("expected non-empty model list")
	}
}

func TestOllamaModels_ListModels(t *testing.T) {
	models := OllamaModels.ListModels()
	if len(models) == 0 {
		t.Fatal("expected non-empty model list")
	}
}

func TestProviderConstants(t *testing.T) {
	providers := []string{
		ProviderOpenAi,
		ProviderAzure,
		ProviderAnthropic,
		ProviderGemini,
		ProviderOllama,
	}

	seen := make(map[string]bool)
	for _, p := range providers {
		if p == "" {
			t.Error("provider constant should not be empty")
		}
		if seen[p] {
			t.Errorf("duplicate provider constant: %q", p)
		}
		seen[p] = true
	}
}
