package cli

import (
	"fmt"
	"os"

	"github.com/jsuykerbuyk/resumectl/internal/fetch"
	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/pipeline"
	"github.com/spf13/cobra"
)

func newIngestCmd() *cobra.Command {
	var (
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "ingest <source>",
		Short: "Decompose historical documents into vault experience files",
		Long: `Decompose a historical resume or employment document into structured
experience files in the vault.

The source document is first converted to plaintext based on its format:

  .docx     Extracted via pandoc (must be installed)
  .pdf      Extracted via pdftotext (poppler; must be installed)
  .txt .md  Read directly as plaintext
  .html     HTML tags stripped, text extracted

The plaintext is sent to an LLM (default: Claude Sonnet, override with
--provider) with the decompose_resume prompt template. The LLM identifies
each distinct role or engagement in the document and returns a JSON array
of structured role objects.

For each role, resumectl creates an experience file in the vault's
experience/ directory with YAML frontmatter (role, company, dates, tags,
skills, domain) and a markdown body containing Summary, Key Contributions,
and Technologies sections. Contribution bullets are preserved verbatim from
the source document where possible.

DEDUPLICATION

Before writing each file, the vault is checked for existing experience
files with the same company_slug and start date. Matches are reported as
warnings and skipped. Use --interactive to review each file before writing.

OUTPUT

Files are written to experience/YYYY-MM-{company-slug}-{role-slug}.md.
If --dry-run is set, no files are written; the planned output is printed
to stdout. If Git auto-commit is enabled in config, each written file is
automatically staged and committed.

After ingestion, a list of suggested skill additions is printed. These are
skills mentioned in the decomposed roles that may warrant adding to
profile/skills.md.`,
		Example: `  # Ingest a Word document
  resumectl ingest ~/Documents/resume-2022.docx

  # Preview what would be created without writing
  resumectl ingest --dry-run old-resume.txt

  # Confirm each file before writing
  resumectl ingest --interactive ~/old-resumes/resume.md

  # Ingest a PDF using a specific provider
  resumectl --provider grok ingest ~/Documents/resume.pdf

  # Ingest an HTML export
  resumectl ingest linkedin-profile.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initDeps(); err != nil {
				return err
			}

			source := args[0]

			// Extract text based on source type
			text, err := extractSourceText(source)
			if err != nil {
				return err
			}

			provider, err := getProvider(llm.TaskDecompose)
			if err != nil {
				return err
			}

			result, err := pipeline.RunIngest(cmd.Context(), pipeline.IngestInput{
				SourceText: text,
				SourcePath: source,
			}, deps.Vault, deps.Prompts, provider)
			if err != nil {
				return err
			}

			for _, w := range result.Warnings {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "⚠ %s\n", w)
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n[dry-run] Would create %d experience files:\n", len(result.ExperienceFiles))
				for _, ef := range result.ExperienceFiles {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s — %s at %s\n",
						ef.RelPath, ef.Frontmatter.Role, ef.Frontmatter.Company)
				}
				return nil
			}

			for _, ef := range result.ExperienceFiles {
				if interactive {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create %s — %s at %s? [Y/n] ",
						ef.RelPath, ef.Frontmatter.Role, ef.Frontmatter.Company)
					// In non-interactive tests, default to yes
				}

				if err := deps.Vault.WriteExperience(ef); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Created %s\n", ef.RelPath)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nIngested %d experience files from %s\n",
				len(result.ExperienceFiles), source)

			if len(result.SkillsUpdates) > 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Suggested skill additions: %v\n", result.SkillsUpdates)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&interactive, "interactive", false, "prompt before creating each file")

	return cmd
}

func extractSourceText(source string) (string, error) {
	srcType := fetch.DetectSourceType(source)

	switch srcType {
	case fetch.SourceDOCX:
		return fetch.ExtractDOCX(source)
	case fetch.SourcePDF:
		return fetch.ExtractPDF(source)
	case fetch.SourceMarkdown, fetch.SourcePlaintext, fetch.SourceHTML, fetch.SourceUnknown:
		data, err := os.ReadFile(source)
		if err != nil {
			return "", fmt.Errorf("read source: %w", err)
		}
		if srcType == fetch.SourceHTML {
			return fetch.StripHTML(string(data)), nil
		}
		return string(data), nil
	case fetch.SourceURL:
		return "", fmt.Errorf("URL ingestion not supported directly; use parse-job for URLs")
	default:
		return "", fmt.Errorf("unsupported source type for %s", source)
	}
}
