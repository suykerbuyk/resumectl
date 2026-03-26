# Configuration Reference

resumectl is configured via a YAML file. All fields are optional — sensible
defaults are used when a field is absent.

## Config File Location

The config file is searched in this order:

1. `--config <path>` CLI flag (highest priority)
2. `$RESUMECTL_CONFIG` environment variable
3. `$XDG_CONFIG_HOME/resumectl/config.yaml`
4. `~/.config/resumectl/config.yaml` (default fallback)

If no config file is found, defaults are used and API keys are read from
environment variables.

## Environment Variable Interpolation

Any value in `${VAR}` syntax is replaced with the corresponding environment
variable at load time. Unset variables resolve to empty strings.

```yaml
providers:
  claude:
    api_key: ${ANTHROPIC_API_KEY}   # resolved from env at load time
```

## API Key Environment Variables

When a provider's `api_key` is empty in config, resumectl checks these
environment variables as a fallback:

| Provider | Primary env var | Alternate env var |
|----------|----------------|-------------------|
| Claude | `ANTHROPIC_API_KEY` | `CLAUDE_API_KEY` |
| OpenAI | `OPENAI_API_KEY` | |
| Gemini | `GEMINI_API_KEY` | `GOOGLE_API_KEY` |
| Grok | `XAI_API_KEY` | `GROK_API_KEY` |

## Field Reference

### vault

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `vault.path` | string | `./vault` | Path to the Obsidian vault root directory. Override with `--vault` flag or `$RESUMECTL_VAULT` env var. |
| `vault.backup_on_write` | bool | `false` | Reserved for future use. |
| `vault.git.auto_commit` | bool | `true` | Automatically `git add` + `git commit` after every vault mutation. No-op if the vault root is not a Git repository. |
| `vault.git.commit_on_ingest` | bool | `true` | Commit after `resumectl ingest` writes experience files. Only applies if `auto_commit` is true. |
| `vault.git.commit_on_coach` | bool | `true` | Commit after coaching write-backs modify experience files. Only applies if `auto_commit` is true. |
| `vault.git.commit_on_build` | bool | `true` | Commit after `resumectl build` writes a generated resume. Only applies if `auto_commit` is true. |

### providers

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `providers.default_for_task` | map[string]string | see below | Maps each task type to a provider name. |
| `providers.claude.api_key` | string | `""` | Anthropic API key. Falls back to `$ANTHROPIC_API_KEY` or `$CLAUDE_API_KEY`. |
| `providers.claude.default_model` | string | `claude-sonnet-4-6` | Model used for most Claude tasks. |
| `providers.claude.synthesis_model` | string | `claude-opus-4-6` | Model used for resume synthesis (highest quality). |
| `providers.openai.api_key` | string | `""` | OpenAI API key. Falls back to `$OPENAI_API_KEY`. |
| `providers.openai.default_model` | string | `gpt-4o-mini` | Model used for OpenAI tasks. Available: `gpt-4o`, `gpt-4o-mini`, `o3-mini`. |
| `providers.gemini.api_key` | string | `""` | Google Gemini API key. Falls back to `$GEMINI_API_KEY` or `$GOOGLE_API_KEY`. |
| `providers.gemini.default_model` | string | `gemini-2.0-flash` | Model used for Gemini tasks. |
| `providers.grok.api_key` | string | `""` | xAI Grok API key. Falls back to `$XAI_API_KEY` or `$GROK_API_KEY`. |
| `providers.grok.base_url` | string | `https://api.x.ai/v1` | Grok API base URL (OpenAI-compatible endpoint). |
| `providers.grok.default_model` | string | `grok-3-mini` | Model used for Grok tasks. |

#### Default task-to-provider routing

| Task | Default Provider | Rationale |
|------|-----------------|-----------|
| `decompose` | claude | Strong document structure understanding |
| `extract_job` | gemini | Fast and cost-effective for extraction |
| `gap_analysis` | claude | Nuanced cross-document reasoning |
| `synthesize_resume` | claude | Highest writing quality |
| `cover_letter` | claude | Tone and personalization critical |
| `coach` | claude | Conversational quality |

Override any task's provider in config or use `--provider` on the command line.

### export

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `export.default_format` | string | `docx` | Default output format for `resumectl export`. |
| `export.pandoc_path` | string | `pandoc` | Path to the pandoc binary. Use an absolute path if pandoc is not on `$PATH`. |
| `export.templates_dir` | string | `~/.config/resumectl/templates` | Directory containing DOCX reference templates for pandoc `--reference-doc`. |

### fetch

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fetch.user_agent` | string | `Mozilla/5.0 (compatible; resumectl/1.0)` | HTTP User-Agent header for URL fetching. |
| `fetch.chromedp_timeout_seconds` | int | `15` | Timeout for headless Chrome rendering (if enabled). |
| `fetch.use_chromedp` | bool | `false` | Use headless Chrome for JavaScript-rendered pages. Requires Chrome/Chromium installed. |

### coach

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `coach.auto_write_suggestions` | bool | `false` | If true, skip the Y/n confirmation prompt for low-risk vault updates during coaching. |
| `coach.session_notes_dir` | string | `sessions/` | Vault-relative directory for coaching session transcripts. |

### gdocs (optional)

Google Docs integration is entirely optional. These fields are only needed if
you use `resumectl ingest` with Google Docs URLs or `resumectl export --gdocs`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `gdocs.credentials_file` | string | `""` | Path to Google OAuth2 client credentials JSON file. |
| `gdocs.token_cache` | string | `""` | Path to cache the OAuth2 refresh token. |
| `gdocs.default_folder_id` | string | `""` | Google Drive folder ID for uploaded documents. |
