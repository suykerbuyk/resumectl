package llm

import (
	"errors"
	"fmt"
)

var (
	ErrNoMessages        = errors.New("llm: request must have at least one message")
	ErrNegativeMaxTokens = errors.New("llm: max_tokens must not be negative")
	ErrNoProvider        = errors.New("llm: no provider configured for task")
	ErrProviderNotFound  = errors.New("llm: provider not found")
)

// InvalidRoleError indicates an unsupported message role.
type InvalidRoleError struct {
	Role string
}

func (e *InvalidRoleError) Error() string {
	return fmt.Sprintf("llm: invalid message role %q (must be system, user, or assistant)", e.Role)
}

// APIError represents an error response from an LLM provider API.
type APIError struct {
	StatusCode int
	Provider   string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llm: %s API error (status %d): %s", e.Provider, e.StatusCode, e.Body)
}
