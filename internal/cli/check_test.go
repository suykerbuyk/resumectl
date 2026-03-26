package cli

import (
	"bytes"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCmd_FullPass(t *testing.T) {
	resetGlobals(t)
	vaultRoot := setupTestVault(t)

	c := config.DefaultConfig()
	c.Vault.Path = vaultRoot
	c.Providers.Claude.APIKey = "test-key"
	c.Providers.OpenAI.APIKey = "test-key"
	c.Providers.Gemini.APIKey = "test-key"
	c.Providers.Grok.APIKey = "test-key"
	cfg = c

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	err := root.Execute()

	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "All checks passed")
	assert.Contains(t, out, "profile/contact.md")
	assert.Contains(t, out, "claude")
}

func TestCheckCmd_MissingVault(t *testing.T) {
	resetGlobals(t)
	c := config.DefaultConfig()
	c.Vault.Path = "/nonexistent/vault"
	c.Providers.Claude.APIKey = "test-key"
	cfg = c

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	err := root.Execute()

	require.Error(t, err)
	out := buf.String()
	assert.Contains(t, out, "does not exist")
	assert.Contains(t, out, "resumectl vault init")
}

func TestCheckCmd_MissingProviderKey(t *testing.T) {
	resetGlobals(t)
	vaultRoot := setupTestVault(t)

	c := config.DefaultConfig()
	c.Vault.Path = vaultRoot
	// Only set claude key — gemini is default for extract_job but has no key
	c.Providers.Claude.APIKey = "test-key"
	cfg = c

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	err := root.Execute()

	require.Error(t, err)
	out := buf.String()
	assert.Contains(t, out, "gemini")
	assert.Contains(t, out, "extract_job")
	assert.Contains(t, out, "no API key")
}

func TestCheckCmd_NoProviderKeys(t *testing.T) {
	resetGlobals(t)
	vaultRoot := setupTestVault(t)

	c := config.DefaultConfig()
	c.Vault.Path = vaultRoot
	// No API keys set at all
	cfg = c

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	err := root.Execute()

	require.Error(t, err)
	out := buf.String()
	assert.Contains(t, out, "no provider API keys configured")
}

func TestCheckCmd_MissingVaultDirs(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir() // empty dir, no subdirs

	c := config.DefaultConfig()
	c.Vault.Path = tmp
	c.Providers.Claude.APIKey = "test-key"
	cfg = c

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"check"})
	err := root.Execute()

	require.Error(t, err)
	out := buf.String()
	assert.Contains(t, out, "missing required directories")
}

func TestCheckCmd_Templates(t *testing.T) {
	resetGlobals(t)
	vaultRoot := setupTestVault(t)

	c := config.DefaultConfig()
	c.Vault.Path = vaultRoot
	c.Providers.Claude.APIKey = "k"
	c.Providers.Gemini.APIKey = "k"
	cfg = c

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"check"})
	_ = root.Execute()

	assert.Contains(t, buf.String(), "6 prompt templates loaded")
}
