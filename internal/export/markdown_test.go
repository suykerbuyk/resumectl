package export

import (
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssembleMarkdown_Full(t *testing.T) {
	resume := &vault.ResumeFile{
		Body: "## Summary\n\nExperienced engineer.\n\n## Experience\n\n### Role at Company\n- Did things.\n",
	}
	contact := &vault.ContactInfo{
		Name:     "John Smith",
		Email:    "john@example.com",
		Phone:    "+1 555 0000",
		Location: "Loveland, CO",
		LinkedIn: "https://linkedin.com/in/john",
		GitHub:   "https://github.com/john",
		Website:  "https://john.dev",
	}

	result, err := AssembleMarkdown(resume, contact)
	require.NoError(t, err)

	assert.Contains(t, result, "# John Smith")
	assert.Contains(t, result, "john@example.com")
	assert.Contains(t, result, "+1 555 0000")
	assert.Contains(t, result, "Loveland, CO")
	assert.Contains(t, result, "linkedin.com/in/john")
	assert.Contains(t, result, "github.com/john")
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "## Summary")
	assert.Contains(t, result, "Experienced engineer")
}

func TestAssembleMarkdown_NoContact(t *testing.T) {
	resume := &vault.ResumeFile{
		Body: "## Summary\n\nContent here.\n",
	}

	result, err := AssembleMarkdown(resume, nil)
	require.NoError(t, err)

	assert.Contains(t, result, "## Summary")
	// No h1 name header — only h2+
	assert.NotContains(t, result, "\n# ")
}

func TestAssembleMarkdown_EmptyBody(t *testing.T) {
	resume := &vault.ResumeFile{Body: ""}
	_, err := AssembleMarkdown(resume, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestAssembleMarkdown_WhitespaceBody(t *testing.T) {
	resume := &vault.ResumeFile{Body: "   \n  \n  "}
	_, err := AssembleMarkdown(resume, nil)
	require.Error(t, err)
}

func TestNormalizeHeadings_DemoteH1(t *testing.T) {
	input := "# John Smith\n\n## Summary\n\nContent.\n"
	result := normalizeHeadings(input)
	assert.Contains(t, result, "## John Smith")
	assert.Contains(t, result, "## Summary")
}

func TestNormalizeHeadings_PreserveH2(t *testing.T) {
	input := "## Summary\n\n## Experience\n"
	result := normalizeHeadings(input)
	assert.Equal(t, input, result)
}

func TestNormalizeHeadings_NoHeadings(t *testing.T) {
	input := "Just plain text.\n"
	result := normalizeHeadings(input)
	assert.Equal(t, input, result)
}

func TestAssembleMarkdown_TrailingNewline(t *testing.T) {
	resume := &vault.ResumeFile{Body: "Content without trailing newline"}
	result, err := AssembleMarkdown(resume, nil)
	require.NoError(t, err)
	assert.True(t, result[len(result)-1] == '\n')
}

func TestAssembleMarkdown_ContactPartial(t *testing.T) {
	resume := &vault.ResumeFile{Body: "## Summary\n\nTest.\n"}
	contact := &vault.ContactInfo{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		// No phone, location, links
	}

	result, err := AssembleMarkdown(resume, contact)
	require.NoError(t, err)
	assert.Contains(t, result, "# Jane Doe")
	assert.Contains(t, result, "jane@example.com")
}
