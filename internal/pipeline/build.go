package pipeline

import (
	"context"
	"fmt"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

// BuildInput holds the input for the resume build pipeline.
type BuildInput struct {
	Job             *vault.JobFile
	ExperienceFiles []*vault.ExperienceFile
	CoverLetter     bool
}

// BuildResult holds the output of the resume build pipeline.
type BuildResult struct {
	Resume      *vault.ResumeFile
	CoverLetter string // empty if not requested
	Warnings    []string
}

// RunBuild synthesizes a targeted resume (and optionally a cover letter)
// using the given experience files and job context.
func RunBuild(
	ctx context.Context,
	input BuildInput,
	v *vault.Vault,
	lib *prompts.Library,
	provider llm.Provider,
	model string,
) (*BuildResult, error) {
	if len(input.ExperienceFiles) == 0 {
		return nil, fmt.Errorf("pipeline: build: no experience files provided")
	}

	// Load profile data
	contact, err := v.LoadContact()
	if err != nil {
		return nil, fmt.Errorf("pipeline: build load contact: %w", err)
	}
	summary, err := v.LoadSummary()
	if err != nil {
		return nil, fmt.Errorf("pipeline: build load summary: %w", err)
	}
	skills, err := v.LoadSkills()
	if err != nil {
		skills = &vault.SkillsInventory{Body: "No skills inventory available."}
	}

	contactStr := fmt.Sprintf("%s | %s | %s | %s",
		contact.Name, contact.Email, contact.Location, contact.LinkedIn)

	synthCtx := prompts.SynthesizeContext{
		ProfileSummary:  summary.Body,
		JobContent:      input.Job.Body,
		ExperienceFiles: input.ExperienceFiles,
		SkillsInventory: skills.Body,
		ContactInfo:     contactStr,
	}

	// Render and call LLM for resume
	rendered, err := lib.Render("synthesize_resume", synthCtx)
	if err != nil {
		return nil, fmt.Errorf("pipeline: build render: %w", err)
	}

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{llm.NewUserMessage(rendered)},
		MaxTokens: 4000,
		Model:     model,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: build LLM: %w", err)
	}

	// Validate output
	companies := make([]string, len(input.ExperienceFiles))
	for i, ef := range input.ExperienceFiles {
		companies[i] = ef.Frontmatter.Company
	}
	warnings := ValidateResumeOutput(resp.Content, companies)

	// Build experience file paths for frontmatter
	expPaths := make([]string, len(input.ExperienceFiles))
	for i, ef := range input.ExperienceFiles {
		expPaths[i] = ef.RelPath
	}

	resume := &vault.ResumeFile{
		Frontmatter: vault.ResumeFrontmatter{
			JobFile:         input.Job.RelPath,
			Model:           resp.Model,
			Status:          "draft",
			ExperienceFiles: expPaths,
			Version:         1,
		},
		Body: resp.Content,
	}

	result := &BuildResult{
		Resume:   resume,
		Warnings: warnings,
	}

	// Optionally build cover letter
	if input.CoverLetter {
		clRendered, err := lib.Render("synthesize_cover_letter", synthCtx)
		if err != nil {
			return nil, fmt.Errorf("pipeline: build cover letter render: %w", err)
		}

		clResp, err := provider.Complete(ctx, llm.CompletionRequest{
			Messages:  []llm.Message{llm.NewUserMessage(clRendered)},
			MaxTokens: 2000,
			Model:     model,
		})
		if err != nil {
			return nil, fmt.Errorf("pipeline: build cover letter LLM: %w", err)
		}
		result.CoverLetter = clResp.Content
	}

	return result, nil
}
