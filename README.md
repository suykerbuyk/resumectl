# resumectl

**AI-augmented resume management CLI for the command line.**

resumectl treats your professional history as a queryable knowledge base stored
in a structured [Obsidian](https://obsidian.md) vault. LLM pipelines parse job
postings, perform gap analysis against your experience, synthesize targeted
resumes and cover letters, and coach you through skill gaps.

## Features

- **Local-first** — vault data stays on disk; only task-specific context is sent to LLM providers
- **Vault-backed** — structured markdown with YAML frontmatter, readable in Obsidian
- **Multi-provider** — Claude, OpenAI, Gemini, and Grok behind a common interface
- **Git-tracked** — every vault mutation is auto-committed for full audit trail
- **Gap analysis** — two-stage matching (local tag intersection + LLM semantic scoring)
- **Coaching loop** — interactive sessions to fill skill gaps and enrich your vault
- **Export** — render to DOCX/PDF via pandoc
- **Preflight check** — `resumectl check` validates config, vault, API keys, and tools before you start

## Installation

### From source

```bash
git clone https://github.com/jsuykerbuyk/resumectl.git
cd resumectl
make build
# Binary is at ./bin/resumectl
```

### Go install

```bash
go install github.com/jsuykerbuyk/resumectl/cmd/resumectl@latest
```

## Quickstart

```bash
# 1. Initialize a vault
resumectl vault init ~/resume-vault

# 2. Import historical resumes
resumectl ingest ~/Documents/resume-2022.docx

# 3. Parse a target job posting
resumectl parse-job https://www.linkedin.com/jobs/view/1234567890

# 4. Run gap analysis
resumectl match jobs/target/2026-03-25-acme-staff-engineer.md

# 5. Generate a tailored resume
resumectl build jobs/target/2026-03-25-acme-staff-engineer.md

# 6. Address skill gaps interactively
resumectl coach jobs/target/2026-03-25-acme-staff-engineer.md

# 7. Export to DOCX
resumectl export resumes/generated/2026-03-25-acme-corp.md
```

## Command Reference

| Command | Description |
|---------|-------------|
| `vault init [path]` | Initialize a new vault directory structure |
| `vault validate` | Validate vault directories exist |
| `vault ls` | List experience files |
| `vault stats` | Print vault statistics |
| `vault index` | Build and display the tag index |
| `ingest <source>` | Decompose historical documents into experience files |
| `parse-job <source>` | Parse a job posting (URL, file, or stdin) |
| `match <job-file>` | Run gap analysis against the vault |
| `build <job-file>` | Synthesize a targeted resume |
| `coach <job-file>` | Interactive coaching to fill skill gaps |
| `export <resume>` | Render resume to DOCX or PDF |
| `check` | Validate config, vault, providers, and tools |

Run `resumectl <command> --help` for detailed usage, flags, and examples.

### Man pages

```bash
make man          # Generate man pages to man/
man man/resumectl.1
```

## Vault Structure

```
vault/
├── experience/          # One file per role (YYYY-MM-slug.md)
├── jobs/
│   └── target/          # Parsed job postings (YYYY-MM-DD-slug.md)
├── profile/
│   ├── contact.md       # Name, email, phone, links
│   ├── summary-core.md  # Professional identity narrative
│   └── skills.md        # Skills inventory table + clusters
├── resumes/
│   └── generated/       # LLM-synthesized resumes
├── cover-letters/
├── projects/
└── sessions/            # Coaching session notes
```

### Experience file format

```yaml
---
role: Principal Storage Architect
company: Seagate Technology
company_slug: seagate
start: "2021-03"
end: "2024-06"
current: false
tags: [storage, nvme, distributed-systems]
skills: [NVMe, Go, C++, SPDK]
domain: storage
highlight: true
visibility: resume
---

## Summary
Led storage architecture initiatives...

## Key Contributions
- Architected next-gen NVMe storage platform handling 2M IOPS.
- Reduced P99 read latency by 40%.
```

## Provider Setup

resumectl supports four LLM providers. Set API keys as environment variables:

```bash
# Anthropic Claude (default for synthesis and gap analysis)
export ANTHROPIC_API_KEY=sk-ant-...

# OpenAI (alternative for any task)
export OPENAI_API_KEY=sk-proj-...

# Google Gemini (default for job extraction)
export GEMINI_API_KEY=AI...

# xAI Grok (alternative for any task)
export XAI_API_KEY=xai-...
```

### Task-to-provider routing

| Task | Default Provider | Rationale |
|------|-----------------|-----------|
| Resume decomposition | Claude Sonnet | Strong document understanding |
| Job extraction | Gemini Flash | Fast and cheap |
| Gap analysis | Claude Sonnet | Nuanced cross-document reasoning |
| Resume synthesis | Claude Opus | Highest writing quality |
| Cover letter | Claude Opus | Tone and personalization |
| Coaching | Configurable | Any provider works |

All defaults are overrideable in config or with `--provider`.

## Configuration

resumectl follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/).
The config file is searched in this order:

1. `--config <path>` flag (highest priority)
2. `$RESUMECTL_CONFIG` environment variable
3. `$XDG_CONFIG_HOME/resumectl/config.yaml`
4. `~/.config/resumectl/config.yaml` (default)

If no config file is found, sensible defaults are used. API keys can be
provided via the config file or environment variables (see Provider Setup above).

Values in `${VAR}` syntax are interpolated from environment variables at
load time, so you never need to put raw API keys in the config file.

See [doc/configuration.md](doc/configuration.md) for a complete field reference.

### Minimal config example

```yaml
vault:
  path: ~/resume-vault

providers:
  claude:
    api_key: ${ANTHROPIC_API_KEY}
```

### Full config example

```yaml
vault:
  path: ~/resume-vault
  backup_on_write: false
  git:
    auto_commit: true        # commit vault after every mutation
    commit_on_ingest: true   # commit after resumectl ingest
    commit_on_coach: true    # commit after coaching write-backs
    commit_on_build: true    # commit generated resume files

providers:
  default_for_task:
    decompose: claude        # which provider handles each task
    extract_job: gemini
    gap_analysis: claude
    synthesize_resume: claude
    cover_letter: claude
    coach: claude

  claude:
    api_key: ${ANTHROPIC_API_KEY}
    default_model: claude-sonnet-4-6
    synthesis_model: claude-opus-4-6

  openai:
    api_key: ${OPENAI_API_KEY}
    default_model: gpt-4o-mini

  gemini:
    api_key: ${GEMINI_API_KEY}
    default_model: gemini-2.0-flash

  grok:
    api_key: ${XAI_API_KEY}
    base_url: https://api.x.ai/v1
    default_model: grok-3-mini

export:
  default_format: docx
  pandoc_path: pandoc
  templates_dir: ~/.config/resumectl/templates

fetch:
  user_agent: "Mozilla/5.0 (compatible; resumectl/1.0)"
  chromedp_timeout_seconds: 15
  use_chromedp: false

coach:
  auto_write_suggestions: false
  session_notes_dir: sessions/

gdocs:
  credentials_file: ~/.config/resumectl/google-credentials.json
  token_cache: ~/.config/resumectl/google-token.json
  default_folder_id: ""
```

## Global Flags

```
--vault PATH       Path to vault root (default: $RESUMECTL_VAULT or ./vault)
--config PATH      Config file path (default: ~/.config/resumectl/config.yaml)
--provider NAME    Override LLM provider (claude|gemini|grok)
--verbose          Enable verbose logging
--dry-run          Run pipeline without writing files
```

## External Dependencies

| Tool | Purpose | Required? |
|------|---------|-----------|
| [pandoc](https://pandoc.org) | DOCX/PDF rendering | For `export` command |
| pdftotext (poppler) | PDF text extraction | For `ingest` with PDFs |

## Development

```bash
make test          # Run unit tests
make test-cover    # Run tests with 80% coverage gate
make lint          # Run golangci-lint
make check         # Run all checks (vet + lint + test-cover)
make docs          # Generate man pages + markdown docs
```

## License

TBD
