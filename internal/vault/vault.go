package vault

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jsuykerbuyk/resumectl/internal/config"
)

// RequiredDirs are the directories that must exist in a valid vault.
var RequiredDirs = []string{
	"experience",
	"jobs/target",
	"profile",
	"resumes/generated",
}

// AllDirs includes all directories created by vault init (superset of RequiredDirs).
var AllDirs = []string{
	"experience",
	"jobs/target",
	"profile",
	"resumes/generated",
	"cover-letters",
	"projects",
	"sessions",
}

// Vault represents an open Obsidian vault for resumectl.
type Vault struct {
	Root   string
	Config *config.Config
	Index  *TagIndex
	git    *gitClient
}

// Open opens a vault at the given root directory using the provided config.
func Open(root string, cfg *config.Config) (*Vault, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("vault: resolve path %s: %w", root, err)
	}

	if !dirExists(absRoot) {
		return nil, fmt.Errorf("vault: directory does not exist: %s", absRoot)
	}

	v := &Vault{
		Root:   absRoot,
		Config: cfg,
		git:    newGitClient(absRoot, nil),
	}

	return v, nil
}

// Validate checks that all required directories exist in the vault.
func (v *Vault) Validate() []string {
	var missing []string
	for _, dir := range RequiredDirs {
		path := filepath.Join(v.Root, dir)
		if !dirExists(path) {
			missing = append(missing, dir)
		}
	}
	return missing
}

// RebuildIndex forces a rebuild of the tag index from all experience files.
func (v *Vault) RebuildIndex() error {
	files, err := v.AllExperience()
	if err != nil {
		return fmt.Errorf("vault: rebuild index: %w", err)
	}
	v.Index = BuildTagIndex(files)
	return nil
}

// EnsureIndex builds the tag index if it hasn't been built yet.
func (v *Vault) EnsureIndex() error {
	if v.Index != nil {
		return nil
	}
	return v.RebuildIndex()
}

// IsGitRepo returns true if the vault root contains a .git directory.
func (v *Vault) IsGitRepo() bool {
	return v.git != nil
}

// Init creates the vault directory structure at the given root path.
func Init(root string) error {
	for _, dir := range AllDirs {
		path := filepath.Join(root, dir)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("vault: init create %s: %w", dir, err)
		}
	}
	return nil
}

// dirExists checks if a directory exists at the given path.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
