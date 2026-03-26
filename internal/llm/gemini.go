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
	GeminiFlash = "gemini-2.0-flash"
	GeminiPro   = "gemini-2.0-pro"

	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"
)

// GeminiProvider implements the Provider interface for Google's Gemini API.
type GeminiProvider struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	model      string
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = GeminiFlash
	}
	return &GeminiProvider{
		APIKey:     apiKey,
		BaseURL:    defaultGeminiBaseURL,
		HTTPClient: http.DefaultClient,
		model:      model,
	}
}

func (p *GeminiProvider) Name() string         { return "gemini" }
func (p *GeminiProvider) DefaultModel() string { return p.model }

// Gemini API types
type geminiRequest struct {
	Contents          []geminiContent  `json:"contents"`
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	GenerationConfig  *geminiGenConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	MaxOutputTokens  int     `json:"maxOutputTokens,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	// Build contents and extract system instruction
	var systemInstruction *geminiContent
	var contents []geminiContent
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: m.Content}},
			}
		} else {
			role := m.Role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, geminiContent{
				Role:  role,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	genConfig := &geminiGenConfig{}
	if req.MaxTokens > 0 {
		genConfig.MaxOutputTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		genConfig.Temperature = req.Temperature
	}
	if req.JSONMode {
		genConfig.ResponseMimeType = "application/json"
	}

	body := geminiRequest{
		Contents:          contents,
		SystemInstruction: systemInstruction,
		GenerationConfig:  genConfig,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("llm: gemini marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.BaseURL, model, p.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("llm: gemini create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm: gemini request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: gemini read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Provider:   "gemini",
			Body:       string(respBody),
		}
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(respBody, &gemResp); err != nil {
		return nil, fmt.Errorf("llm: gemini parse response: %w", err)
	}

	var content string
	if len(gemResp.Candidates) > 0 {
		for _, part := range gemResp.Candidates[0].Content.Parts {
			content += part.Text
		}
	}

	return &CompletionResponse{
		Content:      content,
		InputTokens:  gemResp.UsageMetadata.PromptTokenCount,
		OutputTokens: gemResp.UsageMetadata.CandidatesTokenCount,
		Model:        model,
		Provider:     "gemini",
	}, nil
}
