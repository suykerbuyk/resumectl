package export

import (
	"fmt"
	"strings"
)

// Format represents an export output format.
type Format string

const (
	FormatDOCX Format = "docx"
	FormatPDF  Format = "pdf"
)

// ParseFormat converts a string to a Format, returning an error for unknown formats.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "docx":
		return FormatDOCX, nil
	case "pdf":
		return FormatPDF, nil
	default:
		return "", fmt.Errorf("export: unknown format %q (must be docx or pdf)", s)
	}
}

// Export renders a markdown file to the specified format.
func Export(markdownPath, outputPath string, format Format, templatePath string) error {
	switch format {
	case FormatDOCX:
		return RenderDOCX(markdownPath, outputPath, templatePath)
	case FormatPDF:
		return RenderPDF(markdownPath, outputPath)
	default:
		return fmt.Errorf("export: unsupported format %q", format)
	}
}
