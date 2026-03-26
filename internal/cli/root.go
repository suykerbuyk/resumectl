package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jsuykerbuyk/resumectl/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Version info set via ldflags at build time.
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var (
	cfgFile     string
	vaultPath   string
	providerOvr string
	verbose     bool
	dryRun      bool

	cfg *config.Config
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "resumectl",
		Short: "AI-augmented resume management CLI",
		Long: `resumectl is a local-first, AI-augmented resume management system.

It treats your professional history as a queryable knowledge base stored in a
structured Obsidian vault — a directory of markdown files with YAML frontmatter.
LLM pipelines (Claude, OpenAI, Gemini, Grok) parse job postings, perform gap analysis
against your documented experience, synthesize targeted resumes and cover
letters, and coach you through skill gaps.

The vault stores experience files, job postings, profile data (contact, skills,
professional summary), and generated resumes. Every vault mutation is optionally
auto-committed to Git for a full audit trail.

TYPICAL WORKFLOW

  1. Initialize a vault:       resumectl vault init ~/vault
  2. Import old resumes:       resumectl ingest ~/Documents/resume.docx
  3. Parse a job posting:      resumectl parse-job <url-or-file>
  4. Analyze skill gaps:       resumectl match <job-file>
  5. Fill gaps interactively:  resumectl coach <job-file>
  6. Generate a resume:        resumectl build <job-file>
  7. Export to DOCX/PDF:       resumectl export <resume-file>

CONFIGURATION

resumectl reads configuration from ~/.config/resumectl/config.yaml (override with --config
or $RESUMECTL_CONFIG). If no config file exists, sensible defaults are used.
API keys can be set via the config file or environment variables:

  ANTHROPIC_API_KEY   Claude (Anthropic) API key
  OPENAI_API_KEY      OpenAI API key
  GEMINI_API_KEY      Gemini (Google) API key
  XAI_API_KEY         Grok (xAI) API key

The vault path defaults to ./vault and can be overridden with --vault or
$RESUMECTL_VAULT. Each LLM task (extraction, analysis, synthesis) is routed
to a configurable provider; use --provider to override for a single invocation.

EXTERNAL DEPENDENCIES

  pandoc    Required for DOCX/PDF export (resumectl export)
  pdftotext Required for PDF ingestion (resumectl ingest *.pdf)`,
		Example: `  # End-to-end: parse a job posting and generate a tailored resume
  resumectl parse-job https://linkedin.com/jobs/view/123456
  resumectl match jobs/target/2026-03-25-acme-staff-engineer.md
  resumectl build jobs/target/2026-03-25-acme-staff-engineer.md
  resumectl export resumes/generated/2026-03-25-acme-corp.md

  # Use a specific provider for all operations
  resumectl --provider grok parse-job posting.txt

  # Dry-run to preview without writing files
  resumectl --dry-run build jobs/target/2026-03-25-acme.md

  # Point to a non-default vault
  resumectl --vault ~/my-vault vault stats`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip config loading for version command or if already loaded (tests)
			if cmd.Name() == "version" || cfg != nil {
				return nil
			}
			return loadConfig()
		},
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/resumectl/config.yaml)")
	root.PersistentFlags().StringVar(&vaultPath, "vault", "", "path to Obsidian vault root (default: $RESUMECTL_VAULT or ./vault)")
	root.PersistentFlags().StringVar(&providerOvr, "provider", "", "override default LLM provider (claude|openai|gemini|grok)")
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "run pipeline without writing files")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newVaultCmd())
	root.AddCommand(newIngestCmd())
	root.AddCommand(newParseJobCmd())
	root.AddCommand(newMatchCmd())
	root.AddCommand(newBuildCmd())
	root.AddCommand(newCoachCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newDocCmd())

	return root
}

func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig() error {
	path := cfgFile
	if path == "" {
		if envPath := os.Getenv("RESUMECTL_CONFIG"); envPath != "" {
			path = envPath
		} else {
			path = defaultConfigPath()
			if path == "" {
				cfg = config.DefaultConfig()
				applyOverrides(cfg)
				return nil
			}
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Auto-create a default config on first run (only for the default path)
		if cfgFile == "" && os.Getenv("RESUMECTL_CONFIG") == "" {
			_ = writeDefaultConfig(path)
		}
		cfg = config.DefaultConfig()
		applyOverrides(cfg)
		return nil
	}

	loaded, err := config.Load(path)
	if err != nil {
		return err
	}
	cfg = loaded
	applyOverrides(cfg)

	return nil
}

// applyOverrides applies CLI flag overrides and env var fallbacks to a loaded config.
func applyOverrides(c *config.Config) {
	if vaultPath != "" {
		c.Vault.Path = vaultPath
	} else if envVault := os.Getenv("RESUMECTL_VAULT"); envVault != "" {
		c.Vault.Path = envVault
	}
	applyEnvKeyFallbacks(c)
}

// applyEnvKeyFallbacks fills in provider API keys from environment variables
// when not already set by the config file.
func applyEnvKeyFallbacks(c *config.Config) {
	if c.Providers.Claude.APIKey == "" {
		c.Providers.Claude.APIKey = firstEnv("ANTHROPIC_API_KEY", "CLAUDE_API_KEY")
	}
	if c.Providers.OpenAI.APIKey == "" {
		c.Providers.OpenAI.APIKey = firstEnv("OPENAI_API_KEY")
	}
	if c.Providers.Gemini.APIKey == "" {
		c.Providers.Gemini.APIKey = firstEnv("GEMINI_API_KEY", "GOOGLE_API_KEY")
	}
	if c.Providers.Grok.APIKey == "" {
		c.Providers.Grok.APIKey = firstEnv("XAI_API_KEY", "GROK_API_KEY")
	}
}

// writeDefaultConfig creates an annotated default config file at the given path.
// Creates the parent directory if needed. Errors are ignored — best-effort convenience.
func writeDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(config.DefaultConfigYAML), 0o644)
}

// defaultConfigPath returns the XDG-compliant config file path.
// Checks $XDG_CONFIG_HOME/resumectl/config.yaml first, then falls back to
// ~/.config/resumectl/config.yaml. Returns "" if the home directory cannot
// be determined.
func defaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/resumectl/config.yaml"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.config/resumectl/config.yaml"
}

// firstEnv returns the value of the first non-empty environment variable.
func firstEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}
