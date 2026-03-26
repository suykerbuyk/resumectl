package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	Grok3     = "grok-3"
	Grok3Mini = "grok-3-mini"

	defaultGrokBaseURL = "https://api.x.ai/v1"
)

// GrokProvider implements the Provider interface using the OpenAI-compatible
// xAI Grok API.
type GrokProvider struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	model      string
}

// NewGrokProvider creates a new Grok provider.
func NewGrokProvider(apiKey, model string) *GrokProvider {
	if model == "" {
		model = Grok3Mini
	}
	return &GrokProvider{
		APIKey:     apiKey,
		BaseURL:    defaultGrokBaseURL,
		HTTPClient: http.DefaultClient,
		model:      model,
	}
}

func (p *GrokProvider) Name() string         { return "grok" }
func (p *GrokProvider) DefaultModel() string { return p.model }

// OpenAI-compatible request/response types
type openAIRequest struct {
	Model          string          `json:"model"`
	Messages       []openAIMessage `json:"messages"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (p *GrokProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	var msgs []openAIMessage
	for _, m := range req.Messages {
		msgs = append(msgs, openAIMessage(m))
	}

	body := openAIRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	if req.JSONMode {
		body.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("llm: grok marshal request: %w", err)
	}

	url := p.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("llm: grok create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm: grok request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: grok read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Provider:   "grok",
			Body:       string(respBody),
		}
	}

	var oaiResp openAIResponse
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return nil, fmt.Errorf("llm: grok parse response: %w", err)
	}

	var content string
	if len(oaiResp.Choices) > 0 {
		content = oaiResp.Choices[0].Message.Content
	}

	return &CompletionResponse{
		Content:      content,
		InputTokens:  oaiResp.Usage.PromptTokens,
		OutputTokens: oaiResp.Usage.CompletionTokens,
		Model:        oaiResp.Model,
		Provider:     "grok",
	}, nil
}
