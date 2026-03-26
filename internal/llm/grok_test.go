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

func TestGrok_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		var req openAIRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "grok-3-mini", req.Model)
		assert.Len(t, req.Messages, 1)

		resp := openAIResponse{
			Model: "grok-3-mini",
		}
		resp.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{{Message: struct {
			Content string `json:"content"`
		}{Content: "Hello from Grok"}}}
		resp.Usage.PromptTokens = 12
		resp.Usage.CompletionTokens = 6

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGrokProvider("test-key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("Hello")},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello from Grok", result.Content)
	assert.Equal(t, 12, result.InputTokens)
	assert.Equal(t, 6, result.OutputTokens)
	assert.Equal(t, "grok", result.Provider)
}

func TestGrok_Complete_JSONMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req openAIRequest
		require.NoError(t, json.Unmarshal(body, &req))

		require.NotNil(t, req.ResponseFormat)
		assert.Equal(t, "json_object", req.ResponseFormat.Type)

		resp := openAIResponse{Model: "grok-3-mini"}
		resp.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{{Message: struct {
			Content string `json:"content"`
		}{Content: `{"result": true}`}}}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGrokProvider("key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("JSON")},
		JSONMode: true,
	})
	require.NoError(t, err)
	assert.Equal(t, `{"result": true}`, result.Content)
}

func TestGrok_Complete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	p := NewGrokProvider("key", "")
	p.BaseURL = server.URL

	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("hello")},
	})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	assert.Equal(t, "grok", apiErr.Provider)
}

func TestGrok_DefaultModel(t *testing.T) {
	p := NewGrokProvider("key", "")
	assert.Equal(t, Grok3Mini, p.DefaultModel())
	assert.Equal(t, "grok", p.Name())
}

func TestGrok_CustomModel(t *testing.T) {
	p := NewGrokProvider("key", Grok3)
	assert.Equal(t, Grok3, p.DefaultModel())
}
