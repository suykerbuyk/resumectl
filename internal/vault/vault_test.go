package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_ValidVault(t *testing.T) {
	v := openTestVault(t)
	assert.NotEmpty(t, v.Root)
	assert.NotNil(t, v.Config)
	assert.Nil(t, v.Index) // lazy, not built yet
}

func TestOpen_NonexistentDir(t *testing.T) {
	cfg := defaultTestConfig()
	_, err := Open("/nonexistent/path", cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestValidate_ValidVault(t *testing.T) {
	v := openTestVault(t)
	missing := v.Validate()
	assert.Empty(t, missing)
}

func TestValidate_BrokenVault(t *testing.T) {
	tmp := t.TempDir()
	// Only create experience/ — missing jobs/target, profile, resumes/generated
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "experience"), 0o755))

	cfg := defaultTestConfig()
	v, err := Open(tmp, cfg)
	require.NoError(t, err)

	missing := v.Validate()
	assert.Len(t, missing, 3)
	assert.Contains(t, missing, "jobs/target")
	assert.Contains(t, missing, "profile")
	assert.Contains(t, missing, "resumes/generated")
}

func TestInit(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "new-vault")

	err := Init(root)
	require.NoError(t, err)

	for _, dir := range AllDirs {
		assert.DirExists(t, filepath.Join(root, dir))
	}
}

func TestIsGitRepo_NoGit(t *testing.T) {
	v := openTestVault(t)
	assert.False(t, v.IsGitRepo())
}

func TestIsGitRepo_WithGit(t *testing.T) {
	v := openTestVault(t)
	require.NoError(t, os.MkdirAll(filepath.Join(v.Root, ".git"), 0o755))

	// Re-open to detect .git
	cfg := defaultTestConfig()
	v2, err := Open(v.Root, cfg)
	require.NoError(t, err)
	assert.True(t, v2.IsGitRepo())
}
