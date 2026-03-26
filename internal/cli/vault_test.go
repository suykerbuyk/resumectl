package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestVault(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	src := filepath.Join("..", "..", "testdata", "vault")
	copyDir(t, src, tmp)
	return tmp
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dst, 0o755))
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(t, srcPath, dstPath)
		} else {
			data, err := os.ReadFile(srcPath)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(dstPath, data, 0o644))
		}
	}
}

func runVaultCmd(t *testing.T, vaultRoot string, args ...string) (string, error) {
	t.Helper()
	// Set up global config
	cfg = config.DefaultConfig()
	cfg.Vault.Path = vaultRoot
	cfg.Vault.Git.AutoCommit = false

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestVaultInit(t *testing.T) {
	tmp := t.TempDir()
	initPath := filepath.Join(tmp, "new-vault")

	out, err := runVaultCmd(t, tmp, "vault", "init", initPath)
	require.NoError(t, err)
	assert.Contains(t, out, "Vault initialized")
	assert.DirExists(t, filepath.Join(initPath, "experience"))
	assert.DirExists(t, filepath.Join(initPath, "jobs", "target"))
	assert.DirExists(t, filepath.Join(initPath, "profile"))
	assert.DirExists(t, filepath.Join(initPath, "resumes", "generated"))
}

func TestVaultValidate_Valid(t *testing.T) {
	vaultRoot := setupTestVault(t)
	out, err := runVaultCmd(t, vaultRoot, "vault", "validate")
	require.NoError(t, err)
	assert.Contains(t, out, "valid")
}

func TestVaultValidate_Invalid(t *testing.T) {
	tmp := t.TempDir()
	// Only create experience/ dir — missing others
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "experience"), 0o755))

	_, err := runVaultCmd(t, tmp, "vault", "validate")
	require.Error(t, err)
}

func TestVaultLs(t *testing.T) {
	vaultRoot := setupTestVault(t)
	out, err := runVaultCmd(t, vaultRoot, "vault", "ls")
	require.NoError(t, err)
	assert.Contains(t, out, "seagate")
	assert.Contains(t, out, "Principal Storage Architect")
	assert.Contains(t, out, "Intel")
}

func TestVaultStats(t *testing.T) {
	vaultRoot := setupTestVault(t)
	out, err := runVaultCmd(t, vaultRoot, "vault", "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Experience files: 3")
	assert.Contains(t, out, "Job files:        1")
	assert.Contains(t, out, "Skills tracked:   5")
}

func TestVaultIndex(t *testing.T) {
	vaultRoot := setupTestVault(t)
	out, err := runVaultCmd(t, vaultRoot, "vault", "index")
	require.NoError(t, err)
	assert.Contains(t, out, "Tags:")
	assert.Contains(t, out, "Skills:")
	assert.Contains(t, out, "Domains:")
	assert.Contains(t, out, "storage")
}
