package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", "config", name)
}

func TestLoad_ValidConfig(t *testing.T) {
	cfg, err := Load(testdataPath("valid.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "/tmp/test-vault", cfg.Vault.Path)
	assert.True(t, cfg.Vault.BackupOnWrite)
	assert.True(t, cfg.Vault.Git.AutoCommit)
	assert.Equal(t, "test-key-123", cfg.Providers.Claude.APIKey)
	assert.Equal(t, "claude-sonnet-4-6", cfg.Providers.Claude.DefaultModel)
	assert.Equal(t, "gemini-2.0-flash", cfg.Providers.Gemini.DefaultModel)
	assert.Equal(t, "test-agent/1.0", cfg.Fetch.UserAgent)
	assert.Equal(t, 10, cfg.Fetch.ChromedpTimeout)
	assert.Equal(t, "docx", cfg.Export.DefaultFormat)
	assert.Equal(t, "sessions/", cfg.Coach.SessionNotesDir)
	assert.Equal(t, "claude", cfg.Providers.DefaultForTask["decompose"])
	assert.Equal(t, "gemini", cfg.Providers.DefaultForTask["extract_job"])
}

func TestLoad_EnvVarInterpolation(t *testing.T) {
	t.Setenv("TEST_CLAUDE_KEY", "claude-secret-abc")
	t.Setenv("TEST_GEMINI_KEY", "gemini-secret-xyz")

	cfg, err := Load(testdataPath("with-envvars.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "claude-secret-abc", cfg.Providers.Claude.APIKey)
	assert.Equal(t, "gemini-secret-xyz", cfg.Providers.Gemini.APIKey)
}

func TestLoad_EnvVarUnset(t *testing.T) {
	// Ensure the env vars are NOT set
	t.Setenv("TEST_CLAUDE_KEY", "")
	t.Setenv("TEST_GEMINI_KEY", "")

	cfg, err := Load(testdataPath("with-envvars.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "", cfg.Providers.Claude.APIKey)
	assert.Equal(t, "", cfg.Providers.Gemini.APIKey)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config: read")
}

func TestLoad_MalformedYAML(t *testing.T) {
	_, err := Load(testdataPath("malformed.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config: parse")
}

func TestLoad_DefaultFallback(t *testing.T) {
	// Load valid config that only sets vault.path — everything else should
	// have defaults from DefaultConfig since yaml.Unmarshal merges into it.
	cfg, err := Load(testdataPath("with-envvars.yaml"))
	require.NoError(t, err)

	// These should retain defaults since the file doesn't set them
	assert.True(t, cfg.Vault.Git.AutoCommit)
	assert.Equal(t, "claude-sonnet-4-6", cfg.Providers.Claude.DefaultModel)
	assert.Equal(t, "pandoc", cfg.Export.PandocPath)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "./vault", cfg.Vault.Path)
	assert.True(t, cfg.Vault.Git.AutoCommit)
	assert.Equal(t, "claude", cfg.Providers.DefaultForTask["decompose"])
	assert.Equal(t, "gemini", cfg.Providers.DefaultForTask["extract_job"])
	assert.Equal(t, "claude-sonnet-4-6", cfg.Providers.Claude.DefaultModel)
	assert.Equal(t, "claude-opus-4-6", cfg.Providers.Claude.SynthesisModel)
	assert.Equal(t, "gemini-2.0-flash", cfg.Providers.Gemini.DefaultModel)
	assert.Equal(t, "grok-3-mini", cfg.Providers.Grok.DefaultModel)
	assert.Equal(t, "https://api.x.ai/v1", cfg.Providers.Grok.BaseURL)
	assert.Equal(t, "docx", cfg.Export.DefaultFormat)
	assert.Equal(t, 15, cfg.Fetch.ChromedpTimeout)
	assert.Contains(t, cfg.Fetch.UserAgent, "resumectl")
}

func TestInterpolateEnvVars(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("BAZ_123", "qux")

	tests := []struct {
		input    string
		expected string
	}{
		{"${FOO}", "bar"},
		{"prefix-${FOO}-suffix", "prefix-bar-suffix"},
		{"${FOO}-${BAZ_123}", "bar-qux"},
		{"no-vars-here", "no-vars-here"},
		{"${UNSET_VAR}", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, interpolateEnvVars(tt.input))
		})
	}
}
