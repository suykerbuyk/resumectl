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

func TestGemini_Complete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "gemini-2.0-flash")
		assert.Contains(t, r.URL.RawQuery, "key=test-key")

		body, _ := io.ReadAll(r.Body)
		var req geminiRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Len(t, req.Contents, 1)
		assert.Equal(t, "user", req.Contents[0].Role)

		resp := geminiResponse{}
		resp.Candidates = []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}{
			{Content: struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			}{Parts: []struct {
				Text string `json:"text"`
			}{{Text: "Hello from Gemini"}}}},
		}
		resp.UsageMetadata.PromptTokenCount = 15
		resp.UsageMetadata.CandidatesTokenCount = 8

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiProvider("test-key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("Hello")},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello from Gemini", result.Content)
	assert.Equal(t, 15, result.InputTokens)
	assert.Equal(t, 8, result.OutputTokens)
	assert.Equal(t, "gemini", result.Provider)
}

func TestGemini_Complete_WithSystemAndJSONMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req geminiRequest
		require.NoError(t, json.Unmarshal(body, &req))

		// System instruction should be set
		require.NotNil(t, req.SystemInstruction)
		assert.Equal(t, "Be helpful", req.SystemInstruction.Parts[0].Text)

		// JSON mode should set response mime type
		require.NotNil(t, req.GenerationConfig)
		assert.Equal(t, "application/json", req.GenerationConfig.ResponseMimeType)

		// Assistant role should be mapped to "model"
		assert.Equal(t, "user", req.Contents[0].Role)

		resp := geminiResponse{}
		resp.Candidates = []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}{
			{Content: struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			}{Parts: []struct {
				Text string `json:"text"`
			}{{Text: `{"ok": true}`}}}},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGeminiProvider("key", "")
	p.BaseURL = server.URL

	result, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{
			NewSystemMessage("Be helpful"),
			NewUserMessage("JSON please"),
		},
		JSONMode: true,
	})
	require.NoError(t, err)
	assert.Equal(t, `{"ok": true}`, result.Content)
}

func TestGemini_Complete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	p := NewGeminiProvider("key", "")
	p.BaseURL = server.URL

	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("hello")},
	})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Equal(t, "gemini", apiErr.Provider)
}

func TestGemini_DefaultModel(t *testing.T) {
	p := NewGeminiProvider("key", "")
	assert.Equal(t, GeminiFlash, p.DefaultModel())
	assert.Equal(t, "gemini", p.Name())
}
