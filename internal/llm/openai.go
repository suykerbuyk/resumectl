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
	GPT4o     = "gpt-4o"
	GPT4oMini = "gpt-4o-mini"
	O3Mini    = "o3-mini"

	defaultOpenAIBaseURL = "https://api.openai.com/v1"
)

// OpenAIProvider implements the Provider interface for the OpenAI API.
type OpenAIProvider struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	model      string
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = GPT4oMini
	}
	return &OpenAIProvider{
		APIKey:     apiKey,
		BaseURL:    defaultOpenAIBaseURL,
		HTTPClient: http.DefaultClient,
		model:      model,
	}
}

func (p *OpenAIProvider) Name() string         { return "openai" }
func (p *OpenAIProvider) DefaultModel() string { return p.model }

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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
		return nil, fmt.Errorf("llm: openai marshal request: %w", err)
	}

	url := p.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("llm: openai create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm: openai request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: openai read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Provider:   "openai",
			Body:       string(respBody),
		}
	}

	var oaiResp openAIResponse
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return nil, fmt.Errorf("llm: openai parse response: %w", err)
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
		Provider:     "openai",
	}, nil
}
