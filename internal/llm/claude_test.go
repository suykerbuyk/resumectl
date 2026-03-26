package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaude_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, anthropicVersion, r.Header.Get("anthropic-version"))

		body, _ := io.ReadAll(r.Body)
		var req claudeRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "claude-sonnet-4-6", req.Model)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)

		resp := claudeResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "Hello from Claude"}},
			Model: "claude-sonnet-4-6",
		}
		resp.Usage.InputTokens = 10
		resp.Usage.OutputTokens = 5

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewClaudeProvider("test-key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages:  []Message{NewUserMessage("Hello")},
		MaxTokens: 100,
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello from Claude", result.Content)
	assert.Equal(t, 10, result.InputTokens)
	assert.Equal(t, 5, result.OutputTokens)
	assert.Equal(t, "claude", result.Provider)
}

func TestClaude_Complete_WithSystemAndJSONMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req claudeRequest
		require.NoError(t, json.Unmarshal(body, &req))

		// System messages should be extracted and JSON mode instruction prepended
		assert.Contains(t, req.System, "valid JSON only")
		assert.Contains(t, req.System, "You are a helpful assistant")
		assert.Len(t, req.Messages, 1) // only user message, system extracted

		resp := claudeResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: `{"result": "ok"}`}},
			Model: "claude-sonnet-4-6",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewClaudeProvider("key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{
			NewSystemMessage("You are a helpful assistant"),
			NewUserMessage("Give me JSON"),
		},
		JSONMode: true,
	})
	require.NoError(t, err)
	assert.Equal(t, `{"result": "ok"}`, result.Content)
}

func TestClaude_Complete_APIError(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"unauthorized", http.StatusUnauthorized},
		{"rate_limited", http.StatusTooManyRequests},
		{"server_error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"error": "test error"}`))
			}))
			defer server.Close()

			p := NewClaudeProvider("key", "")
			p.BaseURL = server.URL

			_, err := p.Complete(context.Background(), CompletionRequest{
				Messages: []Message{NewUserMessage("hello")},
			})
			require.Error(t, err)
			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.status, apiErr.StatusCode)
			assert.Equal(t, "claude", apiErr.Provider)
		})
	}
}

func TestClaude_Complete_ValidationError(t *testing.T) {
	p := NewClaudeProvider("key", "")
	_, err := p.Complete(context.Background(), CompletionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestClaude_DefaultModel(t *testing.T) {
	p := NewClaudeProvider("key", "")
	assert.Equal(t, ClaudeSonnet4, p.DefaultModel())
	assert.Equal(t, "claude", p.Name())
}

func TestClaude_CustomModel(t *testing.T) {
	p := NewClaudeProvider("key", ClaudeOpus4)
	assert.Equal(t, ClaudeOpus4, p.DefaultModel())
}
