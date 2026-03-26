package pipeline

import (
	"context"
	"fmt"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

// IngestInput holds the input for the ingestion pipeline.
type IngestInput struct {
	SourceText string // extracted plaintext of source document
	SourcePath string // original file path or URL for attribution
}

// IngestResult holds the output of the ingestion pipeline.
type IngestResult struct {
	ExperienceFiles []*vault.ExperienceFile
	SkillsUpdates   []string // suggested additions to skills.md
	Warnings        []string
}

// ingestedRole is the JSON schema for a single role returned by the LLM.
type ingestedRole struct {
	Role          string   `json:"role"`
	Company       string   `json:"company"`
	CompanySlug   string   `json:"company_slug"`
	Start         string   `json:"start"`
	End           string   `json:"end"`
	Current       bool     `json:"current"`
	Tags          []string `json:"tags"`
	Skills        []string `json:"skills"`
	Domain        string   `json:"domain"`
	Summary       string   `json:"summary"`
	Contributions []string `json:"contributions"`
	Technologies  []string `json:"technologies"`
}

// RunIngest decomposes source text into structured experience files via LLM.
func RunIngest(
	ctx context.Context,
	input IngestInput,
	v *vault.Vault,
	lib *prompts.Library,
	provider llm.Provider,
) (*IngestResult, error) {
	rendered, err := lib.Render("decompose_resume", prompts.DecomposeContext{
		SourceText: input.SourceText,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: ingest render: %w", err)
	}

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{llm.NewUserMessage(rendered)},
		MaxTokens: 4000,
		JSONMode:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: ingest LLM: %w", err)
	}

	roles, err := ParseJSONResponse[[]ingestedRole](resp.Content)
	if err != nil {
		return nil, fmt.Errorf("pipeline: ingest: %w", err)
	}

	result := &IngestResult{}
	allSkills := make(map[string]bool)

	// Load existing experience for dedup checking
	existing, _ := v.AllExperience()

	for _, role := range roles {
		// Build body
		body := "## Summary\n\n" + role.Summary + "\n"
		if len(role.Contributions) > 0 {
			body += "\n## Key Contributions\n\n"
			for _, c := range role.Contributions {
				body += "- " + c + "\n"
			}
		}
		if len(role.Technologies) > 0 {
			body += "\n## Technologies\n\n"
			for i, tech := range role.Technologies {
				if i > 0 {
					body += ", "
				}
				body += tech
			}
			body += "\n"
		}

		ef := &vault.ExperienceFile{
			Frontmatter: vault.ExperienceFrontmatter{
				Role:        role.Role,
				Company:     role.Company,
				CompanySlug: role.CompanySlug,
				Start:       role.Start,
				End:         role.End,
				Current:     role.Current,
				Tags:        role.Tags,
				Skills:      role.Skills,
				Domain:      role.Domain,
				Visibility:  "resume",
			},
			Body: body,
		}

		// Dedup check
		if isDuplicate(ef, existing) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("possible duplicate: %s at %s (%s)", role.Role, role.Company, role.Start))
			continue
		}

		result.ExperienceFiles = append(result.ExperienceFiles, ef)

		for _, s := range role.Skills {
			allSkills[s] = true
		}
	}

	for skill := range allSkills {
		result.SkillsUpdates = append(result.SkillsUpdates, skill)
	}

	return result, nil
}

// isDuplicate checks if an experience file likely duplicates an existing one.
func isDuplicate(ef *vault.ExperienceFile, existing []*vault.ExperienceFile) bool {
	for _, e := range existing {
		if e.Frontmatter.CompanySlug == ef.Frontmatter.CompanySlug &&
			e.Frontmatter.Start == ef.Frontmatter.Start {
			return true
		}
	}
	return false
}
