package forza

import "errors"

var (
	ErrProviderNotFound      = errors.New("provider does not exist")
	ErrModelNotFound         = errors.New("model does not exist for the selected provider")
	ErrMissingRole           = errors.New("agent Role is required (use WithRole())")
	ErrMissingBackstory      = errors.New("agent Backstory is required (use WithBackstory())")
	ErrMissingGoal           = errors.New("agent Goal is required (use WithGoal())")
	ErrMissingPrompt         = errors.New("user prompt is required (use WithUserPrompt())")
	ErrMissingAPIKey         = errors.New("API key not provided")
	ErrMissingEndpoint       = errors.New("endpoint not provided")
	ErrTooManyArgs           = errors.New("too many arguments: only one optional context argument is allowed")
	ErrNilTask               = errors.New("task function is nil")
	ErrCompletionFailed      = errors.New("completion request failed")
	ErrToolCallFailed        = errors.New("tool call execution failed")
	ErrChainInterrupted      = errors.New("chain interrupted by task error")
	ErrMaxToolRoundsExceeded = errors.New("maximum tool call rounds exceeded")
	ErrInvalidConfig         = errors.New("invalid LLM configuration")
	ErrResponseTooLarge      = errors.New("response body exceeds maximum allowed size")
)
