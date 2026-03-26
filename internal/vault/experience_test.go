package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadExperience(t *testing.T) {
	v := openTestVault(t)
	ef, err := v.LoadExperience("experience/2021-03-seagate-principal-architect.md")
	require.NoError(t, err)

	assert.Equal(t, "Principal Storage Architect", ef.Frontmatter.Role)
	assert.Equal(t, "Seagate Technology", ef.Frontmatter.Company)
	assert.Equal(t, "seagate", ef.Frontmatter.CompanySlug)
	assert.Equal(t, "2021-03", ef.Frontmatter.Start)
	assert.Equal(t, "2024-06", ef.Frontmatter.End)
	assert.False(t, ef.Frontmatter.Current)
	assert.Contains(t, ef.Frontmatter.Tags, "storage")
	assert.Contains(t, ef.Frontmatter.Skills, "NVMe")
	assert.Equal(t, "storage", ef.Frontmatter.Domain)
	assert.True(t, ef.Frontmatter.Highlight)
	assert.Equal(t, "resume", ef.Frontmatter.Visibility)
	assert.Contains(t, ef.Body, "NVMe storage platform")
}

func TestLoadExperience_CurrentRole(t *testing.T) {
	v := openTestVault(t)
	ef, err := v.LoadExperience("experience/2023-01-current-role.md")
	require.NoError(t, err)

	assert.True(t, ef.Frontmatter.Current)
	assert.Equal(t, "", ef.Frontmatter.End)
}

func TestLoadExperience_NotFound(t *testing.T) {
	v := openTestVault(t)
	_, err := v.LoadExperience("experience/nonexistent.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault: load experience")
}

func TestAllExperience(t *testing.T) {
	v := openTestVault(t)
	files, err := v.AllExperience()
	require.NoError(t, err)
	assert.Len(t, files, 3)

	companies := make(map[string]bool)
	for _, f := range files {
		companies[f.Frontmatter.CompanySlug] = true
	}
	assert.True(t, companies["seagate"])
	assert.True(t, companies["intel"])
	assert.True(t, companies["cloudscale"])
}

func TestWriteExperience(t *testing.T) {
	v := openTestVault(t)
	ef := &ExperienceFile{
		Frontmatter: ExperienceFrontmatter{
			Role:        "Test Engineer",
			Company:     "Test Corp",
			CompanySlug: "test-corp",
			Start:       "2025-01",
			End:         "2025-12",
			Tags:        []string{"testing"},
			Skills:      []string{"Go"},
			Domain:      "testing",
			Visibility:  "resume",
		},
		Body: "## Summary\n\nTest role.\n",
	}

	err := v.WriteExperience(ef)
	require.NoError(t, err)

	assert.NotEmpty(t, ef.RelPath)
	assert.FileExists(t, ef.Path)

	// Roundtrip: load what we just wrote
	loaded, err := v.LoadExperience(ef.RelPath)
	require.NoError(t, err)
	assert.Equal(t, "Test Engineer", loaded.Frontmatter.Role)
	assert.Equal(t, "Test Corp", loaded.Frontmatter.Company)
	assert.Contains(t, loaded.Body, "Test role.")
}

func TestWriteExperience_WithGit(t *testing.T) {
	v := openTestVault(t)
	mc := &mockCommander{}
	v.git = &gitClient{root: v.Root, cmd: mc}
	v.Config.Vault.Git.AutoCommit = true

	ef := &ExperienceFile{
		Frontmatter: ExperienceFrontmatter{
			Role:        "Git Test",
			Company:     "Git Corp",
			CompanySlug: "git-corp",
			Start:       "2025-06",
			Skills:      []string{"Git"},
		},
		Body: "## Summary\n\nGit test.\n",
	}

	err := v.WriteExperience(ef)
	require.NoError(t, err)
	assert.Len(t, mc.calls, 2) // git add + git commit
}

func TestExperienceFilename(t *testing.T) {
	name := ExperienceFilename("2021-03", "seagate", "Principal Storage Architect")
	assert.Equal(t, filepath.Join("experience", "2021-03-seagate-principal-storage-architect.md"), name)
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Principal Storage Architect", "principal-storage-architect"},
		{"Staff Engineer — Infrastructure", "staff-engineer-infrastructure"},
		{"Go/C++ Developer", "goc-developer"},
		{"  spaces  ", "spaces"},
		{"already-slugified", "already-slugified"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, slugify(tt.input))
		})
	}
}

func TestWriteExperience_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	// Don't create experience/ dir — WriteExperience should create it
	v := &Vault{Root: tmp, Config: defaultTestConfig()}

	ef := &ExperienceFile{
		Frontmatter: ExperienceFrontmatter{
			Role:        "Test",
			Company:     "Test",
			CompanySlug: "test",
			Start:       "2025-01",
		},
		Body: "Test body.\n",
	}

	err := v.WriteExperience(ef)
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(tmp, "experience"))
	assert.FileExists(t, ef.Path)
}

func TestAllExperience_SkipsNonMd(t *testing.T) {
	v := openTestVault(t)
	// Create a non-.md file
	err := os.WriteFile(filepath.Join(v.Root, "experience", "notes.txt"), []byte("not markdown"), 0o644)
	require.NoError(t, err)

	files, err := v.AllExperience()
	require.NoError(t, err)
	assert.Len(t, files, 3) // still only the 3 .md files
}

func defaultTestConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Vault.Git.AutoCommit = false
	return cfg
}
