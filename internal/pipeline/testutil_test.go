package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/config"
	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", name)
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(testdataPath("llm-responses/" + name))
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	return string(data)
}

func mockProvider(t *testing.T, fixtureName string) *llm.MockProvider {
	t.Helper()
	return llm.NewMockProvider(loadFixture(t, fixtureName))
}

func testLibrary(t *testing.T) *prompts.Library {
	t.Helper()
	lib, err := prompts.NewLibrary()
	if err != nil {
		t.Fatalf("load prompts: %v", err)
	}
	return lib
}

func openTestVault(t *testing.T) *vault.Vault {
	t.Helper()
	tmp := t.TempDir()
	src := testdataPath("vault")
	copyDir(t, src, tmp)

	cfg := config.DefaultConfig()
	cfg.Vault.Git.AutoCommit = false
	v, err := vault.Open(tmp, cfg)
	if err != nil {
		t.Fatalf("open test vault: %v", err)
	}
	return v
}

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
