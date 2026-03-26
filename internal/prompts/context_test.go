package prompts

import (
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/stretchr/testify/assert"
)

func TestDecomposeContext_Fields(t *testing.T) {
	ctx := DecomposeContext{SourceText: "resume text"}
	assert.Equal(t, "resume text", ctx.SourceText)
}

func TestExtractJobContext_Fields(t *testing.T) {
	ctx := ExtractJobContext{PostingText: "job posting"}
	assert.Equal(t, "job posting", ctx.PostingText)
}

func TestGapAnalysisContext_NilExperience(t *testing.T) {
	// Ensure nil experience slice doesn't cause issues when used in templates
	ctx := GapAnalysisContext{
		JobContent:      "test",
		ExperienceFiles: nil,
		SkillsInventory: "none",
	}
	assert.Nil(t, ctx.ExperienceFiles)

	// Render to confirm no panic
	lib, err := NewLibrary()
	assert.NoError(t, err)
	_, err = lib.Render("gap_analysis", ctx)
	assert.NoError(t, err)
}

func TestSynthesizeContext_WithExperience(t *testing.T) {
	ctx := SynthesizeContext{
		ProfileSummary: "Senior engineer.",
		JobContent:     "Staff role.",
		ExperienceFiles: []*vault.ExperienceFile{
			{
				Frontmatter: vault.ExperienceFrontmatter{
					Role:    "Architect",
					Company: "TestCo",
				},
				Body: "Built systems.\n",
			},
		},
		SkillsInventory: "Go: Expert",
		ContactInfo:     "contact info",
	}

	assert.Len(t, ctx.ExperienceFiles, 1)
	assert.Equal(t, "Architect", ctx.ExperienceFiles[0].Frontmatter.Role)
}

func TestSynthesizeContext_EmptyExperience(t *testing.T) {
	ctx := SynthesizeContext{
		ProfileSummary:  "test",
		JobContent:      "test",
		ExperienceFiles: []*vault.ExperienceFile{},
	}

	lib, err := NewLibrary()
	assert.NoError(t, err)
	_, err = lib.Render("synthesize_resume", ctx)
	assert.NoError(t, err)
}

func TestCoachContext_Fields(t *testing.T) {
	ctx := CoachContext{
		Skill:           "Rust",
		Required:        true,
		GapType:         "hard",
		JobTitle:        "Staff Engineer",
		JobCompany:      "Acme",
		ExistingContext: "Has C++ experience.",
	}
	assert.Equal(t, "Rust", ctx.Skill)
	assert.True(t, ctx.Required)
	assert.Equal(t, "hard", ctx.GapType)
}
