package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/config"
)

// copyDir recursively copies src to dst.
func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read dir %s: %v", src, err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dst, err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(t, srcPath, dstPath)
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatalf("read %s: %v", srcPath, err)
			}
			if err := os.WriteFile(dstPath, data, 0o644); err != nil {
				t.Fatalf("write %s: %v", dstPath, err)
			}
		}
	}
}

// openTestVault copies the testdata/vault fixture to a temp dir and opens it.
func openTestVault(t *testing.T) *Vault {
	t.Helper()
	tmp := t.TempDir()
	src := filepath.Join("..", "..", "testdata", "vault")
	copyDir(t, src, tmp)

	cfg := config.DefaultConfig()
	cfg.Vault.Git.AutoCommit = false // disable git for most tests
	v, err := Open(tmp, cfg)
	if err != nil {
		t.Fatalf("open test vault: %v", err)
	}
	return v
}

// mockCommander records commands for assertion in tests.
type mockCommander struct {
	calls [][]string
	err   error // if set, all calls return this error
}

func (m *mockCommander) Run(name string, args ...string) error {
	call := append([]string{name}, args...)
	m.calls = append(m.calls, call)
	return m.err
}
