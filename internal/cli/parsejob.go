package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jsuykerbuyk/resumectl/internal/fetch"
	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/pipeline"
	"github.com/spf13/cobra"
)

func newParseJobCmd() *cobra.Command {
	var (
		slug   string
		openEd bool
	)

	cmd := &cobra.Command{
		Use:   "parse-job <source>",
		Short: "Parse a job posting into a structured vault file",
		Long: `Fetch a job posting and extract structured data into a vault job file.

The source argument can be:

  URL       An HTTP(S) URL to a job posting page. The page is fetched and
            HTML is stripped to extract the posting text. Works with most
            ATS platforms (Lever, Greenhouse, Ashby). LinkedIn and other
            JavaScript-rendered pages may return incomplete content.
  File      A local file (.txt, .md, .html) containing the posting text.
  Stdin     Use "-" to read from stdin (e.g. pbpaste | resumectl parse-job -).

The extracted text is sent to an LLM (default: Gemini Flash, override with
--provider) with the extract_job prompt template. The LLM returns a JSON
object with structured fields:

  title, company, company_slug, location, seniority, domain,
  required_skills, preferred_skills, culture_signals, compensation,
  summary, key_requirements (must_have + nice_to_have), raw_text

The parsed data is written as a markdown file with YAML frontmatter to
jobs/target/YYYY-MM-DD-{company-slug}-{title-slug}.md. The body includes
the raw job description, parsed summary, and key requirements sections.

The initial status is set to "targeting". Use --slug to override the
auto-generated filename slug. Use --open to open the file in $EDITOR after
creation for manual review and annotation.

If --dry-run is set, the extracted data is printed but no file is written.`,
		Example: `  # Parse a job posting from a URL
  resumectl parse-job https://www.linkedin.com/jobs/view/1234567890

  # Parse from a different ATS
  resumectl parse-job https://jobs.lever.co/acme/some-uuid

  # Parse from a local text file
  resumectl parse-job job-description.txt

  # Parse and immediately open in your editor
  resumectl parse-job --open job-description.txt

  # Parse from clipboard (macOS)
  pbpaste | resumectl parse-job -

  # Parse from clipboard (Linux/X11)
  xclip -selection clipboard -o | resumectl parse-job -

  # Use Grok instead of the default provider
  resumectl --provider grok parse-job posting.txt

  # Preview extraction without writing
  resumectl --dry-run parse-job posting.txt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initDeps(); err != nil {
				return err
			}

			source := args[0]
			text, sourceURL, err := fetchJobText(cmd.Context(), source)
			if err != nil {
				return err
			}

			provider, err := getProvider(llm.TaskExtractJob)
			if err != nil {
				return err
			}

			jf, err := pipeline.RunParseJob(cmd.Context(), pipeline.ParseJobInput{
				PostingText: text,
				SourceURL:   sourceURL,
				Slug:        slug,
			}, deps.Prompts, provider)
			if err != nil {
				return err
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would create job file:\n")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Title:    %s\n", jf.Frontmatter.Title)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Company:  %s\n", jf.Frontmatter.Company)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Skills:   %v\n", jf.Frontmatter.RequiredSkills)
				return nil
			}

			if err := deps.Vault.WriteJob(jf); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Created %s\n", jf.RelPath)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s at %s (%s)\n",
				jf.Frontmatter.Title, jf.Frontmatter.Company, jf.Frontmatter.Status)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Required: %v\n", jf.Frontmatter.RequiredSkills)

			if openEd {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vi"
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Open with: %s %s\n", editor, jf.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "override generated filename slug")
	cmd.Flags().BoolVar(&openEd, "open", false, "open created file in $EDITOR")

	return cmd
}

func fetchJobText(ctx context.Context, source string) (text, sourceURL string, err error) {
	if source == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), "", nil
	}

	srcType := fetch.DetectSourceType(source)
	if srcType == fetch.SourceURL {
		f := fetch.NewFetcher(15*time.Second, cfg.Fetch.UserAgent)
		html, err := f.Fetch(ctx, source)
		if err != nil {
			return "", "", err
		}
		return fetch.StripHTML(html), source, nil
	}

	// File source
	data, err := os.ReadFile(source)
	if err != nil {
		return "", "", fmt.Errorf("read source: %w", err)
	}
	return string(data), "", nil
}
