package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/spf13/cobra"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Vault management commands",
		Long: `Manage the Obsidian vault that stores your professional history.

The vault is a directory of structured markdown files with YAML frontmatter.
Each file type has a defined schema and lives in a specific subdirectory:

  experience/          One file per role (YYYY-MM-{company}-{role}.md)
                       Frontmatter: role, company, dates, tags, skills, domain
                       Body: Summary, Key Contributions, Technologies, Notes

  jobs/target/         Parsed job postings (YYYY-MM-DD-{company}-{title}.md)
                       Frontmatter: title, company, skills, compensation, status

  profile/contact.md   Name, email, phone, location, LinkedIn, GitHub, website
  profile/summary-core.md  Professional identity narrative for LLM context
  profile/skills.md    Skills inventory table + skill clusters

  resumes/generated/   LLM-synthesized resumes with source tracking
  cover-letters/       Generated cover letters
  sessions/            Coaching session transcripts

The vault path defaults to ./vault and can be set with --vault, $RESUMECTL_VAULT,
or the vault.path field in ~/.config/resumectl/config.yaml. If the vault is a Git repository,
mutations are auto-committed (configurable in vault.git.auto_commit).`,
		Example: `  resumectl vault init ~/resume-vault
  resumectl vault validate
  resumectl vault ls
  resumectl vault stats
  resumectl vault index`,
	}

	cmd.AddCommand(
		newVaultInitCmd(),
		newVaultValidateCmd(),
		newVaultLsCmd(),
		newVaultStatsCmd(),
		newVaultIndexCmd(),
	)

	return cmd
}

func newVaultInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new vault directory structure",
		Long: `Create the directory skeleton for a new resumectl vault.

Creates the following directories:

  experience/          For role/engagement experience files
  jobs/target/         For parsed job posting files
  profile/             For contact, summary, and skills files
  resumes/generated/   For LLM-synthesized resume output
  cover-letters/       For generated cover letters
  projects/            For project documentation
  sessions/            For coaching session notes

If a path argument is given, the vault is created there. Otherwise, the
vault path from config is used (default: ./vault).

After initialization, populate profile/contact.md, profile/summary-core.md,
and profile/skills.md manually or by running 'resumectl ingest' with
historical documents. The profile files are required for resume synthesis.`,
		Example: `  # Initialize in a specific directory
  resumectl vault init ~/Documents/resume-vault

  # Initialize using the configured default path
  resumectl vault init

  # Initialize and verify
  resumectl vault init ~/vault && resumectl --vault ~/vault vault validate`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cfg.Vault.Path
			if len(args) > 0 {
				root = args[0]
			}

			if err := vault.Init(root); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Vault initialized at %s\n", root)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created directories:\n")
			for _, dir := range vault.AllDirs {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s/\n", dir)
			}
			return nil
		},
	}
}

func newVaultValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate vault directory structure",
		Long: `Check that all required vault directories exist.

The following directories are required for resumectl to function:

  experience/          Must exist for ingest, match, build
  jobs/target/         Must exist for parse-job, match, build
  profile/             Must exist for build (contact, summary, skills)
  resumes/generated/   Must exist for build output

Returns exit code 0 if all directories exist, or exit code 1 with a list
of missing directories. Run 'resumectl vault init' to create them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVault()
			if err != nil {
				return err
			}

			missing := v.Validate()
			if len(missing) > 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Vault validation FAILED. Missing directories:\n")
				for _, dir := range missing {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s/\n", dir)
				}
				return fmt.Errorf("vault has %d missing directories", len(missing))
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Vault at %s is valid.\n", v.Root)
			return nil
		},
	}
}

func newVaultLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List experience files in the vault",
		Long: `List all experience files in the vault with role, company, and date range.

Each line shows the vault-relative file path, role title, company name,
and the employment date range. Files with current=true in frontmatter
show "present" as the end date.

Output is one line per file, sorted by filename (which encodes start date).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVault()
			if err != nil {
				return err
			}

			files, err := v.AllExperience()
			if err != nil {
				return err
			}

			if len(files) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No experience files found.")
				return nil
			}

			for _, f := range files {
				dateRange := f.Frontmatter.Start
				if f.Frontmatter.Current {
					dateRange += " – present"
				} else if f.Frontmatter.End != "" {
					dateRange += " – " + f.Frontmatter.End
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-50s  %s @ %s  (%s)\n",
					f.RelPath, f.Frontmatter.Role, f.Frontmatter.Company, dateRange)
			}
			return nil
		},
	}
}

func newVaultStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Print vault statistics",
		Long: `Display summary statistics for the vault.

Reports:
  - Number of experience files in experience/
  - Number of job posting files in jobs/target/
  - Number of skills in the profile/skills.md inventory table
  - Whether the vault root is a Git repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVault()
			if err != nil {
				return err
			}

			experience, err := v.AllExperience()
			if err != nil {
				return err
			}
			jobs, err := v.AllJobs()
			if err != nil {
				return err
			}

			skills, err := v.LoadSkills()
			skillCount := 0
			if err == nil {
				skillCount = len(skills.Skills)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Vault: %s\n", v.Root)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Experience files: %d\n", len(experience))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Job files:        %d\n", len(jobs))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Skills tracked:   %d\n", skillCount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Git repo:         %v\n", v.IsGitRepo())
			return nil
		},
	}
}

func newVaultIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Build and display the tag index",
		Long: `Rebuild the tag index from all experience files and print it.

The tag index is built by scanning the YAML frontmatter of every file in
experience/. Three indices are constructed:

  Tags       From the 'tags' frontmatter field (lowercase)
  Skills     From the 'skills' frontmatter field (lowercase)
  Domains    From the 'domain' frontmatter field (lowercase)

Each entry shows the key, the number of experience files containing it,
and the company slugs of those files. This is the same index used by
'resumectl match' for Stage 1 tag intersection scoring.

Use this command to verify vault coverage before running match or build,
and to identify gaps in your tagging.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVault()
			if err != nil {
				return err
			}

			if err := v.RebuildIndex(); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Tags:")
			printIndex(cmd, v.Index.ByTag)

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nSkills:")
			printIndex(cmd, v.Index.BySkill)

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nDomains:")
			printIndex(cmd, v.Index.ByDomain)

			return nil
		},
	}
}

func printIndex(cmd *cobra.Command, m map[string][]*vault.ExperienceFile) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		files := m[k]
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Frontmatter.CompanySlug
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %d files (%s)\n", k, len(files), strings.Join(names, ", "))
	}
}

func openVault() (*vault.Vault, error) {
	return vault.Open(cfg.Vault.Path, cfg)
}
