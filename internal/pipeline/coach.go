package pipeline

import (
	"context"
	"fmt"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
)

// CoachInput holds the input for a single coaching question.
type CoachInput struct {
	Gap             GapItem
	JobTitle        string
	JobCompany      string
	ExistingContext string // relevant experience context for this gap
}

// CoachSuggestion is the parsed LLM response for a coaching turn.
type CoachSuggestion struct {
	Question   string `json:"question"`
	Context    string `json:"context"`
	Suggestion struct {
		TargetFile string `json:"target_file"`
		UpdateType string `json:"update_type"`
		Content    string `json:"content"`
	} `json:"suggestion"`
}

// CoachResult holds the output of a single coaching turn.
type CoachResult struct {
	Question   string
	Context    string
	Suggestion CoachSuggestion
}

// RunCoach generates a coaching question and vault update suggestion for a
// single gap item. Does NOT perform any I/O — the CLI layer handles
// interactive prompting and vault writes.
func RunCoach(
	ctx context.Context,
	input CoachInput,
	lib *prompts.Library,
	provider llm.Provider,
) (*CoachResult, error) {
	gapType := "preferred"
	if input.Gap.Required {
		gapType = "hard"
	}

	rendered, err := lib.Render("coach_question", prompts.CoachContext{
		Skill:           input.Gap.Skill,
		Required:        input.Gap.Required,
		GapType:         gapType,
		JobTitle:        input.JobTitle,
		JobCompany:      input.JobCompany,
		ExistingContext: input.ExistingContext,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: coach render: %w", err)
	}

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{llm.NewUserMessage(rendered)},
		MaxTokens: 1000,
		JSONMode:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: coach LLM: %w", err)
	}

	suggestion, err := ParseJSONResponse[CoachSuggestion](resp.Content)
	if err != nil {
		return nil, fmt.Errorf("pipeline: coach: %w", err)
	}

	return &CoachResult{
		Question:   suggestion.Question,
		Context:    suggestion.Context,
		Suggestion: suggestion,
	}, nil
}
