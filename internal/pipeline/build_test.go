package pipeline

import (
	"context"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunBuild_Success(t *testing.T) {
	v := openTestVault(t)
	provider := mockProvider(t, "synthesize_resume_response.md")
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	experience, err := v.AllExperience()
	require.NoError(t, err)

	result, err := RunBuild(context.Background(), BuildInput{
		Job:             job,
		ExperienceFiles: experience,
	}, v, lib, provider, "")

	require.NoError(t, err)
	assert.NotNil(t, result.Resume)
	assert.Contains(t, result.Resume.Body, "John Smith")
	assert.Contains(t, result.Resume.Body, "Summary")
	assert.Equal(t, "draft", result.Resume.Frontmatter.Status)
	assert.Equal(t, 1, result.Resume.Frontmatter.Version)
	assert.Len(t, result.Resume.Frontmatter.ExperienceFiles, 3)
	assert.Empty(t, result.CoverLetter)
}

func TestRunBuild_WithCoverLetter(t *testing.T) {
	v := openTestVault(t)
	resumeContent := loadFixture(t, "synthesize_resume_response.md")
	provider := llm.NewMockProvider(resumeContent, "Dear Hiring Manager, I am excited...")
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	experience, err := v.AllExperience()
	require.NoError(t, err)

	result, err := RunBuild(context.Background(), BuildInput{
		Job:             job,
		ExperienceFiles: experience,
		CoverLetter:     true,
	}, v, lib, provider, "")

	require.NoError(t, err)
	assert.NotEmpty(t, result.CoverLetter)
	assert.Contains(t, result.CoverLetter, "Hiring Manager")
	assert.Equal(t, 2, provider.RequestCount()) // resume + cover letter
}

func TestRunBuild_EmptyExperience(t *testing.T) {
	v := openTestVault(t)
	provider := llm.NewMockProvider("unused")
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	_, err = RunBuild(context.Background(), BuildInput{
		Job:             job,
		ExperienceFiles: nil,
	}, v, lib, provider, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no experience files")
}

func TestRunBuild_ValidationWarnings(t *testing.T) {
	v := openTestVault(t)
	// Very short resume response
	provider := llm.NewMockProvider("This is way too short.")
	lib := testLibrary(t)

	job, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	experience, err := v.AllExperience()
	require.NoError(t, err)

	result, err := RunBuild(context.Background(), BuildInput{
		Job:             job,
		ExperienceFiles: experience,
	}, v, lib, provider, "")

	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
}
