package config

// DefaultConfig returns a Config with sane default values.
func DefaultConfig() *Config {
	return &Config{
		Vault: VaultConfig{
			Path:          "./vault",
			BackupOnWrite: false,
			Git: GitConfig{
				AutoCommit:     true,
				CommitOnIngest: true,
				CommitOnCoach:  true,
				CommitOnBuild:  true,
			},
		},
		Providers: ProvidersConfig{
			DefaultForTask: map[string]string{
				"decompose":         "claude",
				"extract_job":       "gemini",
				"gap_analysis":      "claude",
				"synthesize_resume": "claude",
				"cover_letter":      "claude",
				"coach":             "claude",
			},
			Claude: ProviderEntry{
				DefaultModel:   "claude-sonnet-4-6",
				SynthesisModel: "claude-opus-4-6",
			},
			OpenAI: ProviderEntry{
				DefaultModel: "gpt-4o-mini",
			},
			Gemini: ProviderEntry{
				DefaultModel: "gemini-2.0-flash",
			},
			Grok: ProviderEntry{
				BaseURL:      "https://api.x.ai/v1",
				DefaultModel: "grok-3-mini",
			},
		},
		Export: ExportConfig{
			DefaultFormat: "docx",
			PandocPath:    "pandoc",
			TemplatesDir:  "~/.config/resumectl/templates",
		},
		Fetch: FetchConfig{
			ChromedpTimeout: 15,
			UserAgent:       "Mozilla/5.0 (compatible; resumectl/1.0)",
		},
		Coach: CoachConfig{
			SessionNotesDir: "sessions/",
		},
	}
}
