package fetch

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// SourceType identifies the kind of input source.
type SourceType int

const (
	SourceUnknown SourceType = iota
	SourceMarkdown
	SourcePlaintext
	SourceDOCX
	SourcePDF
	SourceHTML
	SourceURL
)

// DetectSourceType determines the source type from a path or URL string.
func DetectSourceType(source string) SourceType {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return SourceURL
	}

	ext := strings.ToLower(filepath.Ext(source))
	switch ext {
	case ".md", ".markdown":
		return SourceMarkdown
	case ".txt":
		return SourcePlaintext
	case ".docx":
		return SourceDOCX
	case ".pdf":
		return SourcePDF
	case ".html", ".htm":
		return SourceHTML
	default:
		return SourceUnknown
	}
}

// ExtractDOCX extracts plaintext from a .docx file using pandoc.
func ExtractDOCX(path string) (string, error) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		return "", fmt.Errorf("fetch: pandoc not found (required for DOCX extraction): %w", err)
	}

	out, err := exec.Command("pandoc", "-f", "docx", "-t", "plain", path).Output()
	if err != nil {
		return "", fmt.Errorf("fetch: extract DOCX %s: %w", path, err)
	}
	return string(out), nil
}

// ExtractPDF extracts plaintext from a PDF file using pdftotext.
func ExtractPDF(path string) (string, error) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", fmt.Errorf("fetch: pdftotext not found (required for PDF extraction): %w", err)
	}

	out, err := exec.Command("pdftotext", "-layout", path, "-").Output()
	if err != nil {
		return "", fmt.Errorf("fetch: extract PDF %s: %w", path, err)
	}
	return string(out), nil
}

// StripHTML does basic HTML tag removal. For production use, a proper
// readability library would be preferred.
func StripHTML(html string) string {
	var result strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			result.WriteRune(' ')
		case !inTag:
			result.WriteRune(r)
		}
	}
	// Collapse whitespace
	text := result.String()
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}
