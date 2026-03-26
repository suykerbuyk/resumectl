package llm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_Valid(t *testing.T) {
	req := &CompletionRequest{
		Messages:  []Message{NewUserMessage("hello")},
		MaxTokens: 100,
	}
	assert.NoError(t, req.Validate())
}

func TestValidate_NoMessages(t *testing.T) {
	req := &CompletionRequest{}
	err := req.Validate()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoMessages))
}

func TestValidate_NegativeMaxTokens(t *testing.T) {
	req := &CompletionRequest{
		Messages:  []Message{NewUserMessage("hello")},
		MaxTokens: -1,
	}
	err := req.Validate()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNegativeMaxTokens))
}

func TestValidate_InvalidRole(t *testing.T) {
	req := &CompletionRequest{
		Messages: []Message{{Role: "invalid", Content: "test"}},
	}
	err := req.Validate()
	require.Error(t, err)
	var roleErr *InvalidRoleError
	assert.True(t, errors.As(err, &roleErr))
	assert.Equal(t, "invalid", roleErr.Role)
}

func TestValidate_AllRoles(t *testing.T) {
	req := &CompletionRequest{
		Messages: []Message{
			NewSystemMessage("sys"),
			NewUserMessage("usr"),
			NewAssistantMessage("ast"),
		},
	}
	assert.NoError(t, req.Validate())
}

func TestNewMessages(t *testing.T) {
	sys := NewSystemMessage("sys content")
	assert.Equal(t, "system", sys.Role)
	assert.Equal(t, "sys content", sys.Content)

	usr := NewUserMessage("usr content")
	assert.Equal(t, "user", usr.Role)

	ast := NewAssistantMessage("ast content")
	assert.Equal(t, "assistant", ast.Role)
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 401, Provider: "claude", Body: "unauthorized"}
	assert.Contains(t, err.Error(), "claude")
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestInvalidRoleError(t *testing.T) {
	err := &InvalidRoleError{Role: "bogus"}
	assert.Contains(t, err.Error(), "bogus")
}
