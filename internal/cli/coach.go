package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/pipeline"
	"github.com/spf13/cobra"
)

func newCoachCmd() *cobra.Command {
	var gapOnly bool

	cmd := &cobra.Command{
		Use:   "coach <job-file>",
		Short: "Interactive coaching loop to address skill gaps",
		Long: `Run an interactive coaching session to address skill gaps.

Gap analysis (resumectl match) is run first to identify skills the target
job requires but your vault does not document. The coach then walks through
each gap one by one.

SESSION FLOW

For each gap, the coach:

  1. Sends the gap details (skill, requirement level, your existing context)
     to an LLM with the coach_question prompt template.
  2. Displays context explaining why this skill matters for the target role.
  3. Asks a targeted question to help you recall relevant experience.
  4. Reads your answer from stdin.
  5. Suggests a vault update: which file to modify, what to add (a skill
     tag, a contribution bullet, or a technology entry).

Type your answer at the prompt to provide context. Type "skip" or press
Enter with no input to skip a gap. Press Ctrl-C to end the session early.

VAULT ENRICHMENT

The coach's primary purpose is to enrich your vault so that future resumes
are stronger. Each coaching session may surface experience you forgot to
document — a personal project, an adjacent skill, or a contribution that
wasn't in your original resume. The suggested updates feed directly back
into the vault, improving match scores and synthesis quality for this job
and future jobs with similar requirements.

Use --gap-only to focus only on required skills (skipping preferred/nice-to-have).`,
		Example: `  # Coach through all gaps for a job
  resumectl coach jobs/target/2026-03-25-acme-staff-engineer.md

  # Only address required skill gaps
  resumectl coach --gap-only jobs/target/2026-03-25-acme.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initDeps(); err != nil {
				return err
			}

			job, err := deps.Vault.LoadJob(args[0])
			if err != nil {
				return err
			}

			// Run match first to identify gaps
			matchProvider, err := getProvider(llm.TaskGapAnalysis)
			if err != nil {
				return err
			}

			matchResult, err := pipeline.RunMatch(cmd.Context(), job, deps.Vault, deps.Prompts, matchProvider, 5)
			if err != nil {
				return err
			}

			gaps := matchResult.Gaps
			if len(gaps) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No gaps found! Your experience is a strong match.")
				return nil
			}

			coachProvider, err := getProvider(llm.TaskCoach)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Found %d gap(s) to address.\n\n", len(gaps))

			scanner := bufio.NewScanner(os.Stdin)

			for i, gap := range gaps {
				if gapOnly && !gap.Required {
					continue
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "─── Gap %d of %d: %s", i+1, len(gaps), gap.Skill)
				if gap.Required {
					_, _ = fmt.Fprint(cmd.OutOrStdout(), " [REQUIRED]")
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), " ───")

				result, err := pipeline.RunCoach(cmd.Context(), pipeline.CoachInput{
					Gap:             gap,
					JobTitle:        job.Frontmatter.Title,
					JobCompany:      job.Frontmatter.Company,
					ExistingContext: gap.Suggestion,
				}, deps.Prompts, coachProvider)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "⚠ coaching error: %v\n\n", err)
					continue
				}

				if result.Context != "" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Context)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nQuestion: %s\n\n> ", result.Question)

				if !scanner.Scan() {
					break
				}
				answer := strings.TrimSpace(scanner.Text())

				if answer == "" || strings.ToLower(answer) == "skip" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Skipped.")
					continue
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nSuggested update:\n")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  File:   %s\n", result.Suggestion.Suggestion.TargetFile)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Action: %s\n", result.Suggestion.Suggestion.UpdateType)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Add:    %s\n\n", result.Suggestion.Suggestion.Content)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Coaching session complete.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&gapOnly, "gap-only", false, "only ask about required skill gaps")

	return cmd
}
