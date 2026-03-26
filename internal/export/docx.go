package export

import (
	"fmt"
	"os/exec"
)

// RenderDOCX renders a markdown file to DOCX using pandoc.
// templatePath is optional — if provided, it sets the --reference-doc for styling.
func RenderDOCX(markdownPath, outputPath, templatePath string) error {
	if _, err := exec.LookPath("pandoc"); err != nil {
		return fmt.Errorf("export: pandoc not found (required for DOCX rendering): %w", err)
	}

	args := []string{
		"-f", "markdown",
		"-t", "docx",
		"-o", outputPath,
	}

	if templatePath != "" {
		args = append(args, "--reference-doc", templatePath)
	}

	args = append(args, markdownPath)

	cmd := exec.Command("pandoc", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("export: pandoc DOCX: %w\n%s", err, string(out))
	}

	return nil
}
