package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/spf13/cobra"
)

const (
	statusPass = "✓"
	statusWarn = "!"
	statusFail = "✗"
)

type checkResult struct {
	failures int
	warnings int
}

func (r *checkResult) pass(w io.Writer, msg string) {
	_, _ = fmt.Fprintf(w, "  %s %s\n", statusPass, msg)
}

func (r *checkResult) warn(w io.Writer, msg, guidance string) {
	r.warnings++
	_, _ = fmt.Fprintf(w, "  %s %s\n", statusWarn, msg)
	_, _ = fmt.Fprintf(w, "    → %s\n", guidance)
}

func (r *checkResult) fail(w io.Writer, msg, guidance string) {
	r.failures++
	_, _ = fmt.Fprintf(w, "  %s %s\n", statusFail, msg)
	_, _ = fmt.Fprintf(w, "    → %s\n", guidance)
}

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Validate configuration, vault, providers, and tools",
		Long: `Run preflight validation to catch configuration issues before they
become runtime errors.

Checks five areas in order:

  Configuration   Config file loads, vault path is set
  Vault           Directory exists, required subdirs present, profile files valid
  Providers       API keys configured for all default task providers
  Tools           External tools (pandoc, pdftotext) available on PATH
  Templates       All prompt templates load without error

Each check reports PASS, WARN (non-blocking), or FAIL (will break at runtime).
Every failure and warning includes guidance on how to resolve the issue.

Exit code 0 if no failures, 1 if any check fails.`,
		Example: `  # Run all preflight checks
  resumectl check

  # Check with a specific vault path
  resumectl --vault ~/my-vault check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(out, "resumectl check — preflight validation")

			result := &checkResult{}

			checkConfiguration(out, result)
			checkVault(out, result)
			checkProviders(out, result)
			checkTools(out, result)
			checkTemplates(out, result)

			// Summary
			_, _ = fmt.Fprintln(out)
			if result.failures == 0 && result.warnings == 0 {
				_, _ = fmt.Fprintln(out, "All checks passed.")
			} else {
				_, _ = fmt.Fprintf(out, "Result: %d failure(s), %d warning(s)\n",
					result.failures, result.warnings)
			}

			if result.failures > 0 {
				return fmt.Errorf("%d preflight check(s) failed", result.failures)
			}
			return nil
		},
	}
}

func checkConfiguration(w io.Writer, r *checkResult) {
	_, _ = fmt.Fprintln(w, "\nConfiguration")

	if cfg == nil {
		r.fail(w, "no configuration loaded",
			"Run resumectl with a valid --config path or ensure ~/.config/resumectl/config.yaml exists")
		return
	}

	// Report config source
	if cfgFile != "" {
		r.pass(w, fmt.Sprintf("Config loaded from %s (--config flag)", cfgFile))
	} else if envPath := os.Getenv("RESUMECTL_CONFIG"); envPath != "" {
		r.pass(w, fmt.Sprintf("Config loaded from %s ($RESUMECTL_CONFIG)", envPath))
	} else {
		path := defaultConfigPath()
		if _, err := os.Stat(path); err == nil {
			r.pass(w, fmt.Sprintf("Config loaded from %s", path))
		} else {
			r.pass(w, "Using default configuration (no config file found)")
		}
	}

	// Vault path
	if cfg.Vault.Path != "" {
		r.pass(w, fmt.Sprintf("Vault path: %s", cfg.Vault.Path))
	} else {
		r.fail(w, "vault path not configured",
			"Set vault.path in config, use --vault, or set $RESUMECTL_VAULT")
	}
}

func checkVault(w io.Writer, r *checkResult) {
	_, _ = fmt.Fprintln(w, "\nVault")

	if cfg == nil || cfg.Vault.Path == "" {
		r.fail(w, "cannot check vault (no path configured)",
			"Set vault.path in config, use --vault, or set $RESUMECTL_VAULT")
		return
	}

	absPath, err := filepath.Abs(cfg.Vault.Path)
	if err != nil {
		r.fail(w, fmt.Sprintf("cannot resolve vault path: %v", err),
			"Check that the vault path is a valid directory path")
		return
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		r.fail(w, fmt.Sprintf("vault directory does not exist: %s", absPath),
			fmt.Sprintf("Create it with: resumectl vault init %s", cfg.Vault.Path))
		return
	}
	if err != nil {
		r.fail(w, fmt.Sprintf("cannot access vault: %v", err),
			"Check directory permissions")
		return
	}
	if !info.IsDir() {
		r.fail(w, fmt.Sprintf("vault path is not a directory: %s", absPath),
			"vault.path must point to a directory, not a file")
		return
	}

	r.pass(w, fmt.Sprintf("Vault directory exists: %s", absPath))

	// Open vault for further checks
	v, err := vault.Open(cfg.Vault.Path, cfg)
	if err != nil {
		r.fail(w, fmt.Sprintf("cannot open vault: %v", err),
			"Check vault directory structure and permissions")
		return
	}

	// Required directories
	missing := v.Validate()
	if len(missing) > 0 {
		r.fail(w, fmt.Sprintf("missing required directories: %s", strings.Join(missing, ", ")),
			fmt.Sprintf("Run: resumectl vault init %s", cfg.Vault.Path))
	} else {
		r.pass(w, "All required directories present")
	}

	// Profile files
	checkProfileFiles(w, r, v)
}

func checkProfileFiles(w io.Writer, r *checkResult, v *vault.Vault) {
	// contact.md
	contact, err := v.LoadContact()
	if err != nil {
		r.fail(w, "profile/contact.md — missing or invalid",
			"Create profile/contact.md with YAML frontmatter: name, email, phone, location, linkedin")
	} else {
		issues := []string{}
		if contact.Name == "" {
			issues = append(issues, "name is empty")
		}
		if contact.Email == "" {
			issues = append(issues, "email is empty")
		}
		if len(issues) > 0 {
			r.fail(w, fmt.Sprintf("profile/contact.md — %s", strings.Join(issues, ", ")),
				"Add the missing fields to the YAML frontmatter in profile/contact.md")
		} else {
			r.pass(w, fmt.Sprintf("profile/contact.md — %s <%s>", contact.Name, contact.Email))
		}
	}

	// summary-core.md
	summary, err := v.LoadSummary()
	if err != nil {
		r.warn(w, "profile/summary-core.md — missing or invalid",
			"Create profile/summary-core.md with your professional identity narrative (used by build)")
	} else {
		wordCount := len(strings.Fields(summary.Body))
		if wordCount < 10 {
			r.warn(w, fmt.Sprintf("profile/summary-core.md — very short (%d words)", wordCount),
				"Add a 1-2 paragraph professional identity narrative for better resume synthesis")
		} else {
			r.pass(w, fmt.Sprintf("profile/summary-core.md — %d words", wordCount))
		}
	}

	// skills.md
	skills, err := v.LoadSkills()
	if err != nil {
		r.warn(w, "profile/skills.md — missing or invalid",
			"Create profile/skills.md with a markdown table: | Skill | Proficiency | Last Used | Years | Category |")
	} else {
		if len(skills.Skills) == 0 {
			r.warn(w, "profile/skills.md — no skills found in table",
				"Add rows to the skills inventory table in profile/skills.md")
		} else {
			r.pass(w, fmt.Sprintf("profile/skills.md — %d skills tracked", len(skills.Skills)))
		}
	}
}

func checkProviders(w io.Writer, r *checkResult) {
	_, _ = fmt.Fprintln(w, "\nProviders")

	if cfg == nil {
		r.fail(w, "cannot check providers (no configuration)",
			"Ensure configuration is loaded")
		return
	}

	// Check each provider's API key status
	type providerInfo struct {
		name   string
		hasKey bool
		envVar string // primary env var name for guidance
	}

	providers := []providerInfo{
		{"claude", cfg.Providers.Claude.APIKey != "", "ANTHROPIC_API_KEY"},
		{"openai", cfg.Providers.OpenAI.APIKey != "", "OPENAI_API_KEY"},
		{"gemini", cfg.Providers.Gemini.APIKey != "", "GEMINI_API_KEY"},
		{"grok", cfg.Providers.Grok.APIKey != "", "XAI_API_KEY"},
	}

	keyedProviders := map[string]bool{}
	for _, p := range providers {
		if p.hasKey {
			keyedProviders[p.name] = true
		}
	}

	if len(keyedProviders) == 0 {
		r.fail(w, "no provider API keys configured",
			"Set at least one: ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, or XAI_API_KEY")
		return
	}

	// Report each provider
	for _, p := range providers {
		if p.hasKey {
			r.pass(w, fmt.Sprintf("%s — API key set (%s)", p.name, p.envVar))
		}
	}

	// Check task routing — each default task must map to a provider with a key
	allTasks := []llm.TaskType{
		llm.TaskDecompose, llm.TaskExtractJob, llm.TaskGapAnalysis,
		llm.TaskSynthesizeResume, llm.TaskCoverLetter, llm.TaskCoach,
	}

	for _, task := range allTasks {
		providerName, ok := cfg.Providers.DefaultForTask[string(task)]
		if !ok {
			r.fail(w, fmt.Sprintf("task %q has no provider mapping", task),
				fmt.Sprintf("Add providers.default_for_task.%s to config", task))
			continue
		}
		if !keyedProviders[providerName] {
			// Find which providers DO have keys for the guidance
			available := []string{}
			for _, p := range providers {
				if p.hasKey {
					available = append(available, p.name)
				}
			}
			r.fail(w, fmt.Sprintf("task %q → %s (no API key)", task, providerName),
				fmt.Sprintf("Set %s's API key, or change to an available provider: %s",
					providerName, strings.Join(available, ", ")))
		}
	}
}

func checkTools(w io.Writer, r *checkResult) {
	_, _ = fmt.Fprintln(w, "\nTools")

	if path, err := exec.LookPath("pandoc"); err == nil {
		r.pass(w, fmt.Sprintf("pandoc found: %s", path))
	} else {
		r.warn(w, "pandoc not found (needed for export command)",
			"Install pandoc: https://pandoc.org/installing.html")
	}

	if path, err := exec.LookPath("pdftotext"); err == nil {
		r.pass(w, fmt.Sprintf("pdftotext found: %s", path))
	} else {
		r.warn(w, "pdftotext not found (needed for PDF ingestion only)",
			"Install poppler-utils: sudo pacman -S poppler (Arch) or sudo apt install poppler-utils (Debian)")
	}
}

func checkTemplates(w io.Writer, r *checkResult) {
	_, _ = fmt.Fprintln(w, "\nTemplates")

	lib, err := prompts.NewLibrary()
	if err != nil {
		r.fail(w, fmt.Sprintf("prompt library failed to load: %v", err),
			"This is a bug — prompt templates are embedded at compile time. Rebuild the binary.")
		return
	}

	for _, name := range prompts.TemplateNames {
		if _, err := lib.Get(name); err != nil {
			r.fail(w, fmt.Sprintf("template %q failed: %v", name, err),
				"This is a bug — rebuild the binary.")
		}
	}

	r.pass(w, fmt.Sprintf("All %d prompt templates loaded", len(prompts.TemplateNames)))
}
