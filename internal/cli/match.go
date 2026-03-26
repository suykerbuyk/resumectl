package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/pipeline"
	"github.com/spf13/cobra"
)

func newMatchCmd() *cobra.Command {
	var (
		format string
		write  bool
		topN   int
	)

	cmd := &cobra.Command{
		Use:   "match <job-file>",
		Short: "Perform gap analysis between a job file and the vault",
		Long: `Perform a two-stage gap analysis between a parsed job file and your vault.

STAGE 1: LOCAL TAG INTERSECTION

The vault's tag index is built from all experience files. The job's required
and preferred skills are matched against experience file tags and skills using
case-insensitive string intersection. Each experience file receives a relevance
score (0.0-1.0) based on the fraction of job skills it matches. The top N
files (default: 5, set with --top) are selected as candidates.

STAGE 2: LLM SEMANTIC ANALYSIS

The top candidate experience files (Summary + Key Contributions only, to
manage token budget) along with the job posting and your skills inventory are
sent to an LLM (default: Claude Sonnet, override with --provider) with the
gap_analysis prompt template.

The LLM returns a structured analysis containing:

  Overall Score     0-100 match percentage
  Strong Matches    Skills with high-confidence evidence in your vault
  Partial Matches   Skills with adjacent but incomplete evidence
  Gaps              Skills the job requires but your vault doesn't document

Each gap includes a coaching suggestion — a question to help you recall
experience. Run 'resumectl coach' to address gaps interactively.

OUTPUT FORMATS

  text (default)    Human-readable formatted report with match scores
  json              Machine-readable JSON for scripting or piping

EXIT CODES

  0    Analysis completed successfully
  1    Error (missing job file, LLM failure, etc.)`,
		Example: `  # Run gap analysis with formatted output
  resumectl match jobs/target/2026-03-25-acme-staff-engineer.md

  # Output as JSON for scripting
  resumectl match --format json jobs/target/2026-03-25-acme.md

  # Limit to top 3 experience files
  resumectl match --top 3 jobs/target/2026-03-25-acme.md

  # Pipe JSON output to jq
  resumectl match --format json jobs/target/acme.md | jq '.gaps[]'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initDeps(); err != nil {
				return err
			}

			job, err := deps.Vault.LoadJob(args[0])
			if err != nil {
				return err
			}

			provider, err := getProvider(llm.TaskGapAnalysis)
			if err != nil {
				return err
			}

			result, err := pipeline.RunMatch(cmd.Context(), job, deps.Vault, deps.Prompts, provider, topN)
			if err != nil {
				return err
			}

			switch strings.ToLower(format) {
			case "json":
				data, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			default:
				printMatchResult(cmd, result, job.Frontmatter.Title, job.Frontmatter.Company)
			}

			if write && !dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n(--write: gap analysis write-back not yet implemented)\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "output format (text|json)")
	cmd.Flags().BoolVar(&write, "write", false, "write gap analysis into job file")
	cmd.Flags().IntVar(&topN, "top", 5, "number of top matching experience files")

	return cmd
}

func printMatchResult(cmd *cobra.Command, r *pipeline.MatchResult, title, company string) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Gap Analysis: %s — %s\n", company, title)
	_, _ = fmt.Fprintf(out, "%s\n", strings.Repeat("=", 50))
	_, _ = fmt.Fprintf(out, "Overall Match Score: %.0f/100\n\n", r.OverallScore*100)
	_, _ = fmt.Fprintln(out, r.Narrative)

	if len(r.StrongMatches) > 0 {
		_, _ = fmt.Fprintf(out, "\nStrong Matches (%d skills)\n", len(r.StrongMatches))
		for _, m := range r.StrongMatches {
			_, _ = fmt.Fprintf(out, "  %-20s → %s (%.0f%%)\n",
				m.Skill, strings.Join(m.ExperienceRefs, ", "), m.Confidence*100)
		}
	}

	if len(r.PartialMatches) > 0 {
		_, _ = fmt.Fprintf(out, "\nPartial Matches (%d skills)\n", len(r.PartialMatches))
		for _, m := range r.PartialMatches {
			_, _ = fmt.Fprintf(out, "  %-20s → %s (%.0f%%)\n",
				m.Skill, strings.Join(m.ExperienceRefs, ", "), m.Confidence*100)
			if m.GapNote != "" {
				_, _ = fmt.Fprintf(out, "    Note: %s\n", m.GapNote)
			}
		}
	}

	if len(r.Gaps) > 0 {
		_, _ = fmt.Fprintf(out, "\nGaps (%d skills)\n", len(r.Gaps))
		for _, g := range r.Gaps {
			req := "preferred"
			if g.Required {
				req = "REQUIRED"
			}
			_, _ = fmt.Fprintf(out, "  %-20s [%s]\n", g.Skill, req)
			if g.Suggestion != "" {
				_, _ = fmt.Fprintf(out, "    → %s\n", g.Suggestion)
			}
		}
	}
}
