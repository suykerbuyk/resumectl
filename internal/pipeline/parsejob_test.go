package pipeline

import (
	"context"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunParseJob_Success(t *testing.T) {
	provider := mockProvider(t, "parsejob_response.json")
	lib := testLibrary(t)

	jf, err := RunParseJob(context.Background(), ParseJobInput{
		PostingText: "We are hiring a Staff Engineer at Acme Corp.",
		SourceURL:   "https://example.com/jobs/123",
	}, lib, provider)

	require.NoError(t, err)
	assert.Equal(t, "Staff Software Engineer — Infrastructure", jf.Frontmatter.Title)
	assert.Equal(t, "Acme Corp", jf.Frontmatter.Company)
	assert.Equal(t, "acme-corp", jf.Frontmatter.CompanySlug)
	assert.Equal(t, "targeting", jf.Frontmatter.Status)
	assert.Equal(t, "staff", jf.Frontmatter.Seniority)
	assert.Contains(t, jf.Frontmatter.RequiredSkills, "Go")
	assert.Contains(t, jf.Frontmatter.PreferredSkills, "Rust")
	require.NotNil(t, jf.Frontmatter.Compensation)
	assert.Equal(t, 200000, jf.Frontmatter.Compensation.RangeLow)
	assert.Contains(t, jf.Body, "## Job Description")
	assert.Contains(t, jf.Body, "## Parsed Summary")
	assert.Equal(t, 1, provider.RequestCount())
}

func TestRunParseJob_MalformedJSON(t *testing.T) {
	provider := llm.NewMockProvider("not valid json")
	lib := testLibrary(t)

	_, err := RunParseJob(context.Background(), ParseJobInput{
		PostingText: "Some job posting.",
	}, lib, provider)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse JSON")
}

func TestRunParseJob_EmptyTitle(t *testing.T) {
	provider := llm.NewMockProvider(`{"title": "", "company": "Acme"}`)
	lib := testLibrary(t)

	_, err := RunParseJob(context.Background(), ParseJobInput{
		PostingText: "Some job posting.",
	}, lib, provider)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty title")
}

func TestRunParseJob_EmptyCompany(t *testing.T) {
	provider := llm.NewMockProvider(`{"title": "Engineer", "company": ""}`)
	lib := testLibrary(t)

	_, err := RunParseJob(context.Background(), ParseJobInput{
		PostingText: "Some job posting.",
	}, lib, provider)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty company")
}

func TestRunParseJob_CodeFencedResponse(t *testing.T) {
	fixture := loadFixture(t, "parsejob_response.json")
	wrapped := "```json\n" + fixture + "\n```"
	provider := llm.NewMockProvider(wrapped)
	lib := testLibrary(t)

	jf, err := RunParseJob(context.Background(), ParseJobInput{
		PostingText: "Hiring a Staff Engineer.",
	}, lib, provider)

	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", jf.Frontmatter.Company)
}
