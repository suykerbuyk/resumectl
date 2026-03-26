package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContactInfo holds the profile/contact.md frontmatter fields.
type ContactInfo struct {
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	Phone    string `yaml:"phone,omitempty"`
	Location string `yaml:"location,omitempty"`
	LinkedIn string `yaml:"linkedin,omitempty"`
	GitHub   string `yaml:"github,omitempty"`
	Website  string `yaml:"website,omitempty"`
}

// SummaryFrontmatter holds the profile/summary-core.md frontmatter fields.
type SummaryFrontmatter struct {
	Updated string   `yaml:"updated,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}

// ProfileSummary represents the professional identity narrative.
type ProfileSummary struct {
	Frontmatter SummaryFrontmatter
	Body        string
}

// SkillsFrontmatter holds the profile/skills.md frontmatter fields.
type SkillsFrontmatter struct {
	Updated string `yaml:"updated,omitempty"`
}

// SkillEntry represents a single row in the skills inventory table.
type SkillEntry struct {
	Skill       string
	Proficiency string
	LastUsed    string
	Years       string
	Category    string
}

// SkillsInventory represents the full skills file.
type SkillsInventory struct {
	Frontmatter SkillsFrontmatter
	Skills      []SkillEntry
	Body        string // raw body for cluster data and unstructured content
}

// LoadContact reads and parses the profile/contact.md file.
func (v *Vault) LoadContact() (*ContactInfo, error) {
	relPath := filepath.Join("profile", "contact.md")
	absPath := filepath.Join(v.Root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("vault: load contact: %w", err)
	}

	ci, _, err := ParseFrontmatter[ContactInfo](data)
	if err != nil {
		return nil, fmt.Errorf("vault: load contact: %w", err)
	}

	return &ci, nil
}

// LoadSummary reads and parses the profile/summary-core.md file.
func (v *Vault) LoadSummary() (*ProfileSummary, error) {
	relPath := filepath.Join("profile", "summary-core.md")
	absPath := filepath.Join(v.Root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("vault: load summary: %w", err)
	}

	fm, body, err := ParseFrontmatter[SummaryFrontmatter](data)
	if err != nil {
		return nil, fmt.Errorf("vault: load summary: %w", err)
	}

	return &ProfileSummary{
		Frontmatter: fm,
		Body:        body,
	}, nil
}

// LoadSkills reads and parses the profile/skills.md file, extracting the
// skills inventory table into structured SkillEntry values.
func (v *Vault) LoadSkills() (*SkillsInventory, error) {
	relPath := filepath.Join("profile", "skills.md")
	absPath := filepath.Join(v.Root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("vault: load skills: %w", err)
	}

	fm, body, err := ParseFrontmatter[SkillsFrontmatter](data)
	if err != nil {
		return nil, fmt.Errorf("vault: load skills: %w", err)
	}

	skills := parseSkillsTable(body)

	return &SkillsInventory{
		Frontmatter: fm,
		Skills:      skills,
		Body:        body,
	}, nil
}

// parseSkillsTable extracts SkillEntry values from a markdown table in the body.
func parseSkillsTable(body string) []SkillEntry {
	var skills []SkillEntry
	scanner := bufio.NewScanner(strings.NewReader(body))
	inTable := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "|") {
			if inTable {
				break // end of table
			}
			continue
		}

		// Skip header row and separator row
		if !inTable {
			inTable = true
			scanner.Scan() // skip separator row (|---|---|...)
			continue
		}

		cols := splitTableRow(line)
		if len(cols) >= 5 {
			skills = append(skills, SkillEntry{
				Skill:       cols[0],
				Proficiency: cols[1],
				LastUsed:    cols[2],
				Years:       cols[3],
				Category:    cols[4],
			})
		}
	}

	return skills
}

// splitTableRow splits a markdown table row into trimmed column values.
func splitTableRow(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		cols = append(cols, strings.TrimSpace(p))
	}
	return cols
}
