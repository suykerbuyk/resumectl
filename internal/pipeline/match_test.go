package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMatch_Success(t *testing.T) {
	v := openTestVault(t)
	provider := mockProvider(t, "gap_analysis_response.json")
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	result, err := RunMatch(context.Background(), job, v, lib, provider, 5)
	require.NoError(t, err)

	assert.Equal(t, 0.78, result.OverallScore)
	assert.Contains(t, result.Narrative, "Strong match")
	assert.Len(t, result.StrongMatches, 2)
	assert.Len(t, result.PartialMatches, 1)
	assert.Len(t, result.Gaps, 2)

	// Verify strong match content
	assert.Equal(t, "Go", result.StrongMatches[0].Skill)
	assert.Equal(t, 0.95, result.StrongMatches[0].Confidence)

	// Verify gap content
	assert.Equal(t, "Rust", result.Gaps[0].Skill)
	assert.False(t, result.Gaps[0].Required)
}

func TestRunMatch_DefaultTopN(t *testing.T) {
	v := openTestVault(t)
	provider := mockProvider(t, "gap_analysis_response.json")
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	// topN=0 should default to 5
	result, err := RunMatch(context.Background(), job, v, lib, provider, 0)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRunMatch_LLMError(t *testing.T) {
	v := openTestVault(t)
	// Provider with invalid JSON to trigger parse error
	provider := mockProvider(t, "parsejob_response.json") // wrong shape for MatchResult
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	// This should still succeed since parsejob response is valid JSON,
	// just wrong shape — the unmarshal will populate zero values
	result, err := RunMatch(context.Background(), job, v, lib, provider, 5)
	require.NoError(t, err)
	// Score will be zero since it's the wrong JSON shape
	assert.Equal(t, 0.0, result.OverallScore)
}
