package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

// ParseJobInput holds the input for the job parsing pipeline.
type ParseJobInput struct {
	PostingText string // extracted plaintext of the job posting
	SourceURL   string // original URL for attribution
	Slug        string // optional filename slug override
}

// parsedJob is the JSON schema returned by the LLM.
type parsedJob struct {
	Title           string              `json:"title"`
	Company         string              `json:"company"`
	CompanySlug     string              `json:"company_slug"`
	Location        string              `json:"location"`
	Seniority       string              `json:"seniority"`
	Domain          string              `json:"domain"`
	RequiredSkills  []string            `json:"required_skills"`
	PreferredSkills []string            `json:"preferred_skills"`
	CultureSignals  []string            `json:"culture_signals"`
	Compensation    *vault.Compensation `json:"compensation"`
	Summary         string              `json:"summary"`
	KeyRequirements struct {
		MustHave   []string `json:"must_have"`
		NiceToHave []string `json:"nice_to_have"`
	} `json:"key_requirements"`
	RawText string `json:"raw_text"`
}

// RunParseJob extracts structured job data from raw posting text via LLM.
func RunParseJob(
	ctx context.Context,
	input ParseJobInput,
	lib *prompts.Library,
	provider llm.Provider,
) (*vault.JobFile, error) {
	rendered, err := lib.Render("extract_job", prompts.ExtractJobContext{
		PostingText: input.PostingText,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: parse-job render: %w", err)
	}

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{llm.NewUserMessage(rendered)},
		MaxTokens: 2000,
		JSONMode:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("pipeline: parse-job LLM: %w", err)
	}

	parsed, err := ParseJSONResponse[parsedJob](resp.Content)
	if err != nil {
		return nil, fmt.Errorf("pipeline: parse-job: %w", err)
	}

	if parsed.Title == "" {
		return nil, fmt.Errorf("pipeline: parse-job: LLM returned empty title")
	}
	if parsed.Company == "" {
		return nil, fmt.Errorf("pipeline: parse-job: LLM returned empty company")
	}

	dateParsed := time.Now().Format("2006-01-02")

	// Build the body from parsed sections
	body := "## Job Description (Raw)\n\n"
	if parsed.RawText != "" {
		body += parsed.RawText + "\n"
	} else {
		body += input.PostingText + "\n"
	}
	body += "\n## Parsed Summary\n\n" + parsed.Summary + "\n"
	if len(parsed.KeyRequirements.MustHave) > 0 {
		body += "\n## Key Requirements\n\n### Must Have\n"
		for _, r := range parsed.KeyRequirements.MustHave {
			body += "- " + r + "\n"
		}
	}
	if len(parsed.KeyRequirements.NiceToHave) > 0 {
		body += "\n### Nice to Have\n"
		for _, r := range parsed.KeyRequirements.NiceToHave {
			body += "- " + r + "\n"
		}
	}

	jf := &vault.JobFile{
		Frontmatter: vault.JobFrontmatter{
			Title:           parsed.Title,
			Company:         parsed.Company,
			CompanySlug:     parsed.CompanySlug,
			SourceURL:       input.SourceURL,
			DateParsed:      dateParsed,
			Status:          "targeting",
			Seniority:       parsed.Seniority,
			Domain:          parsed.Domain,
			RequiredSkills:  parsed.RequiredSkills,
			PreferredSkills: parsed.PreferredSkills,
			CultureSignals:  parsed.CultureSignals,
			Compensation:    parsed.Compensation,
			Location:        parsed.Location,
		},
		Body: body,
	}

	return jf, nil
}
