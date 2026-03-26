package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResumeFrontmatter holds the YAML frontmatter fields for a generated resume file.
type ResumeFrontmatter struct {
	JobFile         string   `yaml:"job_file"`
	Generated       string   `yaml:"generated"`
	Model           string   `yaml:"model"`
	Status          string   `yaml:"status"`
	ExperienceFiles []string `yaml:"experience_files,omitempty"`
	Version         int      `yaml:"version"`
}

// ResumeFile represents a loaded generated resume from the vault.
type ResumeFile struct {
	Frontmatter ResumeFrontmatter
	Body        string
	Path        string
	RelPath     string
}

// LoadResume reads and parses a resume file at the given vault-relative path.
func (v *Vault) LoadResume(relPath string) (*ResumeFile, error) {
	absPath := filepath.Join(v.Root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("vault: load resume %s: %w", relPath, err)
	}

	fm, body, err := ParseFrontmatter[ResumeFrontmatter](data)
	if err != nil {
		return nil, fmt.Errorf("vault: load resume %s: %w", relPath, err)
	}

	return &ResumeFile{
		Frontmatter: fm,
		Body:        body,
		Path:        absPath,
		RelPath:     relPath,
	}, nil
}

// WriteResume writes a resume file to the vault and optionally commits it.
func (v *Vault) WriteResume(rf *ResumeFile) error {
	if rf.RelPath == "" {
		return fmt.Errorf("vault: write resume: RelPath is required")
	}
	rf.Path = filepath.Join(v.Root, rf.RelPath)

	data, err := RenderFile(rf.Frontmatter, rf.Body)
	if err != nil {
		return fmt.Errorf("vault: write resume: %w", err)
	}

	dir := filepath.Dir(rf.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("vault: create resume dir: %w", err)
	}

	if err := os.WriteFile(rf.Path, data, 0o644); err != nil {
		return fmt.Errorf("vault: write resume %s: %w", rf.RelPath, err)
	}

	if v.git != nil && v.Config.Vault.Git.AutoCommit && v.Config.Vault.Git.CommitOnBuild {
		msg := fmt.Sprintf("resumectl: build %s", rf.RelPath)
		return v.git.commit([]string{rf.RelPath}, msg)
	}

	return nil
}
