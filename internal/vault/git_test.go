package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitClient_Commit(t *testing.T) {
	mc := &mockCommander{}
	gc := &gitClient{root: "/test/vault", cmd: mc}

	err := gc.commit([]string{"experience/test.md"}, "resumectl: test commit")
	require.NoError(t, err)

	require.Len(t, mc.calls, 2)
	// First call: git add
	assert.Equal(t, "git", mc.calls[0][0])
	assert.Contains(t, mc.calls[0], "add")
	assert.Contains(t, mc.calls[0], "experience/test.md")
	// Second call: git commit
	assert.Equal(t, "git", mc.calls[1][0])
	assert.Contains(t, mc.calls[1], "commit")
	assert.Contains(t, mc.calls[1], "resumectl: test commit")
}

func TestGitClient_CommitMultiplePaths(t *testing.T) {
	mc := &mockCommander{}
	gc := &gitClient{root: "/test", cmd: mc}

	err := gc.commit([]string{"a.md", "b.md"}, "multi")
	require.NoError(t, err)

	assert.Contains(t, mc.calls[0], "a.md")
	assert.Contains(t, mc.calls[0], "b.md")
}

func TestGitClient_NilIsNoop(t *testing.T) {
	var gc *gitClient
	err := gc.commit([]string{"test.md"}, "should not fail")
	assert.NoError(t, err)
}

func TestGitClient_AddError(t *testing.T) {
	mc := &mockCommander{err: fmt.Errorf("git add failed")}
	gc := &gitClient{root: "/test", cmd: mc}

	err := gc.commit([]string{"test.md"}, "msg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault: git add")
}

func TestNewGitClient_NoGitDir(t *testing.T) {
	tmp := t.TempDir()
	gc := newGitClient(tmp, nil)
	assert.Nil(t, gc)
}

func TestNewGitClient_WithGitDir(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, ".git"), 0o755))

	gc := newGitClient(tmp, nil)
	assert.NotNil(t, gc)
	assert.Equal(t, tmp, gc.root)
}
