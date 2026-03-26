package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JobFrontmatter holds the YAML frontmatter fields for a job posting file.
type JobFrontmatter struct {
	Title           string        `yaml:"title"`
	Company         string        `yaml:"company"`
	CompanySlug     string        `yaml:"company_slug"`
	SourceURL       string        `yaml:"source_url,omitempty"`
	DateParsed      string        `yaml:"date_parsed"`
	Status          string        `yaml:"status"`
	Seniority       string        `yaml:"seniority,omitempty"`
	Domain          string        `yaml:"domain,omitempty"`
	RequiredSkills  []string      `yaml:"required_skills,omitempty"`
	PreferredSkills []string      `yaml:"preferred_skills,omitempty"`
	CultureSignals  []string      `yaml:"culture_signals,omitempty"`
	Compensation    *Compensation `yaml:"compensation,omitempty"`
	Location        string        `yaml:"location,omitempty"`
	Tags            []string      `yaml:"tags,omitempty"`
}

// Compensation holds salary/equity information from a job posting.
type Compensation struct {
	RangeLow  int    `yaml:"range_low,omitempty" json:"range_low,omitempty"`
	RangeHigh int    `yaml:"range_high,omitempty" json:"range_high,omitempty"`
	Currency  string `yaml:"currency,omitempty" json:"currency,omitempty"`
	Equity    bool   `yaml:"equity" json:"equity"`
}

// JobFile represents a loaded job posting file from the vault.
type JobFile struct {
	Frontmatter JobFrontmatter
	Body        string
	Path        string
	RelPath     string
}

// LoadJob reads and parses a job file at the given vault-relative path.
func (v *Vault) LoadJob(relPath string) (*JobFile, error) {
	absPath := filepath.Join(v.Root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("vault: load job %s: %w", relPath, err)
	}

	fm, body, err := ParseFrontmatter[JobFrontmatter](data)
	if err != nil {
		return nil, fmt.Errorf("vault: load job %s: %w", relPath, err)
	}

	return &JobFile{
		Frontmatter: fm,
		Body:        body,
		Path:        absPath,
		RelPath:     relPath,
	}, nil
}

// AllJobs reads and parses all job files in the vault's jobs/target/ directory.
func (v *Vault) AllJobs() ([]*JobFile, error) {
	dir := filepath.Join(v.Root, "jobs", "target")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("vault: read jobs dir: %w", err)
	}

	var files []*JobFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		relPath := filepath.Join("jobs", "target", entry.Name())
		jf, err := v.LoadJob(relPath)
		if err != nil {
			return nil, err
		}
		files = append(files, jf)
	}

	return files, nil
}

// WriteJob writes a job file to the vault and optionally commits it.
func (v *Vault) WriteJob(jf *JobFile) error {
	if jf.RelPath == "" {
		jf.RelPath = JobFilename(jf.Frontmatter.DateParsed, jf.Frontmatter.CompanySlug, jf.Frontmatter.Title)
	}
	jf.Path = filepath.Join(v.Root, jf.RelPath)

	data, err := RenderFile(jf.Frontmatter, jf.Body)
	if err != nil {
		return fmt.Errorf("vault: write job: %w", err)
	}

	dir := filepath.Dir(jf.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("vault: create job dir: %w", err)
	}

	if err := os.WriteFile(jf.Path, data, 0o644); err != nil {
		return fmt.Errorf("vault: write job %s: %w", jf.RelPath, err)
	}

	if v.git != nil && v.Config.Vault.Git.AutoCommit {
		msg := fmt.Sprintf("resumectl: parse-job %s", jf.Frontmatter.CompanySlug)
		return v.git.commit([]string{jf.RelPath}, msg)
	}

	return nil
}

// JobFilename generates the canonical filename for a job file.
func JobFilename(dateParsed, companySlug, title string) string {
	titleSlug := slugify(title)
	return filepath.Join("jobs", "target", fmt.Sprintf("%s-%s-%s.md", dateParsed, companySlug, titleSlug))
}

// ValidJobStatuses are the allowed values for JobFrontmatter.Status.
var ValidJobStatuses = []string{
	"targeting", "applied", "interviewing", "closed", "rejected", "offer",
}

// ValidateStatus checks if the job status is one of the allowed values.
func (jf *JobFrontmatter) ValidateStatus() error {
	for _, s := range ValidJobStatuses {
		if jf.Status == s {
			return nil
		}
	}
	return fmt.Errorf("vault: invalid job status %q (must be one of %v)", jf.Status, ValidJobStatuses)
}
