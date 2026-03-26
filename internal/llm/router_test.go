package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_ForTask_Success(t *testing.T) {
	mock := NewMockProvider("ok")
	providers := map[string]Provider{
		"claude": mock,
	}
	taskMap := map[string]string{
		"decompose":    "claude",
		"extract_job":  "claude",
		"gap_analysis": "claude",
	}

	r := NewRouter(providers, taskMap)

	p, err := r.ForTask(TaskDecompose)
	require.NoError(t, err)
	assert.Equal(t, "mock", p.Name())

	p, err = r.ForTask(TaskExtractJob)
	require.NoError(t, err)
	assert.NotNil(t, p)

	p, err = r.ForTask(TaskGapAnalysis)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestRouter_ForTask_MultipleProviders(t *testing.T) {
	claude := NewMockProvider("claude-resp")
	claude.name = "claude"
	gemini := NewMockProvider("gemini-resp")
	gemini.name = "gemini"

	providers := map[string]Provider{
		"claude": claude,
		"gemini": gemini,
	}
	taskMap := map[string]string{
		"decompose":   "claude",
		"extract_job": "gemini",
	}

	r := NewRouter(providers, taskMap)

	p, err := r.ForTask(TaskDecompose)
	require.NoError(t, err)
	assert.Equal(t, "claude", p.Name())

	p, err = r.ForTask(TaskExtractJob)
	require.NoError(t, err)
	assert.Equal(t, "gemini", p.Name())
}

func TestRouter_ForTask_UnknownTask(t *testing.T) {
	r := NewRouter(map[string]Provider{}, map[string]string{})
	_, err := r.ForTask(TaskDecompose)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoProvider)
}

func TestRouter_ForTask_MissingProvider(t *testing.T) {
	taskMap := map[string]string{
		"decompose": "nonexistent",
	}
	r := NewRouter(map[string]Provider{}, taskMap)

	_, err := r.ForTask(TaskDecompose)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderNotFound)
}

func TestRouter_AllTaskTypes(t *testing.T) {
	mock := NewMockProvider("ok")
	providers := map[string]Provider{"mock": mock}
	taskMap := map[string]string{
		"decompose":         "mock",
		"extract_job":       "mock",
		"gap_analysis":      "mock",
		"synthesize_resume": "mock",
		"cover_letter":      "mock",
		"coach":             "mock",
	}

	r := NewRouter(providers, taskMap)

	for _, task := range []TaskType{
		TaskDecompose, TaskExtractJob, TaskGapAnalysis,
		TaskSynthesizeResume, TaskCoverLetter, TaskCoach,
	} {
		p, err := r.ForTask(task)
		require.NoError(t, err, "task: %s", task)
		assert.NotNil(t, p)
	}
}
