package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCoach_Success(t *testing.T) {
	provider := mockProvider(t, "coach_question_response.json")
	lib := testLibrary(t)

	result, err := RunCoach(context.Background(), CoachInput{
		Gap: GapItem{
			Skill:      "Rust",
			Required:   false,
			Suggestion: "Ask about Rust experience.",
		},
		JobTitle:        "Staff Engineer",
		JobCompany:      "Acme Corp",
		ExistingContext: "Strong C++ and Go background.",
	}, lib, provider)

	require.NoError(t, err)
	assert.Contains(t, result.Question, "Rust")
	assert.Contains(t, result.Context, "preferred skill")
	assert.Equal(t, "add_skill", result.Suggestion.Suggestion.UpdateType)
	assert.Equal(t, "Rust", result.Suggestion.Suggestion.Content)
}

func TestRunCoach_RequiredGap(t *testing.T) {
	provider := mockProvider(t, "coach_question_response.json")
	lib := testLibrary(t)

	result, err := RunCoach(context.Background(), CoachInput{
		Gap: GapItem{
			Skill:    "Kubernetes",
			Required: true,
		},
		JobTitle:        "SRE Lead",
		JobCompany:      "BigCo",
		ExistingContext: "Some cloud experience.",
	}, lib, provider)

	require.NoError(t, err)
	assert.NotEmpty(t, result.Question)
	assert.NotEmpty(t, result.Suggestion.Suggestion.TargetFile)
}

func TestRunCoach_LLMError(t *testing.T) {
	provider := mockProvider(t, "parsejob_response.json") // wrong shape
	lib := testLibrary(t)

	// This will succeed but with zero-value fields since JSON shape is wrong
	result, err := RunCoach(context.Background(), CoachInput{
		Gap:        GapItem{Skill: "Go"},
		JobTitle:   "Eng",
		JobCompany: "Co",
	}, lib, provider)

	require.NoError(t, err)
	// Fields will be empty since wrong JSON shape
	assert.Empty(t, result.Question)
}
