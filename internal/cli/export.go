package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jsuykerbuyk/resumectl/internal/export"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var (
		format       string
		output       string
		templatePath string
	)

	cmd := &cobra.Command{
		Use:   "export <resume-file>",
		Short: "Render a resume to DOCX or PDF",
		Long: `Render a generated resume markdown file to DOCX or PDF.

ASSEMBLY

Before rendering, the resume markdown is assembled into a canonical format:

  1. Contact information from profile/contact.md is injected as a header
     block (name as h1, contact details and links below).
  2. Heading levels in the resume body are normalized: any h1 headings are
     demoted to h2 (h1 is reserved for the candidate's name).
  3. A trailing newline is ensured.

RENDERING

The assembled markdown is rendered via pandoc (must be installed and on PATH).

  DOCX    pandoc -f markdown -t docx [--reference-doc template] -o output
  PDF     pandoc -f markdown -t pdf -o output (requires LaTeX or wkhtmltopdf)

TEMPLATES

Use --template to specify a Word reference document (.docx) that defines
fonts, margins, heading styles, and page layout. Pandoc applies these styles
to the generated document. Create a reference document by exporting a styled
Word file and removing its content.

OUTPUT

If --output is not specified, the output file is named after the input file
with the format extension (e.g. resume.md becomes resume.docx). If
--dry-run is set, assembly is performed but rendering is skipped.

PREREQUISITES

  pandoc    Required. Install via your package manager.
  LaTeX     Required for PDF output. Install texlive or equivalent.`,
		Example: `  # Export to DOCX (default format)
  resumectl export resumes/generated/2026-03-25-acme-staff-engineer.md

  # Export to PDF
  resumectl export --format pdf resumes/generated/2026-03-25-acme.md

  # Export with a custom Word template
  resumectl export --template ~/.resumectl/templates/modern.docx resume.md

  # Export to a specific output path
  resumectl export --output ~/Desktop/resume.docx resume.md

  # Preview assembly without rendering
  resumectl --dry-run export resume.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initDeps(); err != nil {
				return err
			}

			resumePath := args[0]

			// Load resume and contact for markdown assembly
			resume, err := deps.Vault.LoadResume(resumePath)
			if err != nil {
				return err
			}

			contact, _ := deps.Vault.LoadContact() // ok if missing

			assembled, err := export.AssembleMarkdown(resume, contact)
			if err != nil {
				return err
			}

			// Write assembled markdown to temp file
			tmpDir, err := os.MkdirTemp("", "resumectl-export-*")
			if err != nil {
				return fmt.Errorf("create temp dir: %w", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			mdPath := filepath.Join(tmpDir, "resume.md")
			if err := os.WriteFile(mdPath, []byte(assembled), 0o644); err != nil {
				return fmt.Errorf("write temp markdown: %w", err)
			}

			fmt, err := export.ParseFormat(format)
			if err != nil {
				return err
			}

			if output == "" {
				base := strings.TrimSuffix(filepath.Base(resumePath), ".md")
				output = base + "." + string(fmt)
			}

			if dryRun {
				_, _ = cmd.OutOrStdout().Write([]byte("[dry-run] Would export to " + output + "\n"))
				return nil
			}

			if err := export.Export(mdPath, output, fmt, templatePath); err != nil {
				return err
			}

			_, _ = cmd.OutOrStdout().Write([]byte("✓ Exported to " + output + "\n"))
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "docx", "output format (docx|pdf)")
	cmd.Flags().StringVar(&output, "output", "", "output file path")
	cmd.Flags().StringVar(&templatePath, "template", "", "DOCX reference template path")

	return cmd
}
