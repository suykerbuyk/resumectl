package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all resumectl configuration.
type Config struct {
	Vault     VaultConfig     `yaml:"vault"`
	Providers ProvidersConfig `yaml:"providers"`
	Export    ExportConfig    `yaml:"export"`
	GDocs     GDocsConfig     `yaml:"gdocs"`
	Fetch     FetchConfig     `yaml:"fetch"`
	Coach     CoachConfig     `yaml:"coach"`
}

type VaultConfig struct {
	Path          string    `yaml:"path"`
	BackupOnWrite bool      `yaml:"backup_on_write"`
	Git           GitConfig `yaml:"git"`
}

type GitConfig struct {
	AutoCommit     bool `yaml:"auto_commit"`
	CommitOnIngest bool `yaml:"commit_on_ingest"`
	CommitOnCoach  bool `yaml:"commit_on_coach"`
	CommitOnBuild  bool `yaml:"commit_on_build"`
}

type ProvidersConfig struct {
	DefaultForTask map[string]string `yaml:"default_for_task"`
	Claude         ProviderEntry     `yaml:"claude"`
	OpenAI         ProviderEntry     `yaml:"openai"`
	Gemini         ProviderEntry     `yaml:"gemini"`
	Grok           ProviderEntry     `yaml:"grok"`
}

type ProviderEntry struct {
	APIKey         string `yaml:"api_key"`
	BaseURL        string `yaml:"base_url"`
	DefaultModel   string `yaml:"default_model"`
	SynthesisModel string `yaml:"synthesis_model"`
}

type ExportConfig struct {
	DefaultFormat string `yaml:"default_format"`
	PandocPath    string `yaml:"pandoc_path"`
	TemplatesDir  string `yaml:"templates_dir"`
}

type GDocsConfig struct {
	CredentialsFile string `yaml:"credentials_file"`
	TokenCache      string `yaml:"token_cache"`
	DefaultFolderID string `yaml:"default_folder_id"`
}

type FetchConfig struct {
	UseChromedp     bool   `yaml:"use_chromedp"`
	ChromedpTimeout int    `yaml:"chromedp_timeout_seconds"`
	UserAgent       string `yaml:"user_agent"`
}

type CoachConfig struct {
	AutoWriteSuggestions bool   `yaml:"auto_write_suggestions"`
	SessionNotesDir      string `yaml:"session_notes_dir"`
}

// envVarPattern matches ${VAR_NAME} placeholders.
var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Load reads a YAML config file from the given path, interpolates environment
// variables in the form ${VAR}, and returns a Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	expanded := interpolateEnvVars(string(data))

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	return cfg, nil
}

// interpolateEnvVars replaces all ${VAR} placeholders with their environment
// variable values. Unset variables are replaced with empty strings.
func interpolateEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		return os.Getenv(varName)
	})
}
