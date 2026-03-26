package pipeline

import (
	"context"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunIngest_Success(t *testing.T) {
	v := openTestVault(t)
	provider := mockProvider(t, "decompose_resume_response.json")
	lib := testLibrary(t)

	result, err := RunIngest(context.Background(), IngestInput{
		SourceText: "Full resume text with multiple roles.",
		SourcePath: "/path/to/resume.docx",
	}, v, lib, provider)

	require.NoError(t, err)
	// 3 roles in fixture, but some may be deduplicated against test vault
	// Test vault has seagate, intel, cloudscale — all 3 match the fixture
	// So all 3 should be flagged as duplicates
	assert.Len(t, result.Warnings, 3, "all 3 roles should be duplicates of test vault")
	assert.Empty(t, result.ExperienceFiles, "duplicates should be excluded")
}

func TestRunIngest_NoDuplicates(t *testing.T) {
	// Use empty vault to avoid dedup
	v := openTestVault(t)
	// Remove existing experience files to test without dedup
	fixture := `[{
		"role": "New Role",
		"company": "New Corp",
		"company_slug": "new-corp",
		"start": "2025-01",
		"end": "2025-12",
		"current": false,
		"tags": ["testing"],
		"skills": ["Go"],
		"domain": "software",
		"summary": "A new role.",
		"contributions": ["Built things."],
		"technologies": ["Go"]
	}]`
	provider := llm.NewMockProvider(fixture)
	lib := testLibrary(t)

	result, err := RunIngest(context.Background(), IngestInput{
		SourceText: "Resume with one new role.",
	}, v, lib, provider)

	require.NoError(t, err)
	assert.Len(t, result.ExperienceFiles, 1)
	assert.Equal(t, "New Role", result.ExperienceFiles[0].Frontmatter.Role)
	assert.Equal(t, "New Corp", result.ExperienceFiles[0].Frontmatter.Company)
	assert.Contains(t, result.ExperienceFiles[0].Body, "## Summary")
	assert.Contains(t, result.ExperienceFiles[0].Body, "## Key Contributions")
	assert.Contains(t, result.ExperienceFiles[0].Body, "Built things.")
	assert.Contains(t, result.SkillsUpdates, "Go")
	assert.Empty(t, result.Warnings)
}

func TestRunIngest_InvalidJSON(t *testing.T) {
	v := openTestVault(t)
	provider := llm.NewMockProvider("not json")
	lib := testLibrary(t)

	_, err := RunIngest(context.Background(), IngestInput{
		SourceText: "Some text.",
	}, v, lib, provider)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse JSON")
}

func TestRunIngest_SingleRole(t *testing.T) {
	v := openTestVault(t)
	fixture := `[{
		"role": "Solo Engineer",
		"company": "Solo Inc",
		"company_slug": "solo-inc",
		"start": "2024-06",
		"end": "present",
		"current": true,
		"tags": ["solo"],
		"skills": ["Python"],
		"domain": "software",
		"summary": "Solo role.",
		"contributions": ["Did everything."],
		"technologies": ["Python"]
	}]`
	provider := llm.NewMockProvider(fixture)
	lib := testLibrary(t)

	result, err := RunIngest(context.Background(), IngestInput{
		SourceText: "Resume with one role.",
	}, v, lib, provider)

	require.NoError(t, err)
	assert.Len(t, result.ExperienceFiles, 1)
	assert.True(t, result.ExperienceFiles[0].Frontmatter.Current)
}
