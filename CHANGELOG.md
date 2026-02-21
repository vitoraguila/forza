# Changelog

## [Unreleased] - v0.2.0

### Breaking Changes
- Exported core types: `Agent`, `LLMConfig`, `LLMAgent`, `Pipeline`, `FunctionShape`, `FunctionProps`
- `NewLLMTask()` now returns `(LLMAgent, error)` instead of panicking
- `Completion()` now returns `(string, error)` instead of just `string`
- `RunConcurrently()` now returns `([]string, error)`
- `CreateChain()` returns a function with signature `func() (string, error)`
- Renamed `WithTempature()` to `WithTemperature()`
- Model field names updated (e.g., `Gpt35turbo` -> `GPT35Turbo`)
- Pipeline type renamed from `forza` to `Pipeline`

### Added
- **Anthropic/Claude support**: `ProviderAnthropic` with Claude 3 Haiku, 3.5 Sonnet, 3.7 Sonnet, 4 Sonnet, 4 Opus
- **Google Gemini support**: `ProviderGemini` with Gemini 2.0 Flash, 2.5 Pro, 2.5 Flash
- **Ollama support**: `ProviderOllama` for local LLMs (Llama 3, Mistral, Mixtral, Phi3, Gemma2, or any custom model)
- Provider registry pattern for extensibility
- `RunSequentially()` method on Pipeline
- `WithMaxTokens()` configuration option
- `WithAnthropicCredentials()`, `WithGeminiCredentials()`, `WithOllamaCredentials()` config methods
- Proper error types in `errors.go` (no more panics in library code)
- Temperature is now passed to the OpenAI API request
- New models: GPT-5, O1, O1-mini for OpenAI
- Comprehensive test suite (87%+ coverage)
- GitHub Actions CI pipeline
- Makefile with common targets
- golangci-lint configuration
- New examples: anthropic-completion, gemini-completion, ollama-completion

### Fixed
- Chain execution bug: tasks were being skipped due to incorrect index logic
- Temperature was configured but never sent to the API
- All examples updated to use `GPT4oMini` instead of deprecated `GPT3.5-turbo`

### Changed
- All panics replaced with proper error returns
- Error handling throughout follows Go conventions
- Tool/function system decoupled from OpenAI-specific types

## [v0.1.0] - Initial Release

- LLM agents with Role, Backstory, Goal
- OpenAI and Azure OpenAI support
- Task pipelines: concurrent execution and chains
- Function calling support
- Web scraper tool
