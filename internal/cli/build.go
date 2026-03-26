package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/pipeline"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var (
		coverLetter   bool
		maxExperience int
		output        string
	)

	cmd := &cobra.Command{
		Use:   "build <job-file>",
		Short: "Synthesize a targeted resume for a specific job",
		Long: `Synthesize a targeted resume for a specific job posting.

BUILD PROCESS

  1. The job file is loaded and gap analysis (resumectl match) is run
     automatically to identify the most relevant experience files.
  2. Experience files referenced in strong and partial matches are loaded.
     If no matches are found, all experience files are used as fallback.
  3. Profile data is assembled: contact info from profile/contact.md,
     professional identity from profile/summary-core.md, and skills from
     profile/skills.md.
  4. All context is rendered into the synthesize_resume prompt template
     and sent to an LLM (default: Claude Opus for highest writing quality;
     the synthesis model is configurable in ~/.config/resumectl/config.yaml).
  5. The LLM generates a complete, ATS-compatible resume in markdown with
     a tailored summary, experience sections ordered by relevance, and
     rewritten contribution bullets with quantified impact.

OUTPUT VALIDATION

The generated resume is checked for:
  - Presence of a summary/profile section
  - Word count within the 600-800 word target (warnings if outside)
  - No hallucinated company names (cross-checked against input files)

COVER LETTER

With --cover-letter, a second LLM call generates a cover letter using the
same context. The cover letter is printed to stdout (not written to a
separate file).

OUTPUT

The resume is written to resumes/generated/YYYY-MM-DD-{company-slug}.md
with YAML frontmatter tracking the source job file, LLM model used,
experience files included, and version number. Override the path with
--output. Use --dry-run to preview without writing.`,
		Example: `  # Generate a resume for a parsed job posting
  resumectl build jobs/target/2026-03-25-acme-staff-engineer.md

  # Also generate a cover letter
  resumectl build --cover-letter jobs/target/2026-03-25-acme.md

  # Preview the generated resume without writing
  resumectl --dry-run build jobs/target/2026-03-25-acme.md

  # Write to a custom output path
  resumectl build --output my-resume.md jobs/target/2026-03-25-acme.md

  # Limit to top 3 most relevant experience files
  resumectl build --max-experience 3 jobs/target/2026-03-25-acme.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initDeps(); err != nil {
				return err
			}

			job, err := deps.Vault.LoadJob(args[0])
			if err != nil {
				return err
			}

			// Select experience files via match
			provider, err := getProvider(llm.TaskGapAnalysis)
			if err != nil {
				return err
			}

			matchResult, err := pipeline.RunMatch(cmd.Context(), job, deps.Vault, deps.Prompts, provider, maxExperience)
			if err != nil {
				return fmt.Errorf("match for file selection: %w", err)
			}

			// Collect experience files referenced in strong + partial matches
			expPaths := make(map[string]bool)
			for _, m := range matchResult.StrongMatches {
				for _, ref := range m.ExperienceRefs {
					expPaths[ref] = true
				}
			}
			for _, m := range matchResult.PartialMatches {
				for _, ref := range m.ExperienceRefs {
					expPaths[ref] = true
				}
			}

			var experienceFiles []*vault.ExperienceFile
			for path := range expPaths {
				ef, err := deps.Vault.LoadExperience(path)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "⚠ skipping %s: %v\n", path, err)
					continue
				}
				experienceFiles = append(experienceFiles, ef)
			}

			// Fallback: if no matches found, use all experience
			if len(experienceFiles) == 0 {
				experienceFiles, err = deps.Vault.AllExperience()
				if err != nil {
					return err
				}
			}

			synthProvider, err := getProvider(llm.TaskSynthesizeResume)
			if err != nil {
				return err
			}

			result, err := pipeline.RunBuild(cmd.Context(), pipeline.BuildInput{
				Job:             job,
				ExperienceFiles: experienceFiles,
				CoverLetter:     coverLetter,
			}, deps.Vault, deps.Prompts, synthProvider, cfg.Providers.Claude.SynthesisModel)
			if err != nil {
				return err
			}

			for _, w := range result.Warnings {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "⚠ %s\n", w)
			}

			if output == "" {
				datestamp := time.Now().Format("2006-01-02")
				output = filepath.Join("resumes", "generated",
					fmt.Sprintf("%s-%s.md", datestamp, job.Frontmatter.CompanySlug))
			}
			result.Resume.RelPath = output
			result.Resume.Frontmatter.Generated = time.Now().Format("2006-01-02")

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would write resume to %s\n", output)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Resume.Body)
				return nil
			}

			if err := deps.Vault.WriteResume(result.Resume); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Created %s\n", result.Resume.RelPath)

			if result.CoverLetter != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n--- Cover Letter ---\n%s\n", result.CoverLetter)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&coverLetter, "cover-letter", false, "also generate a cover letter")
	cmd.Flags().IntVar(&maxExperience, "max-experience", 5, "max experience files to include")
	cmd.Flags().StringVar(&output, "output", "", "output path (default: resumes/generated/<date>-<slug>.md)")

	return cmd
}
