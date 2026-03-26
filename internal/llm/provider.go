package llm

import "context"

// Provider is the interface that all LLM backends must implement.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Name() string
	DefaultModel() string
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"` // "user" | "assistant" | "system"
	Content string `json:"content"`
}

// CompletionRequest holds the parameters for an LLM completion call.
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Model       string    `json:"model"`
	JSONMode    bool      `json:"json_mode"`
}

// CompletionResponse holds the result of an LLM completion call.
type CompletionResponse struct {
	Content      string `json:"content"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Model        string `json:"model"`
	Provider     string `json:"provider"`
}

// NewSystemMessage creates a system-role message.
func NewSystemMessage(content string) Message {
	return Message{Role: "system", Content: content}
}

// NewUserMessage creates a user-role message.
func NewUserMessage(content string) Message {
	return Message{Role: "user", Content: content}
}

// NewAssistantMessage creates an assistant-role message.
func NewAssistantMessage(content string) Message {
	return Message{Role: "assistant", Content: content}
}

// Validate checks that a CompletionRequest has the minimum required fields.
func (r *CompletionRequest) Validate() error {
	if len(r.Messages) == 0 {
		return ErrNoMessages
	}
	if r.MaxTokens < 0 {
		return ErrNegativeMaxTokens
	}
	for _, m := range r.Messages {
		if m.Role != "system" && m.Role != "user" && m.Role != "assistant" {
			return &InvalidRoleError{Role: m.Role}
		}
	}
	return nil
}
