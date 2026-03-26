package prompts

import (
	"strings"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLibrary(t *testing.T) *Library {
	t.Helper()
	lib, err := NewLibrary()
	require.NoError(t, err)
	return lib
}

func TestRender_DecomposeResume(t *testing.T) {
	lib := testLibrary(t)
	ctx := DecomposeContext{SourceText: "I worked at Acme Corp as a Senior Engineer from 2020 to 2024."}

	result, err := lib.Render("decompose_resume", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "Acme Corp")
	assert.Contains(t, result, "JSON array")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_ExtractJob(t *testing.T) {
	lib := testLibrary(t)
	ctx := ExtractJobContext{PostingText: "We are hiring a Staff Engineer for our infrastructure team."}

	result, err := lib.Render("extract_job", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "Staff Engineer")
	assert.Contains(t, result, "infrastructure")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_GapAnalysis(t *testing.T) {
	lib := testLibrary(t)
	ctx := GapAnalysisContext{
		JobContent: "Staff Engineer role requiring Go and Kubernetes.",
		ExperienceFiles: []*vault.ExperienceFile{
			{
				RelPath: "experience/test.md",
				Body:    "## Summary\n\nBuilt Go services.\n",
			},
		},
		SkillsInventory: "Go: Advanced, Kubernetes: Intermediate",
	}

	result, err := lib.Render("gap_analysis", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "experience/test.md")
	assert.Contains(t, result, "Go services")
	assert.Contains(t, result, "Go: Advanced")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_GapAnalysis_EmptyExperience(t *testing.T) {
	lib := testLibrary(t)
	ctx := GapAnalysisContext{
		JobContent:      "Some job",
		ExperienceFiles: nil, // empty slice
		SkillsInventory: "None",
	}

	result, err := lib.Render("gap_analysis", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "Some job")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_SynthesizeResume(t *testing.T) {
	lib := testLibrary(t)
	ctx := SynthesizeContext{
		ProfileSummary: "Seasoned systems engineer.",
		JobContent:     "Staff Engineer at Acme Corp.",
		ExperienceFiles: []*vault.ExperienceFile{
			{
				Frontmatter: vault.ExperienceFrontmatter{
					Role:    "Principal Architect",
					Company: "BigCo",
				},
				Body: "## Summary\n\nLed architecture.\n",
			},
		},
		SkillsInventory: "Go, Kubernetes, NVMe",
		ContactInfo:     "John Smith | john@example.com",
	}

	result, err := lib.Render("synthesize_resume", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "Principal Architect")
	assert.Contains(t, result, "BigCo")
	assert.Contains(t, result, "John Smith")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_SynthesizeCoverLetter(t *testing.T) {
	lib := testLibrary(t)
	ctx := SynthesizeContext{
		ProfileSummary: "Seasoned engineer.",
		JobContent:     "Staff Engineer role.",
		ExperienceFiles: []*vault.ExperienceFile{
			{
				Frontmatter: vault.ExperienceFrontmatter{
					Role:    "Engineer",
					Company: "TestCo",
				},
				Body: "Did things.\n",
			},
		},
		ContactInfo: "Jane Doe",
	}

	result, err := lib.Render("synthesize_cover_letter", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "TestCo")
	assert.Contains(t, result, "Jane Doe")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_CoachQuestion(t *testing.T) {
	lib := testLibrary(t)
	ctx := CoachContext{
		Skill:           "Rust",
		Required:        true,
		GapType:         "hard",
		JobTitle:        "Staff Engineer",
		JobCompany:      "Acme Corp",
		ExistingContext: "Strong C++ background documented.",
	}

	result, err := lib.Render("coach_question", ctx)
	require.NoError(t, err)
	assert.Contains(t, result, "Rust")
	assert.Contains(t, result, "Acme Corp")
	assert.Contains(t, result, "C++ background")
	assert.NotContains(t, result, "<no value>")
}

func TestRender_UnknownTemplate(t *testing.T) {
	lib := testLibrary(t)
	_, err := lib.Render("nonexistent", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

func TestRender_NoNoValueMarkers(t *testing.T) {
	lib := testLibrary(t)

	// Render all templates with their proper context types
	contexts := map[string]any{
		"decompose_resume":        DecomposeContext{SourceText: "test text"},
		"extract_job":             ExtractJobContext{PostingText: "test posting"},
		"gap_analysis":            GapAnalysisContext{JobContent: "test job", SkillsInventory: "none"},
		"synthesize_resume":       SynthesizeContext{ProfileSummary: "test", JobContent: "test"},
		"synthesize_cover_letter": SynthesizeContext{ProfileSummary: "test", JobContent: "test"},
		"coach_question":          CoachContext{Skill: "Go", JobTitle: "Eng", JobCompany: "Co"},
	}

	for name, ctx := range contexts {
		t.Run(name, func(t *testing.T) {
			result, err := lib.Render(name, ctx)
			require.NoError(t, err)
			assert.NotContains(t, result, "<no value>",
				"template %q should not produce <no value> markers", name)
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		minToken int
		maxToken int
	}{
		{"", 0, 0},
		{"hello", 1, 3},
		{strings.Repeat("a", 400), 90, 110},
		{strings.Repeat("word ", 100), 100, 150},
	}
	for _, tt := range tests {
		tokens := EstimateTokens(tt.input)
		assert.GreaterOrEqual(t, tokens, tt.minToken)
		assert.LessOrEqual(t, tokens, tt.maxToken)
	}
}

func TestRender_TokenBudgets(t *testing.T) {
	lib := testLibrary(t)

	// Decompose with typical input (~1500 chars) should produce reasonable output
	longText := strings.Repeat("Senior Engineer at Acme doing distributed systems work. ", 30)
	result, err := lib.Render("decompose_resume", DecomposeContext{SourceText: longText})
	require.NoError(t, err)
	tokens := EstimateTokens(result)
	// Template instructions + input should be in a reasonable range
	assert.Greater(t, tokens, 100, "decompose prompt should be substantial")
	assert.Less(t, tokens, 10000, "decompose prompt should not be enormous")
}
