package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadContact(t *testing.T) {
	v := openTestVault(t)
	ci, err := v.LoadContact()
	require.NoError(t, err)

	assert.Equal(t, "John Smith", ci.Name)
	assert.Equal(t, "john@example.com", ci.Email)
	assert.Equal(t, "+1 555 000 0000", ci.Phone)
	assert.Equal(t, "Loveland, CO", ci.Location)
	assert.Equal(t, "https://linkedin.com/in/johnsmith", ci.LinkedIn)
	assert.Equal(t, "https://github.com/johnsmith", ci.GitHub)
	assert.Equal(t, "https://johnsmith.dev", ci.Website)
}

func TestLoadContact_Missing(t *testing.T) {
	tmp := t.TempDir()
	v := &Vault{Root: tmp, Config: defaultTestConfig()}
	_, err := v.LoadContact()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault: load contact")
}

func TestLoadSummary(t *testing.T) {
	v := openTestVault(t)
	ps, err := v.LoadSummary()
	require.NoError(t, err)

	assert.Equal(t, "2026-03-25", ps.Frontmatter.Updated)
	assert.Contains(t, ps.Frontmatter.Tags, "summary")
	assert.Contains(t, ps.Body, "Seasoned systems engineer")
}

func TestLoadSummary_Missing(t *testing.T) {
	tmp := t.TempDir()
	v := &Vault{Root: tmp, Config: defaultTestConfig()}
	_, err := v.LoadSummary()
	require.Error(t, err)
}

func TestLoadSkills(t *testing.T) {
	v := openTestVault(t)
	si, err := v.LoadSkills()
	require.NoError(t, err)

	assert.Equal(t, "2026-03-25", si.Frontmatter.Updated)
	assert.Len(t, si.Skills, 5)

	// Check first skill
	assert.Equal(t, "NVMe", si.Skills[0].Skill)
	assert.Equal(t, "Expert", si.Skills[0].Proficiency)
	assert.Equal(t, "2024", si.Skills[0].LastUsed)
	assert.Equal(t, "8", si.Skills[0].Years)
	assert.Equal(t, "Storage", si.Skills[0].Category)

	// Check Go skill
	assert.Equal(t, "Go", si.Skills[1].Skill)
	assert.Equal(t, "Advanced", si.Skills[1].Proficiency)

	// Body should contain cluster data
	assert.Contains(t, si.Body, "Skill Clusters")
}

func TestLoadSkills_Missing(t *testing.T) {
	tmp := t.TempDir()
	v := &Vault{Root: tmp, Config: defaultTestConfig()}
	_, err := v.LoadSkills()
	require.Error(t, err)
}

func TestParseSkillsTable_Empty(t *testing.T) {
	skills := parseSkillsTable("No table here.\n")
	assert.Empty(t, skills)
}

func TestSplitTableRow(t *testing.T) {
	cols := splitTableRow("| NVMe | Expert | 2024 | 8 | Storage |")
	assert.Equal(t, []string{"NVMe", "Expert", "2024", "8", "Storage"}, cols)
}
