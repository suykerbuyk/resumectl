package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAndLoadResume(t *testing.T) {
	v := openTestVault(t)
	rf := &ResumeFile{
		Frontmatter: ResumeFrontmatter{
			JobFile:   "jobs/target/2026-03-25-acme-staff-engineer.md",
			Generated: "2026-03-25",
			Model:     "claude-opus-4",
			Status:    "draft",
			ExperienceFiles: []string{
				"experience/2021-03-seagate-principal-architect.md",
				"experience/2023-01-current-role.md",
			},
			Version: 1,
		},
		Body:    "# John Smith\n\n## Summary\n\nExperienced infrastructure engineer.\n",
		RelPath: "resumes/generated/2026-03-25-acme-staff-engineer.md",
	}

	err := v.WriteResume(rf)
	require.NoError(t, err)
	assert.FileExists(t, rf.Path)

	loaded, err := v.LoadResume(rf.RelPath)
	require.NoError(t, err)
	assert.Equal(t, "draft", loaded.Frontmatter.Status)
	assert.Equal(t, 1, loaded.Frontmatter.Version)
	assert.Equal(t, "claude-opus-4", loaded.Frontmatter.Model)
	assert.Len(t, loaded.Frontmatter.ExperienceFiles, 2)
	assert.Contains(t, loaded.Body, "infrastructure engineer")
}

func TestWriteResume_MissingRelPath(t *testing.T) {
	v := openTestVault(t)
	rf := &ResumeFile{
		Frontmatter: ResumeFrontmatter{Status: "draft"},
		Body:        "content",
	}

	err := v.WriteResume(rf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RelPath is required")
}

func TestLoadResume_NotFound(t *testing.T) {
	v := openTestVault(t)
	_, err := v.LoadResume("resumes/generated/nonexistent.md")
	require.Error(t, err)
}

func TestWriteResume_VersionTracking(t *testing.T) {
	v := openTestVault(t)

	// Write version 1
	rf := &ResumeFile{
		Frontmatter: ResumeFrontmatter{
			JobFile:   "jobs/target/test.md",
			Generated: "2026-03-25",
			Status:    "draft",
			Version:   1,
		},
		Body:    "Version 1 content.\n",
		RelPath: "resumes/generated/test-resume.md",
	}
	require.NoError(t, v.WriteResume(rf))

	// Write version 2 to same path
	rf.Frontmatter.Version = 2
	rf.Body = "Version 2 content.\n"
	require.NoError(t, v.WriteResume(rf))

	loaded, err := v.LoadResume(rf.RelPath)
	require.NoError(t, err)
	assert.Equal(t, 2, loaded.Frontmatter.Version)
	assert.Contains(t, loaded.Body, "Version 2")
}
