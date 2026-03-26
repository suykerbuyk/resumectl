package export

import (
	"fmt"
	"os/exec"
)

// RenderPDF renders a markdown file to PDF using pandoc.
// Requires a PDF engine (pdflatex, xelatex, or wkhtmltopdf).
func RenderPDF(markdownPath, outputPath string) error {
	if _, err := exec.LookPath("pandoc"); err != nil {
		return fmt.Errorf("export: pandoc not found (required for PDF rendering): %w", err)
	}

	cmd := exec.Command("pandoc",
		"-f", "markdown",
		"-t", "pdf",
		"-o", outputPath,
		markdownPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("export: pandoc PDF: %w\n%s", err, string(out))
	}

	return nil
}
