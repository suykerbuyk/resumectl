package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMock_Returns_CannedResponse(t *testing.T) {
	p := NewMockProvider("response one", "response two")

	r1, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("first")},
	})
	require.NoError(t, err)
	assert.Equal(t, "response one", r1.Content)
	assert.Equal(t, "mock", r1.Provider)

	r2, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("second")},
	})
	require.NoError(t, err)
	assert.Equal(t, "response two", r2.Content)
}

func TestMock_Records_Requests(t *testing.T) {
	p := NewMockProvider("ok")
	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("test message")},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, p.RequestCount())
	assert.Equal(t, "test message", p.LastRequest().Messages[0].Content)
}

func TestMock_Exhausted(t *testing.T) {
	p := NewMockProvider("only one")
	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("first")},
	})
	require.NoError(t, err)

	_, err = p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("second")},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}

func TestMock_Validates_Request(t *testing.T) {
	p := NewMockProvider("ok")
	_, err := p.Complete(context.Background(), CompletionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestMock_EmptyResponses(t *testing.T) {
	p := NewMockProvider()
	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{NewUserMessage("hello")},
	})
	require.Error(t, err)
}

func TestMock_NameAndModel(t *testing.T) {
	p := NewMockProvider()
	assert.Equal(t, "mock", p.Name())
	assert.Equal(t, "mock-model", p.DefaultModel())
}

func TestMock_LastRequest_Empty(t *testing.T) {
	p := NewMockProvider()
	assert.Nil(t, p.LastRequest())
}
