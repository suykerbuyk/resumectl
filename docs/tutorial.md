# resumectl Tutorial

A step-by-step guide to setting up, configuring, and using resumectl — the
AI-augmented resume management CLI.

This tutorial walks through the complete workflow: from first install to
exporting a polished, job-targeted resume. Each section builds on the previous
one, so work through them in order if this is your first time.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Installation](#2-installation)
3. [Initialize Your Vault](#3-initialize-your-vault)
4. [Configuration](#4-configuration)
5. [Set Up Your Profile](#5-set-up-your-profile)
6. [Run Preflight Checks](#6-run-preflight-checks)
7. [Ingest Historical Resumes](#7-ingest-historical-resumes)
8. [Manage Your Vault](#8-manage-your-vault)
9. [Parse a Job Posting](#9-parse-a-job-posting)
10. [Run Gap Analysis](#10-run-gap-analysis)
11. [Coaching: Fill the Gaps](#11-coaching-fill-the-gaps)
12. [Build a Targeted Resume](#12-build-a-targeted-resume)
13. [Export to DOCX or PDF](#13-export-to-docx-or-pdf)
14. [End-to-End Workflow Summary](#14-end-to-end-workflow-summary)
15. [Global Flags Reference](#15-global-flags-reference)
16. [Tips and Best Practices](#16-tips-and-best-practices)
17. [Troubleshooting](#17-troubleshooting)

---

## 1. Prerequisites

### Required

- **Go 1.22+** — needed to build from source (`go version` to check)
- **At least one LLM API key** — resumectl supports four providers:

  | Provider | Environment Variable | Alternate Variable | Sign-up |
  |----------|---------------------|--------------------|---------|
  | Anthropic Claude | `ANTHROPIC_API_KEY` | `CLAUDE_API_KEY` | [console.anthropic.com](https://console.anthropic.com) |
  | OpenAI | `OPENAI_API_KEY` | — | [platform.openai.com](https://platform.openai.com) |
  | Google Gemini | `GEMINI_API_KEY` | `GOOGLE_API_KEY` | [aistudio.google.com](https://aistudio.google.com) |
  | xAI Grok | `XAI_API_KEY` | `GROK_API_KEY` | [console.x.ai](https://console.x.ai) |

  Claude is the recommended default. Gemini Flash is used by default for job
  extraction (fast and cheap). You can use a single provider for everything if
  you prefer.

### Optional

- **pandoc** — required only for the `export` command (DOCX/PDF rendering).
  Install via your package manager:

  ```bash
  # Arch Linux
  sudo pacman -S pandoc

  # Debian / Ubuntu
  sudo apt install pandoc

  # macOS
  brew install pandoc

  # Fedora
  sudo dnf install pandoc
  ```

- **pdftotext** (part of poppler) — required only if you want to ingest PDF
  files with `resumectl ingest`:

  ```bash
  # Arch Linux
  sudo pacman -S poppler

  # Debian / Ubuntu
  sudo apt install poppler-utils

  # macOS
  brew install poppler

  # Fedora
  sudo dnf install poppler-utils
  ```

- **LaTeX** — required only for PDF export (`resumectl export --format pdf`).
  Install texlive or equivalent.

- **Git** — recommended for vault version tracking. If your vault directory is
  a Git repository, resumectl auto-commits every vault mutation, giving you a
  full audit trail of changes.

---

## 2. Installation

### From source (recommended)

```bash
git clone https://github.com/jsuykerbuyk/resumectl.git
cd resumectl
make build
```

The binary is placed at `./bin/resumectl`. You can run it directly or install
it system-wide:

```bash
# Install to ~/.local/bin (non-root) or /usr/local/bin (root)
make install

# Or specify a custom prefix
make install PREFIX=/opt/resumectl
```

`make install` also installs man pages to `PREFIX/share/man/man1/`.

### Via `go install`

```bash
go install github.com/jsuykerbuyk/resumectl/cmd/resumectl@latest
```

This places the binary in your `$GOPATH/bin` (usually `~/go/bin`).

### Verify the installation

```bash
resumectl version
# Output: resumectl v0.1.0 (commit: abc1234, built: 2026-03-25T14:30:00Z)
```

---

## 3. Initialize Your Vault

The vault is a directory of structured markdown files — your professional
knowledge base. It is organized like an [Obsidian](https://obsidian.md) vault,
so you can browse and edit it with Obsidian or any text editor.

### Create the vault

```bash
resumectl vault init ~/resume-vault
```

This creates the following directory structure:

```
~/resume-vault/
├── experience/          # One file per role or engagement
├── jobs/
│   └── target/          # Parsed job postings
├── profile/
│   ├── (contact.md)     # You create this — see Section 5
│   ├── (summary-core.md)
│   └── (skills.md)
├── resumes/
│   └── generated/       # LLM-synthesized resumes
├── cover-letters/       # Generated cover letters
├── projects/            # Optional: project showcases
└── sessions/            # Coaching session transcripts
```

### (Optional) Make it a Git repository

```bash
cd ~/resume-vault
git init
git add .
git commit -m "Initial vault structure"
```

With Git enabled, resumectl auto-commits every change it makes — ingested
experience files, parsed job postings, generated resumes, and coaching
updates. This gives you a complete audit trail and easy rollback.

### Verify the vault

```bash
resumectl --vault ~/resume-vault vault validate
```

This confirms all required directories exist. If any are missing, re-run
`vault init`.

---

## 4. Configuration

resumectl uses a YAML config file. On first run, a default config with helpful
comments is written to `~/.config/resumectl/config.yaml`.

### Config file location

The config is found in this precedence order:

1. `--config <path>` CLI flag (highest priority)
2. `$RESUMECTL_CONFIG` environment variable
3. `$XDG_CONFIG_HOME/resumectl/config.yaml`
4. `~/.config/resumectl/config.yaml` (default)

If no config file exists, resumectl uses sensible defaults and reads API keys
from environment variables. You can get by with zero configuration if your
environment variables are set.

### Minimal configuration

If you just want to get started, set your API key(s) as environment variables
and point resumectl at your vault:

```bash
# In your shell profile (~/.bashrc, ~/.zshrc, etc.)
export ANTHROPIC_API_KEY="sk-ant-api03-..."
export GEMINI_API_KEY="AI..."

# Tell resumectl where your vault lives
export RESUMECTL_VAULT="$HOME/resume-vault"
```

That's it. resumectl will use default provider routing and models.

### Full configuration

For more control, create `~/.config/resumectl/config.yaml`:

```yaml
# resumectl configuration
# See: docs/configuration.md for full field reference.
#
# Values in ${VAR} syntax are replaced with environment variables at load time,
# so you never need to put raw API keys in this file.

vault:
  # Path to your Obsidian vault root directory.
  # Override with --vault flag or $RESUMECTL_VAULT.
  path: ~/resume-vault

  git:
    # Auto-commit vault changes to Git (no-op if vault is not a Git repo).
    auto_commit: true
    commit_on_ingest: true    # commit after resumectl ingest
    commit_on_coach: true     # commit after coaching write-backs
    commit_on_build: true     # commit after resume generation

providers:
  # Which provider handles each task type.
  # Options: claude, openai, gemini, grok
  default_for_task:
    decompose: claude          # resume ingestion/decomposition
    extract_job: gemini        # job posting parsing (fast + cheap)
    gap_analysis: claude       # skill gap analysis
    synthesize_resume: claude  # resume generation
    cover_letter: claude       # cover letter generation
    coach: claude              # interactive coaching

  # Anthropic Claude — default for synthesis and analysis.
  claude:
    api_key: ${ANTHROPIC_API_KEY}
    default_model: claude-sonnet-4-6       # used for most tasks
    synthesis_model: claude-opus-4-6       # used for resume generation (highest quality)

  # OpenAI — alternative for any task.
  openai:
    api_key: ${OPENAI_API_KEY}
    default_model: gpt-4o-mini

  # Google Gemini — default for job extraction (fast + cheap).
  gemini:
    api_key: ${GEMINI_API_KEY}
    default_model: gemini-2.0-flash

  # xAI Grok — alternative for any task.
  grok:
    api_key: ${XAI_API_KEY}
    base_url: https://api.x.ai/v1
    default_model: grok-3-mini

export:
  default_format: docx                        # docx or pdf
  pandoc_path: pandoc                         # path to pandoc binary
  templates_dir: ~/.config/resumectl/templates # DOCX reference templates

fetch:
  user_agent: "Mozilla/5.0 (compatible; resumectl/1.0)"
  # use_chromedp: false                       # enable for JS-heavy job pages
  # chromedp_timeout_seconds: 15

coach:
  auto_write_suggestions: false               # skip Y/n for safe vault updates
  session_notes_dir: sessions/                # vault-relative path
```

### Understanding provider routing

Each LLM task is routed to a specific provider. The defaults are chosen for
an optimal balance of speed, cost, and quality:

| Task | Default Provider | Why |
|------|-----------------|-----|
| `decompose` | Claude Sonnet | Strong document structure understanding |
| `extract_job` | Gemini Flash | Fast, cheap extraction |
| `gap_analysis` | Claude Sonnet | Nuanced cross-document reasoning |
| `synthesize_resume` | Claude Opus | Highest writing quality |
| `cover_letter` | Claude Opus | Tone and personalization |
| `coach` | Claude Sonnet | Conversational quality |

You can override routing per-task in config, or per-invocation with
`--provider`:

```bash
# Use Grok for this specific parse-job invocation
resumectl --provider grok parse-job posting.txt
```

### Environment variable interpolation

Config values in `${VAR}` syntax are replaced with environment variables at
load time. This lets you keep secrets out of the config file:

```yaml
providers:
  claude:
    api_key: ${ANTHROPIC_API_KEY}   # resolved from env, never stored in plaintext
```

---

## 5. Set Up Your Profile

Your profile is the foundation of every resume resumectl generates. It
consists of three files in the `profile/` directory. Create these manually
or use `resumectl ingest` (Section 7) to auto-populate experience files and
then fill in the profile by hand.

### 5.1 Contact Information

Create `profile/contact.md` in your vault:

```yaml
---
name: Jane Doe
email: jane.doe@example.com
phone: "+1 555 123 4567"
location: Denver, CO
linkedin: https://linkedin.com/in/janedoe
github: https://github.com/janedoe
website: https://janedoe.dev
---
```

The `name` and `email` fields are required. All others are optional but
recommended — they appear in the header of every generated resume and
exported document.

### 5.2 Professional Summary

Create `profile/summary-core.md`:

```yaml
---
updated: "2026-03-25"
tags: [summary, identity]
---

# Professional Identity

Full-stack engineer with 12 years of experience building high-performance
distributed systems. Career arc spans embedded firmware through cloud-native
platforms, with a consistent focus on reliability and performance at scale.

Known for bridging the gap between architecture and hands-on implementation.
Strong track record of leading cross-functional teams through complex
technical migrations and delivering measurable improvements in system
throughput and operational efficiency.
```

Write 1-3 paragraphs that capture your professional identity — not a resume
summary, but a narrative about who you are as a professional. The LLM uses
this as context to personalize every resume it generates. Focus on:

- Your career trajectory and arc
- Your core technical strengths
- What differentiates you from other candidates
- The kind of impact you consistently deliver

### 5.3 Skills Inventory

Create `profile/skills.md`:

```yaml
---
updated: "2026-03-25"
---

## Skills Inventory

| Skill | Proficiency | Last Used | Years | Category |
|---|---|---|---|---|
| Go | Expert | 2026 | 8 | Languages |
| Python | Advanced | 2026 | 12 | Languages |
| Rust | Intermediate | 2025 | 2 | Languages |
| Kubernetes | Advanced | 2026 | 5 | Cloud |
| AWS | Advanced | 2026 | 7 | Cloud |
| PostgreSQL | Expert | 2026 | 10 | Databases |
| Redis | Advanced | 2025 | 6 | Databases |
| Docker | Advanced | 2026 | 6 | DevOps |
| Terraform | Intermediate | 2025 | 3 | DevOps |
| System Design | Expert | 2026 | 10 | Architecture |

## Skill Clusters

### Backend Systems
Go, Python, gRPC, Kafka, PostgreSQL, Redis

### Cloud Infrastructure
Kubernetes, AWS, GCP, Terraform, Helm, Docker

### Observability
Prometheus, Grafana, OpenTelemetry, Datadog
```

The skills inventory has two parts:

1. **Skills table** — a markdown table with five columns. This is parsed
   row-by-row and used for gap analysis scoring and resume synthesis.

   | Column | Description |
   |--------|-------------|
   | Skill | Skill name (match capitalization used in job postings) |
   | Proficiency | Expert, Advanced, Intermediate, or Beginner |
   | Last Used | Year (YYYY) you last used this skill |
   | Years | Total years of experience |
   | Category | Grouping: Languages, Cloud, Databases, etc. |

2. **Skill clusters** (optional) — free-form groupings below the table. These
   provide additional context to the LLM during coaching and synthesis.

Be thorough here. The skills inventory is the primary input to gap analysis.
Skills not listed here will show up as gaps even if your experience files
mention them. Update this file as you add experience and go through coaching
sessions.

---

## 6. Run Preflight Checks

Before using resumectl for real work, run the preflight validator to catch
issues early:

```bash
resumectl check
```

This validates five areas in order:

### What it checks

**Configuration** — Config file loads, vault path is set.

**Vault** — Directory exists, required subdirectories present, profile files
valid (contact has name + email, summary has content, skills table has rows).

**Providers** — At least one API key is set. Every task in `default_for_task`
maps to a provider that has an API key.

**Tools** — pandoc is on PATH (warning if missing). pdftotext is on PATH
(warning if missing).

**Templates** — All 6 prompt templates load without error.

### Reading the output

Each check reports one of three statuses:

| Symbol | Meaning | Impact |
|--------|---------|--------|
| `✓` | Pass | No issues |
| `!` | Warning | Non-blocking — some features may not work |
| `✗` | Fail | Will cause runtime errors — fix before proceeding |

Every failure and warning includes a `→` line with an actionable fix.

### Example output

```
resumectl check — preflight validation

Configuration
  ✓ Config loaded from /home/jane/.config/resumectl/config.yaml
  ✓ Vault path: /home/jane/resume-vault

Vault
  ✓ Vault directory exists: /home/jane/resume-vault
  ✓ All required directories present
  ✓ profile/contact.md — Jane Doe <jane.doe@example.com>
  ✓ profile/summary-core.md — 87 words
  ✓ profile/skills.md — 10 skills tracked

Providers
  ✓ claude — API key set (ANTHROPIC_API_KEY)
  ✓ gemini — API key set (GEMINI_API_KEY)

Tools
  ✓ pandoc found: /usr/bin/pandoc
  ! pdftotext not found (needed for PDF ingestion only)
    → Install poppler-utils: sudo pacman -S poppler (Arch) or sudo apt install poppler-utils (Debian)

Templates
  ✓ All 6 prompt templates loaded

Result: 0 failure(s), 1 warning(s)
```

Fix any failures before proceeding. Warnings are informational — you can
proceed, but some features (like PDF ingestion or DOCX export) may not work.

---

## 7. Ingest Historical Resumes

If you have existing resumes, resumectl can decompose them into structured
experience files automatically. This is the fastest way to populate your vault.

### Supported formats

| Format | Extension | Requires |
|--------|-----------|----------|
| Word | `.docx` | pandoc |
| PDF | `.pdf` | pdftotext (poppler) |
| Plain text | `.txt` | nothing |
| Markdown | `.md` | nothing |
| HTML | `.html` | nothing |

### Ingest a document

```bash
resumectl ingest ~/Documents/resume-2024.docx
```

What happens behind the scenes:

1. The document is converted to plaintext (via pandoc for DOCX, pdftotext for
   PDF, or read directly for text/markdown).
2. The plaintext is sent to an LLM (Claude Sonnet by default) with a
   specialized prompt to decompose it into individual roles.
3. For each role, resumectl creates a structured experience file in
   `experience/` with YAML frontmatter (role, company, dates, skills, tags)
   and a markdown body (summary, key contributions, technologies).
4. Before writing, each file is checked for duplicates (matching company +
   start date). Duplicates are skipped with a warning.
5. If your vault is a Git repository, each file is auto-committed.
6. After ingestion, a list of suggested skills is printed for you to add to
   `profile/skills.md`.

### Preview before writing

Use `--dry-run` to see what would be created without writing files:

```bash
resumectl --dry-run ingest old-resume.txt
```

### Confirm each file interactively

Use `--interactive` to review and approve each experience file before it is
written:

```bash
resumectl ingest --interactive resume.docx
```

### Use a different provider

```bash
resumectl --provider grok ingest resume.pdf
```

### What the output looks like

After ingestion, your `experience/` directory will contain files like:

```
experience/
├── 2019-06-intel-storage-engineer.md
├── 2021-03-seagate-principal-architect.md
└── 2023-01-cloudscale-staff-engineer.md
```

Each file looks like this:

```yaml
---
role: Principal Storage Architect
company: Seagate Technology
company_slug: seagate
start: "2021-03"
end: "2024-06"
current: false
location: Remote / Bloomington, MN
employment_type: full-time
tags:
  - storage
  - nvme
  - distributed-systems
  - architecture
skills:
  - NVMe
  - Go
  - C++
  - SPDK
domain: storage
highlight: true
visibility: resume
created: "2026-03-25"
updated: "2026-03-25"
---

## Summary

Led storage architecture initiatives at Seagate, focusing on NVMe and
distributed storage systems.

## Key Contributions

- Architected next-gen NVMe storage platform handling 2M IOPS.
- Drove adoption of ZNS across 3 product lines.
- Reduced P99 read latency by 40% through zone scheduling redesign.

## Technologies

NVMe, ZNS, SPDK, Go, C++, Linux kernel, distributed storage
```

### After ingestion: review and enrich

Ingested files are a starting point. Review each one and:

- **Add missing contributions** — the LLM can only extract what was in your
  original document. Add achievements, metrics, and context you remember.
- **Fix tags and skills** — ensure they match your skills inventory naming.
- **Set `highlight: true`** on your strongest roles (used for resume
  ordering).
- **Set `visibility`** — `resume` (default, always considered), `hidden`
  (excluded), or `cover-letter` (only in cover letters).
- **Update skills.md** — add any suggested skills from the ingestion output.

---

## 8. Manage Your Vault

resumectl provides several commands to inspect and manage your vault.

### List experience files

```bash
resumectl vault ls
```

Output:

```
experience/2019-06-intel-storage-engineer.md     Storage Engineer @ Intel  (2019-06 – 2021-02)
experience/2021-03-seagate-principal-architect.md Principal Architect @ Seagate  (2021-03 – 2024-06)
experience/2023-01-cloudscale-staff-engineer.md  Staff Engineer @ CloudScale  (2023-01 – present)
```

### View vault statistics

```bash
resumectl vault stats
```

Output:

```
Vault: /home/jane/resume-vault
  Experience files: 3
  Job files:        0
  Skills tracked:   10
  Git repo:         true
```

### View the tag index

The tag index shows how your experience files are tagged, and which skills
and domains are covered. This is the same index used for Stage 1 gap analysis.

```bash
resumectl vault index
```

Output:

```
Tags:
  storage              2 files (intel, seagate)
  nvme                 2 files (intel, seagate)
  kubernetes           1 files (cloudscale)
  go                   2 files (seagate, cloudscale)
  distributed-systems  2 files (seagate, cloudscale)

Skills:
  NVMe                 2 files (intel, seagate)
  Go                   2 files (seagate, cloudscale)
  Kubernetes           1 files (cloudscale)
  C++                  1 files (seagate)

Domains:
  storage              2 files (intel, seagate)
  cloud                1 files (cloudscale)
```

Use `vault index` to verify coverage before running `match`. If a skill you
have isn't showing up, check that it's in the `skills` or `tags` frontmatter
of the relevant experience files.

### Validate vault structure

```bash
resumectl vault validate
```

Reports any missing required directories.

### Manually add experience files

You don't have to use `ingest`. You can create experience files manually.
Follow the naming convention `YYYY-MM-{company-slug}-{role-slug}.md` and
include the required frontmatter fields. See the example in Section 7 above.

At minimum, include:

- `role` (string, required)
- `company` (string, required)
- `company_slug` (string, required)
- `start` (string, "YYYY-MM", required)
- `current` (bool, required)
- `tags` (string array — used for tag intersection scoring)
- `skills` (string array — used for tag intersection scoring)
- `domain` (string — used for domain matching)

---

## 9. Parse a Job Posting

When you find a job you want to target, parse it into a structured vault file:

### From a URL

```bash
resumectl parse-job https://jobs.lever.co/acme/abc123
```

resumectl fetches the page, strips HTML, and sends the text to an LLM
(Gemini Flash by default) to extract structured data: title, company,
required/preferred skills, seniority, compensation, culture signals, and
more.

### From a local file

```bash
resumectl parse-job job-description.txt
```

Works with `.txt`, `.md`, and `.html` files.

### From the clipboard

```bash
# macOS
pbpaste | resumectl parse-job -

# Linux (X11)
xclip -selection clipboard -o | resumectl parse-job -

# Linux (Wayland)
wl-paste | resumectl parse-job -
```

The `-` argument tells resumectl to read from stdin.

### Preview without writing

```bash
resumectl --dry-run parse-job posting.txt
```

### Open in your editor after parsing

```bash
resumectl parse-job --open job-description.txt
```

Opens the created file in `$EDITOR` so you can review and annotate it.

### Override the filename slug

```bash
resumectl parse-job --slug acme-dream-job posting.txt
```

### Output

The parsed job is written to `jobs/target/YYYY-MM-DD-{company-slug}-{title-slug}.md`:

```yaml
---
title: Staff Software Engineer — Infrastructure
company: Acme Corp
company_slug: acme-corp
source_url: https://jobs.lever.co/acme/abc123
date_parsed: "2026-03-25"
status: targeting
seniority: staff
domain: infrastructure
required_skills:
  - Go
  - Kubernetes
  - distributed-systems
  - storage
preferred_skills:
  - Rust
  - NVMe
  - eBPF
culture_signals:
  - high-performance
  - systems-focused
compensation:
  range_low: 200000
  range_high: 260000
  currency: USD
  equity: true
location: Remote
tags:
  - infrastructure
  - storage
  - cloud
  - go
---

## Job Description (Raw)

We are looking for a Staff Software Engineer to join our infrastructure team...

## Parsed Summary

Staff-level infrastructure role focused on distributed storage systems at scale.

## Key Requirements

### Must Have
- 8+ years of systems programming experience
- Strong Go or Rust background
- Experience with Kubernetes at scale

### Nice to Have
- NVMe or storage subsystem experience
- eBPF for observability or networking
```

### Track job status

After parsing, the job status is set to `targeting`. Update it manually as
you progress through the application process:

- `targeting` — initial interest, analyzing fit
- `applied` — application submitted
- `interviewing` — active interview process
- `closed` — position filled or withdrawn
- `rejected` — rejected by company
- `offer` — offer received

Edit the `status` field in the job file's frontmatter directly.

---

## 10. Run Gap Analysis

Gap analysis tells you how well your experience matches a target job. It uses
a two-stage process: fast local scoring followed by deep LLM analysis.

```bash
resumectl match jobs/target/2026-03-25-acme-corp-staff-engineer.md
```

### Stage 1: Local tag intersection

resumectl builds a tag index from all your experience files (the same one
shown by `vault index`). It scores each experience file against the job's
required and preferred skills using case-insensitive string matching. The top
N files (default: 5) are selected as candidates.

### Stage 2: LLM semantic analysis

The top candidate experience files, the full job posting, and your skills
inventory are sent to an LLM for deep semantic analysis. The LLM evaluates
nuances that string matching can't — adjacent skills, transferable experience,
leadership overlap, and domain alignment.

### Output

```
Gap Analysis: Staff Software Engineer — Infrastructure @ Acme Corp

Overall Score: 78/100

Strong Matches:
  ✓ Go (Expert, 8 years) — evidence in Seagate, CloudScale roles
  ✓ Distributed Systems — 5 years across 2 roles
  ✓ Storage Architecture — core domain match

Partial Matches:
  ~ Kubernetes — 2 years, job asks for "at scale" (>100 nodes)
    Gap: Document cluster sizes and operational scope

Gaps:
  ✗ eBPF — no evidence in vault
    → Have you used eBPF for tracing, networking, or observability in any
      project? Consider personal projects or experiments.
  ✗ Rust — Intermediate proficiency, job prefers "strong background"
    → What Rust projects have you built? Any production deployments?

Narrative: Strong systems and storage background aligns well with this role.
The main gaps are in eBPF (preferred, not required) and Kubernetes scale.
Consider coaching to surface relevant experience.
```

### Options

```bash
# Output as JSON (for scripting)
resumectl match --format json jobs/target/acme.md

# Limit to top 3 experience files
resumectl match --top 3 jobs/target/acme.md

# Pipe JSON to jq
resumectl match --format json jobs/target/acme.md | jq '.gaps[]'
```

### Interpreting the results

- **Strong matches** — skills with high-confidence evidence in your vault.
  These will feature prominently in your generated resume.
- **Partial matches** — skills with adjacent or incomplete evidence. Coaching
  can help surface experience you haven't documented.
- **Gaps** — skills the job requires that your vault doesn't cover. Each gap
  includes a coaching suggestion.
- **Overall score** — 0-100 percentage. Above 70 is a strong match. Below 50
  suggests significant gaps.

---

## 11. Coaching: Fill the Gaps

The coaching loop walks you through each skill gap identified by `match`,
asking targeted questions to help you recall experience you may not have
documented.

```bash
resumectl coach jobs/target/2026-03-25-acme-corp-staff-engineer.md
```

### How it works

1. Gap analysis runs automatically to identify gaps.
2. For each gap, the coach:
   - Explains why this skill matters for the target role.
   - Asks a targeted question to jog your memory.
   - Reads your answer from the terminal.
   - Suggests a vault update: which file to modify, what to add.
3. You approve or skip each suggestion.

### Example session

```
Gap: eBPF (preferred)

Context: This role involves building observability tooling for a distributed
storage platform. eBPF is used for low-overhead tracing and network monitoring
without kernel module development.

Question: Have you used eBPF, BPF, or similar kernel tracing tools (perf,
SystemTap, DTrace) for performance analysis, network monitoring, or
observability? Even experimental or personal project use counts.

Your answer: > I used perf and ftrace extensively at Seagate for NVMe latency
profiling. I also wrote a small eBPF program for a personal project to trace
block I/O patterns.

Suggested update:
  File: experience/2021-03-seagate-principal-architect.md
  Add bullet: "Used perf and ftrace for NVMe latency profiling across the
  storage stack"
  Add tag: ebpf

Apply? [Y/n] y
Updated experience/2021-03-seagate-principal-architect.md
```

### Options

```bash
# Only coach on required skill gaps (skip preferred/nice-to-have)
resumectl coach --gap-only jobs/target/acme.md
```

### Tips

- **Be specific** — mention projects, scale, production vs. prototype.
- **Include metrics** — numbers strengthen resume bullets.
- **Don't stretch** — if you genuinely lack a skill, skip it. Honest gaps are
  better than fabricated experience.
- **Run match again** after coaching to see your updated score.

---

## 12. Build a Targeted Resume

Once your vault is populated and gaps are addressed, generate a targeted
resume:

```bash
resumectl build jobs/target/2026-03-25-acme-corp-staff-engineer.md
```

### What happens

1. Gap analysis runs automatically to identify the most relevant experience
   files.
2. Your profile data is loaded: contact info, professional summary, and skills
   inventory.
3. The most relevant experience files (up to 5 by default) are selected based
   on match scores.
4. Everything is sent to an LLM (Claude Opus by default — the highest quality
   model for writing) with a synthesis prompt.
5. The LLM generates a complete, ATS-compatible resume with:
   - A tailored professional summary specific to this job
   - Experience sections ordered by relevance (not chronology)
   - Rewritten contribution bullets with quantified impact
   - Core competencies aligned to the job requirements
6. The output is validated: presence of summary section, word count (target
   600-800 words), no hallucinated company names.
7. Written to `resumes/generated/YYYY-MM-DD-{company-slug}.md`.

### Options

```bash
# Also generate a cover letter (printed to stdout)
resumectl build --cover-letter jobs/target/acme.md

# Limit to top 3 experience files
resumectl build --max-experience 3 jobs/target/acme.md

# Write to a custom output path
resumectl build --output ~/Desktop/acme-resume.md jobs/target/acme.md

# Preview without writing
resumectl --dry-run build jobs/target/acme.md
```

### Output format

The generated resume includes YAML frontmatter for traceability:

```yaml
---
job_file: jobs/target/2026-03-25-acme-corp-staff-engineer.md
generated: "2026-03-25T14:30:00Z"
model: claude-opus-4-6
status: draft
experience_files:
  - experience/2021-03-seagate-principal-architect.md
  - experience/2023-01-cloudscale-staff-engineer.md
version: 1
---

# Jane Doe
jane.doe@example.com | +1 555 123 4567 | Denver, CO
https://linkedin.com/in/janedoe | https://github.com/janedoe

## Professional Summary

Seasoned systems engineer with 12+ years building high-performance distributed
systems and storage platforms...

## Core Competencies

Go - Kubernetes - Distributed Systems - Storage Architecture - NVMe - AWS

## Professional Experience

### Principal Storage Architect
**Seagate Technology** | Mar 2021 - Jun 2024

- Architected next-generation NVMe storage platform handling 2M+ IOPS...
...
```

### Review and edit

The generated resume is a draft. Review it and make manual edits:

- Verify all claims are accurate.
- Adjust tone and emphasis.
- Trim or expand sections.
- Set `status: ready` when satisfied.

---

## 13. Export to DOCX or PDF

Convert your finished resume to a polished document:

### Export to DOCX (default)

```bash
resumectl export resumes/generated/2026-03-25-acme-corp.md
```

Output: `2026-03-25-acme-corp.docx` in the current directory.

### Export to PDF

```bash
resumectl export --format pdf resumes/generated/2026-03-25-acme-corp.md
```

Requires pandoc and a LaTeX distribution (texlive).

### Custom output path

```bash
resumectl export --output ~/Desktop/Jane-Doe-Resume.docx resumes/generated/acme.md
```

### Use a Word template for styling

```bash
resumectl export --template ~/.config/resumectl/templates/modern.docx resume.md
```

The template is a Word reference document that defines fonts, margins, heading
styles, and page layout. Pandoc applies these styles to the generated document.

To create a template:
1. Open Word and style a document with your preferred fonts, margins, and
   heading styles.
2. Delete all content (keep the styles).
3. Save as `.docx`.
4. Pass it with `--template`.

### What happens during export

1. **Assembly** — Contact information from `profile/contact.md` is injected as
   a header (name as h1, contact details below). Heading levels in the resume
   body are normalized (h1 demoted to h2, since h1 is reserved for your name).

2. **Rendering** — The assembled markdown is converted via pandoc to the
   target format.

### Preview assembly without rendering

```bash
resumectl --dry-run export resume.md
```

---

## 14. End-to-End Workflow Summary

Here is the complete workflow from installation to exported resume:

```bash
# 1. Install
git clone https://github.com/jsuykerbuyk/resumectl.git
cd resumectl && make install

# 2. Set API keys
export ANTHROPIC_API_KEY="sk-ant-..."
export GEMINI_API_KEY="AI..."

# 3. Initialize vault
resumectl vault init ~/resume-vault
cd ~/resume-vault && git init && git add . && git commit -m "Init vault"
export RESUMECTL_VAULT=~/resume-vault

# 4. Create profile files (see Section 5)
# Edit: profile/contact.md, profile/summary-core.md, profile/skills.md

# 5. Run preflight checks
resumectl check

# 6. Ingest your existing resume
resumectl ingest ~/Documents/my-resume.docx

# 7. Review and enrich ingested experience files
# Edit files in experience/ to add missing achievements and fix tags

# 8. Parse a target job posting
resumectl parse-job https://jobs.lever.co/acme/abc123

# 9. Run gap analysis
resumectl match jobs/target/2026-03-25-acme-corp-staff-engineer.md

# 10. Fill gaps with coaching
resumectl coach jobs/target/2026-03-25-acme-corp-staff-engineer.md

# 11. Generate a targeted resume
resumectl build jobs/target/2026-03-25-acme-corp-staff-engineer.md

# 12. Optional: generate with cover letter
resumectl build --cover-letter jobs/target/2026-03-25-acme-corp-staff-engineer.md

# 13. Export to DOCX
resumectl export resumes/generated/2026-03-25-acme-corp.md

# 14. Review the generated document and apply
```

For subsequent job applications, repeat steps 8-13. Your vault accumulates
knowledge over time — each coaching session and ingestion enriches it, making
future resumes stronger.

---

## 15. Global Flags Reference

These flags apply to every command:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--vault` | string | `./vault` | Path to vault root (also `$RESUMECTL_VAULT`) |
| `--config` | string | `~/.config/resumectl/config.yaml` | Config file path (also `$RESUMECTL_CONFIG`) |
| `--provider` | string | (from config) | Override LLM provider: claude, openai, gemini, grok |
| `--verbose` | bool | false | Enable verbose logging |
| `--dry-run` | bool | false | Run pipeline without writing files |

---

## 16. Tips and Best Practices

### Vault management

- **Use Git** — initialize your vault as a Git repo. resumectl auto-commits
  every change, giving you a full audit trail and easy rollback.
- **One role per file** — don't combine multiple positions at the same company
  into one file. Each role should be a separate experience file.
- **Tag generously** — the gap analysis tag intersection stage depends on
  having thorough tags and skills in your experience frontmatter. More tags
  mean better matching.
- **Keep skills.md current** — this is the primary scoring input for gap
  analysis. Update it after every coaching session.

### Working with providers

- **Set multiple API keys** — even if you primarily use Claude, having Gemini
  set up for job extraction saves money (Flash is very cheap for extraction).
- **Use `--provider` to experiment** — try different providers for the same
  task to compare output quality.
- **Claude Opus for synthesis** — resume generation benefits from the highest
  quality writing. The default config uses Opus for this task specifically.

### Resume quality

- **Coach before building** — run `coach` at least once before `build`. Even
  one coaching session often surfaces undocumented experience that
  significantly improves the generated resume.
- **Review every resume** — LLM output is a draft, not a finished product.
  Verify all claims, adjust emphasis, and ensure the tone matches your voice.
- **Iterate** — enrich your vault, re-run `match` to see improved scores,
  then `build` again. Each iteration produces a better resume.
- **Don't over-optimize** — a 60-70% match score is realistic for most
  senior roles. Not every gap needs to be filled.

### Multiple job applications

- Parse each job posting separately with `parse-job`.
- Run `match` for each to understand fit.
- Use `build` to generate a unique, targeted resume for each position.
- Your vault grows with each coaching session, improving matches across
  all future applications.

### Organizing your vault in Obsidian

Since the vault is standard Obsidian-compatible markdown with YAML frontmatter,
you can:

- Browse and edit files visually in Obsidian.
- Use Obsidian's graph view to see relationships between experience files.
- Use Obsidian's search to find specific skills or companies across your vault.
- Add your own notes and annotations — resumectl only reads the frontmatter
  and standard body sections.

---

## 17. Troubleshooting

### "no provider API keys configured"

No API keys are set. Either:
- Set environment variables: `export ANTHROPIC_API_KEY=sk-ant-...`
- Or add keys to config: `providers.claude.api_key: ${ANTHROPIC_API_KEY}`

Run `resumectl check` to verify.

### "task X → provider Y (no API key)"

A task is routed to a provider that doesn't have an API key. Either:
- Set the API key for that provider.
- Or change the routing in `providers.default_for_task` to use a provider
  that has a key.

### "vault directory does not exist"

The configured vault path doesn't exist. Run:
```bash
resumectl vault init <path>
```

### "missing required directories"

The vault exists but is missing subdirectories. Re-run `vault init` to create
them:
```bash
resumectl vault init ~/resume-vault
```

### "pandoc not found"

The `export` command requires pandoc. Install it:
```bash
# See Section 1 for platform-specific install commands
sudo pacman -S pandoc    # Arch
sudo apt install pandoc  # Debian/Ubuntu
brew install pandoc      # macOS
```

### "pdftotext not found"

Only needed for ingesting PDF files. Install poppler:
```bash
sudo pacman -S poppler         # Arch
sudo apt install poppler-utils # Debian/Ubuntu
brew install poppler           # macOS
```

### LLM returns empty or malformed JSON

Occasionally, an LLM response may not parse correctly. This is transient.
Retry the command. If it persists:
- Try a different provider: `--provider openai`
- Check your API key hasn't expired or hit quota limits.
- Use `--verbose` for detailed request/response logging.

### "429 Too Many Requests"

You've hit the provider's rate limit. Wait a moment and retry. If persistent,
check your billing/quota on the provider's dashboard.

### Export produces a document with no styling

Use a Word reference template for styled output:
```bash
resumectl export --template my-template.docx resume.md
```

Without a template, pandoc uses its default styling.

### Job posting URL returns incomplete content

Some job sites (LinkedIn, Workday) render content with JavaScript that simple
HTTP fetching can't capture. Options:
- Copy the job description text and save it to a `.txt` file, then parse that.
- Use stdin: copy the text, then `pbpaste | resumectl parse-job -`
- Enable headless Chrome in config: `fetch.use_chromedp: true` (requires
  Chrome/Chromium installed).

---

## Further Reading

- [Configuration Reference](configuration.md) — complete field reference for
  `config.yaml`
- [README](../README.md) — project overview and quickstart
- `resumectl <command> --help` — built-in help for every command
- `man resumectl` — man pages (after `make install` or `make man`)
