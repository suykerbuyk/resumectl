package pipeline

import (
	"context"
	"fmt"
	"sort"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

// MatchResult holds the results of a gap analysis.
type MatchResult struct {
	StrongMatches  []MatchedSkill `json:"strong_matches"`
	PartialMatches []MatchedSkill `json:"partial_matches"`
	Gaps           []GapItem      `json:"gaps"`
	OverallScore   float64        `json:"overall_score"`
	Narrative      string         `json:"narrative"`
}

// MatchedSkill represents a skill that matches between experience and job.
type MatchedSkill struct {
	Skill          string   `json:"skill"`
	ExperienceRefs []string `json:"evidence"`
	Confidence     float64  `json:"confidence"`
	GapNote        string   `json:"gap_note,omitempty"`
}

// GapItem represents a skill gap.
type GapItem struct {
	Skill      string `json:"skill"`
	Required   bool   `json:"required"`
	Suggestion string `json:"suggestion"`
}

// ScoredExperience pairs an experience file with its relevance score.
type ScoredExperience struct {
	File  *vault.ExperienceFile
	Score float64
}

// RunMatch performs a two-stage gap analysis: local tag intersection followed
// by LLM semantic scoring.
func RunMatch(
	ctx context.Context,
	job *vault.JobFile,
	v *vault.Vault,
	lib *prompts.Library,
	provider llm.Provider,
	topN int,
) (*MatchResult, error) {
	if topN <= 0 {
		topN = 5
	}

	// Stage 1: Local tag intersection
	if err := v.EnsureIndex(); err != nil {
		return nil, fmt.Errorf("pipeline: match index: %w", err)
	}

	allSkills := make([]string, 0, len(job.Frontmatter.RequiredSkills)+len(job.Frontmatter.PreferredSkills))
	allSkills = append(allSkills, job.Frontmatter.RequiredSkills...)
	allSkills = append(allSkills, job.Frontmatter.PreferredSkills...)
	candidates := v.Index.Query(job.Frontmatter.Tags, allSkills)

	// Score and sort candidates
	scored := make([]ScoredExperience, len(candidates))
	for i, c := range candidates {
		scored[i] = ScoredExperience{
			File:  c,
			Score: v.Index.Score(c, job),
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Take top N
	if len(scored) > topN {
		scored = scored[:topN]
	}

	topFiles := make([]*vault.ExperienceFile, len(scored))
	for i, s := range scored {
		topFiles[i] = s.File
	}

	// Stage 2: LLM semantic gap analysis
	skills, err := v.LoadSkills()
	if err != nil {
		skills = &vault.SkillsInventory{Body: "No skills inventory available."}
	}

	rendered, err := lib.Render("gap_analysis", prompts.GapAnalysisContext{
		JobContent:      job.Body,
		ExperienceFiles: topFiles,
		SkillsInventory: skills.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: match render: %w", err)
	}

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{llm.NewUserMessage(rendered)},
		MaxTokens: 3000,
		JSONMode:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: match LLM: %w", err)
	}

	result, err := ParseJSONResponse[MatchResult](resp.Content)
	if err != nil {
		return nil, fmt.Errorf("pipeline: match: %w", err)
	}

	return &result, nil
}
