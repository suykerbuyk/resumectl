package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExperienceFrontmatter holds the YAML frontmatter fields for an experience file.
type ExperienceFrontmatter struct {
	Role           string   `yaml:"role"`
	Company        string   `yaml:"company"`
	CompanySlug    string   `yaml:"company_slug"`
	Start          string   `yaml:"start"`
	End            string   `yaml:"end,omitempty"`
	Current        bool     `yaml:"current"`
	Location       string   `yaml:"location,omitempty"`
	EmploymentType string   `yaml:"employment_type,omitempty"`
	Tags           []string `yaml:"tags,omitempty"`
	Skills         []string `yaml:"skills,omitempty"`
	Domain         string   `yaml:"domain,omitempty"`
	Highlight      bool     `yaml:"highlight"`
	Visibility     string   `yaml:"visibility,omitempty"`
	Created        string   `yaml:"created,omitempty"`
	Updated        string   `yaml:"updated,omitempty"`
}

// ExperienceFile represents a loaded experience file from the vault.
type ExperienceFile struct {
	Frontmatter ExperienceFrontmatter
	Body        string // markdown body after frontmatter
	Path        string // absolute path on disk
	RelPath     string // vault-relative path
}

// LoadExperience reads and parses an experience file at the given vault-relative path.
func (v *Vault) LoadExperience(relPath string) (*ExperienceFile, error) {
	absPath := filepath.Join(v.Root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("vault: load experience %s: %w", relPath, err)
	}

	fm, body, err := ParseFrontmatter[ExperienceFrontmatter](data)
	if err != nil {
		return nil, fmt.Errorf("vault: load experience %s: %w", relPath, err)
	}

	return &ExperienceFile{
		Frontmatter: fm,
		Body:        body,
		Path:        absPath,
		RelPath:     relPath,
	}, nil
}

// AllExperience reads and parses all experience files in the vault's experience/ directory.
func (v *Vault) AllExperience() ([]*ExperienceFile, error) {
	dir := filepath.Join(v.Root, "experience")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("vault: read experience dir: %w", err)
	}

	var files []*ExperienceFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		relPath := filepath.Join("experience", entry.Name())
		ef, err := v.LoadExperience(relPath)
		if err != nil {
			return nil, err
		}
		files = append(files, ef)
	}

	return files, nil
}

// WriteExperience writes an experience file to the vault and optionally commits it.
func (v *Vault) WriteExperience(ef *ExperienceFile) error {
	if ef.RelPath == "" {
		ef.RelPath = ExperienceFilename(ef.Frontmatter.Start, ef.Frontmatter.CompanySlug, ef.Frontmatter.Role)
	}
	ef.Path = filepath.Join(v.Root, ef.RelPath)

	data, err := RenderFile(ef.Frontmatter, ef.Body)
	if err != nil {
		return fmt.Errorf("vault: write experience: %w", err)
	}

	dir := filepath.Dir(ef.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("vault: create experience dir: %w", err)
	}

	if err := os.WriteFile(ef.Path, data, 0o644); err != nil {
		return fmt.Errorf("vault: write experience %s: %w", ef.RelPath, err)
	}

	if v.git != nil && v.Config.Vault.Git.AutoCommit {
		msg := fmt.Sprintf("resumectl: add experience %s at %s", ef.Frontmatter.Company, ef.Frontmatter.Role)
		return v.git.commit([]string{ef.RelPath}, msg)
	}

	return nil
}

// ExperienceFilename generates the canonical filename for an experience file.
func ExperienceFilename(start, companySlug, role string) string {
	roleSlug := slugify(role)
	return filepath.Join("experience", fmt.Sprintf("%s-%s-%s.md", start, companySlug, roleSlug))
}

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
