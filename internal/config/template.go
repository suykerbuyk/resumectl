package config

// DefaultConfigYAML is the annotated default config file written on first run.
// All values shown are the defaults. Uncomment and modify as needed.
const DefaultConfigYAML = `# resumectl configuration
# See: doc/configuration.md for full field reference.
#
# Values in ${VAR} syntax are replaced with environment variables at load time.
# API keys can also be set via environment variables directly — see below.

vault:
  # Path to your Obsidian vault root directory.
  # Override with --vault flag or $RESUMECTL_VAULT.
  # path: ./vault

  git:
    # Automatically git-commit after every vault mutation.
    # Requires the vault root to be a git repository.
    auto_commit: true
    # commit_on_ingest: true
    # commit_on_coach: true
    # commit_on_build: true

providers:
  # Which provider handles each task type.
  # Options: claude, openai, gemini, grok
  default_for_task:
    decompose: claude
    extract_job: gemini
    gap_analysis: claude
    synthesize_resume: claude
    cover_letter: claude
    coach: claude

  # Anthropic Claude — default for synthesis and analysis.
  # API key: set ANTHROPIC_API_KEY env var or uncomment below.
  claude:
    # api_key: ${ANTHROPIC_API_KEY}
    default_model: claude-sonnet-4-6
    synthesis_model: claude-opus-4-6

  # OpenAI — alternative for any task.
  # API key: set OPENAI_API_KEY env var or uncomment below.
  openai:
    # api_key: ${OPENAI_API_KEY}
    default_model: gpt-4o-mini

  # Google Gemini — default for job extraction (fast + cheap).
  # API key: set GEMINI_API_KEY env var or uncomment below.
  gemini:
    # api_key: ${GEMINI_API_KEY}
    default_model: gemini-2.0-flash

  # xAI Grok — alternative for any task.
  # API key: set XAI_API_KEY env var or uncomment below.
  grok:
    # api_key: ${XAI_API_KEY}
    base_url: https://api.x.ai/v1
    default_model: grok-3-mini

export:
  default_format: docx
  # pandoc_path: pandoc
  # templates_dir: ~/.config/resumectl/templates

# fetch:
#   user_agent: "Mozilla/5.0 (compatible; resumectl/1.0)"
#   chromedp_timeout_seconds: 15
#   use_chromedp: false

# coach:
#   auto_write_suggestions: false
#   session_notes_dir: sessions/
`
