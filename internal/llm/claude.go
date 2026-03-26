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
	ClaudeOpus4   = "claude-opus-4-6"
	ClaudeSonnet4 = "claude-sonnet-4-6"
	ClaudeHaiku4  = "claude-haiku-4-5-20251001"

	defaultClaudeBaseURL = "https://api.anthropic.com"
	anthropicVersion     = "2023-06-01"
)

// ClaudeProvider implements the Provider interface for Anthropic's Claude API.
type ClaudeProvider struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	model      string
}

// NewClaudeProvider creates a new Claude provider with the given API key and model.
func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	if model == "" {
		model = ClaudeSonnet4
	}
	return &ClaudeProvider{
		APIKey:     apiKey,
		BaseURL:    defaultClaudeBaseURL,
		HTTPClient: http.DefaultClient,
		model:      model,
	}
}

func (p *ClaudeProvider) Name() string         { return "claude" }
func (p *ClaudeProvider) DefaultModel() string { return p.model }

// claudeRequest is the Anthropic Messages API request body.
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse is the Anthropic Messages API response body.
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *ClaudeProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Separate system messages from conversation messages
	var systemPrompt string
	var msgs []claudeMessage
	for _, m := range req.Messages {
		if m.Role == "system" {
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += m.Content
		} else {
			msgs = append(msgs, claudeMessage(m))
		}
	}

	// If JSON mode, prepend instruction to system prompt
	if req.JSONMode {
		jsonInstruction := "You must respond with valid JSON only. No preamble, no explanation, no markdown formatting."
		if systemPrompt != "" {
			systemPrompt = jsonInstruction + "\n\n" + systemPrompt
		} else {
			systemPrompt = jsonInstruction
		}
	}

	body := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  msgs,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("llm: claude marshal request: %w", err)
	}

	url := p.BaseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("llm: claude create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm: claude request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: claude read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Provider:   "claude",
			Body:       string(respBody),
		}
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("llm: claude parse response: %w", err)
	}

	var content string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &CompletionResponse{
		Content:      content,
		InputTokens:  claudeResp.Usage.InputTokens,
		OutputTokens: claudeResp.Usage.OutputTokens,
		Model:        claudeResp.Model,
		Provider:     "claude",
	}, nil
}
