package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Help(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "resumectl")
	assert.Contains(t, buf.String(), "AI-augmented")
}

func TestRootCmd_UnknownSubcommand(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"nonexistent"})

	err := root.Execute()
	require.Error(t, err)
}

func TestVersionCmd(t *testing.T) {
	Version = "1.2.3"
	Commit = "abc123"
	BuildDate = "2026-03-25"
	defer func() {
		Version = "dev"
		Commit = "none"
		BuildDate = "unknown"
	}()

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "1.2.3")
	assert.Contains(t, buf.String(), "abc123")
	assert.Contains(t, buf.String(), "2026-03-25")
}

func TestRootCmd_GlobalFlags(t *testing.T) {
	root := NewRootCmd()

	assert.NotNil(t, root.PersistentFlags().Lookup("config"))
	assert.NotNil(t, root.PersistentFlags().Lookup("vault"))
	assert.NotNil(t, root.PersistentFlags().Lookup("provider"))
	assert.NotNil(t, root.PersistentFlags().Lookup("verbose"))
	assert.NotNil(t, root.PersistentFlags().Lookup("dry-run"))
}

func TestLoadConfig_FromFile(t *testing.T) {
	cfgFile = filepath.Join("..", "..", "testdata", "config", "valid.yaml")
	defer func() { cfgFile = "" }()

	err := loadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "/tmp/test-vault", cfg.Vault.Path)
}

func TestLoadConfig_MissingFileUsesDefaults(t *testing.T) {
	cfgFile = "/nonexistent/config.yaml"
	defer func() { cfgFile = "" }()

	err := loadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "./vault", cfg.Vault.Path)
}

func TestLoadConfig_EnvConfigPath(t *testing.T) {
	cfgFile = ""
	t.Setenv("RESUMECTL_CONFIG", filepath.Join("..", "..", "testdata", "config", "valid.yaml"))

	err := loadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "/tmp/test-vault", cfg.Vault.Path)
}

func TestLoadConfig_VaultPathOverride(t *testing.T) {
	cfgFile = filepath.Join("..", "..", "testdata", "config", "valid.yaml")
	vaultPath = "/override/vault"
	defer func() {
		cfgFile = ""
		vaultPath = ""
	}()

	err := loadConfig()
	require.NoError(t, err)
	assert.Equal(t, "/override/vault", cfg.Vault.Path)
}

func TestLoadConfig_VaultEnvOverride(t *testing.T) {
	cfgFile = filepath.Join("..", "..", "testdata", "config", "valid.yaml")
	vaultPath = ""
	t.Setenv("RESUMECTL_VAULT", "/env/vault")
	defer func() { cfgFile = "" }()

	err := loadConfig()
	require.NoError(t, err)
	assert.Equal(t, "/env/vault", cfg.Vault.Path)
}

func TestLoadConfig_DefaultPathNoFile(t *testing.T) {
	cfgFile = ""
	// Ensure no env vars interfere
	t.Setenv("RESUMECTL_CONFIG", "")

	err := loadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	// Should get defaults since config file likely doesn't exist in test
	assert.NotEmpty(t, cfg.Vault.Path)
}

func TestLoadConfig_MalformedFile(t *testing.T) {
	cfgFile = filepath.Join("..", "..", "testdata", "config", "malformed.yaml")
	defer func() { cfgFile = "" }()

	err := loadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config: parse")
}
