package fetch

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectSourceType(t *testing.T) {
	tests := []struct {
		input    string
		expected SourceType
	}{
		{"https://linkedin.com/jobs/123", SourceURL},
		{"http://example.com/job.html", SourceURL},
		{"resume.md", SourceMarkdown},
		{"resume.markdown", SourceMarkdown},
		{"resume.txt", SourcePlaintext},
		{"resume.docx", SourceDOCX},
		{"resume.pdf", SourcePDF},
		{"page.html", SourceHTML},
		{"page.htm", SourceHTML},
		{"unknown.xyz", SourceUnknown},
		{"noextension", SourceUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectSourceType(tt.input))
		})
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			"simple tags",
			"<html><body><h1>Title</h1><p>Content here.</p></body></html>",
			"Title Content here.",
		},
		{
			"nested tags",
			"<div><span>Hello</span> <strong>World</strong></div>",
			"Hello World",
		},
		{
			"no tags",
			"Plain text without HTML.",
			"Plain text without HTML.",
		},
		{
			"navigation stripped",
			`<nav><a href="/">Home</a><a href="/about">About</a></nav>
			<main><article><h1>Job Posting</h1><p>We need an engineer.</p></article></main>`,
			"Job Posting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripHTML(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestStripHTML_CollapseWhitespace(t *testing.T) {
	result := StripHTML("<p>word1</p>   <p>word2</p>")
	// Should not have excessive whitespace
	assert.NotContains(t, result, "  ")
}

func TestExtractDOCX_NoPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err == nil {
		t.Skip("pandoc is available — this test covers the missing-pandoc path")
	}
	_, err := ExtractDOCX("nonexistent.docx")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pandoc not found")
}

func TestExtractDOCX_WithPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not available")
	}
	// Test with a non-existent file — pandoc will error
	_, err := ExtractDOCX("nonexistent.docx")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extract DOCX")
}

func TestExtractPDF_NoPdftotext(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err == nil {
		t.Skip("pdftotext is available — this test covers the missing-pdftotext path")
	}
	_, err := ExtractPDF("nonexistent.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pdftotext not found")
}

func TestExtractPDF_WithPdftotext(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not available")
	}
	_, err := ExtractPDF("nonexistent.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extract PDF")
}
