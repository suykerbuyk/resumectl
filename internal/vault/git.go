package vault

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// Commander abstracts command execution for testability.
type Commander interface {
	Run(name string, args ...string) error
}

// execCommander is the real implementation that calls os/exec.
type execCommander struct{}

func (execCommander) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", name, err, string(out))
	}
	return nil
}

// gitClient handles Git auto-commit operations on the vault.
type gitClient struct {
	root string
	cmd  Commander
}

// newGitClient creates a gitClient for the given vault root.
// Returns nil if the root is not a Git repository.
func newGitClient(root string, cmd Commander) *gitClient {
	gitDir := filepath.Join(root, ".git")
	if !dirExists(gitDir) {
		return nil
	}
	if cmd == nil {
		cmd = execCommander{}
	}
	return &gitClient{root: root, cmd: cmd}
}

// commit stages the given vault-relative paths and commits with the message.
func (g *gitClient) commit(paths []string, message string) error {
	if g == nil {
		return nil
	}

	addArgs := append([]string{"-C", g.root, "add", "--"}, paths...)
	if err := g.cmd.Run("git", addArgs...); err != nil {
		return fmt.Errorf("vault: git add: %w", err)
	}

	if err := g.cmd.Run("git", "-C", g.root, "commit", "-m", message); err != nil {
		return fmt.Errorf("vault: git commit: %w", err)
	}

	return nil
}
