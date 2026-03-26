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

func TestOpenAI_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		var req openAIRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "gpt-4o-mini", req.Model)
		assert.Len(t, req.Messages, 1)

		resp := openAIResponse{
			Model: "gpt-4o-mini",
		}
		resp.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{{Message: struct {
			Content string `json:"content"`
		}{Content: "Hello from OpenAI"}}}
		resp.Usage.PromptTokens = 15
		resp.Usage.CompletionTokens = 8

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider("test-key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("Hello")},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello from OpenAI", result.Content)
	assert.Equal(t, 15, result.InputTokens)
	assert.Equal(t, 8, result.OutputTokens)
	assert.Equal(t, "openai", result.Provider)
}

func TestOpenAI_Complete_JSONMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req openAIRequest
		require.NoError(t, json.Unmarshal(body, &req))

		require.NotNil(t, req.ResponseFormat)
		assert.Equal(t, "json_object", req.ResponseFormat.Type)

		resp := openAIResponse{Model: "gpt-4o-mini"}
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

	p := NewOpenAIProvider("key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("JSON")},
		JSONMode: true,
	})
	require.NoError(t, err)
	assert.Equal(t, `{"result": true}`, result.Content)
}

func TestOpenAI_Complete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid api key"))
	}))
	defer server.Close()

	p := NewOpenAIProvider("bad-key", "")
	p.BaseURL = server.URL

	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("hello")},
	})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	assert.Equal(t, "openai", apiErr.Provider)
}

func TestOpenAI_DefaultModel(t *testing.T) {
	p := NewOpenAIProvider("key", "")
	assert.Equal(t, GPT4oMini, p.DefaultModel())
	assert.Equal(t, "openai", p.Name())
}

func TestOpenAI_CustomModel(t *testing.T) {
	p := NewOpenAIProvider("key", GPT4o)
	assert.Equal(t, GPT4o, p.DefaultModel())
}
