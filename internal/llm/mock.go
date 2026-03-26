package llm

import (
	"context"
	"fmt"
	"sync"
)

// MockProvider is a test double that returns pre-configured responses.
type MockProvider struct {
	mu        sync.Mutex
	responses []string
	index     int
	Requests  []CompletionRequest // recorded for assertion
	name      string
	model     string
}

// NewMockProvider creates a MockProvider that returns the given responses
// sequentially. After all responses are exhausted, further calls return an error.
func NewMockProvider(responses ...string) *MockProvider {
	return &MockProvider{
		responses: responses,
		name:      "mock",
		model:     "mock-model",
	}
}

func (p *MockProvider) Name() string         { return p.name }
func (p *MockProvider) DefaultModel() string { return p.model }

func (p *MockProvider) Complete(_ context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.Requests = append(p.Requests, req)

	if p.index >= len(p.responses) {
		return nil, fmt.Errorf("llm: mock provider exhausted (%d responses consumed)", len(p.responses))
	}

	content := p.responses[p.index]
	p.index++

	return &CompletionResponse{
		Content:      content,
		InputTokens:  100,
		OutputTokens: 50,
		Model:        p.model,
		Provider:     p.name,
	}, nil
}

// RequestCount returns the number of requests made to this mock.
func (p *MockProvider) RequestCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.Requests)
}

// LastRequest returns the most recent request, or nil if none.
func (p *MockProvider) LastRequest() *CompletionRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Requests) == 0 {
		return nil
	}
	r := p.Requests[len(p.Requests)-1]
	return &r
}
