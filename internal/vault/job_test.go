package vault

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadJob(t *testing.T) {
	v := openTestVault(t)
	jf, err := v.LoadJob("jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)

	assert.Equal(t, "Staff Software Engineer — Infrastructure", jf.Frontmatter.Title)
	assert.Equal(t, "Acme Corp", jf.Frontmatter.Company)
	assert.Equal(t, "acme-corp", jf.Frontmatter.CompanySlug)
	assert.Equal(t, "targeting", jf.Frontmatter.Status)
	assert.Equal(t, "staff", jf.Frontmatter.Seniority)
	assert.Contains(t, jf.Frontmatter.RequiredSkills, "Go")
	assert.Contains(t, jf.Frontmatter.RequiredSkills, "Kubernetes")
	assert.Contains(t, jf.Frontmatter.PreferredSkills, "Rust")
	assert.Contains(t, jf.Frontmatter.PreferredSkills, "NVMe")
	require.NotNil(t, jf.Frontmatter.Compensation)
	assert.Equal(t, 200000, jf.Frontmatter.Compensation.RangeLow)
	assert.Equal(t, 260000, jf.Frontmatter.Compensation.RangeHigh)
	assert.True(t, jf.Frontmatter.Compensation.Equity)
	assert.Contains(t, jf.Body, "distributed storage systems")
}

func TestLoadJob_NotFound(t *testing.T) {
	v := openTestVault(t)
	_, err := v.LoadJob("jobs/target/nonexistent.md")
	require.Error(t, err)
}

func TestAllJobs(t *testing.T) {
	v := openTestVault(t)
	jobs, err := v.AllJobs()
	require.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "acme-corp", jobs[0].Frontmatter.CompanySlug)
}

func TestWriteJob(t *testing.T) {
	v := openTestVault(t)
	jf := &JobFile{
		Frontmatter: JobFrontmatter{
			Title:          "Test Role",
			Company:        "Test Inc",
			CompanySlug:    "test-inc",
			DateParsed:     "2026-03-26",
			Status:         "targeting",
			RequiredSkills: []string{"Go", "Rust"},
		},
		Body: "## Job Description\n\nTest job posting.\n",
	}

	err := v.WriteJob(jf)
	require.NoError(t, err)
	assert.NotEmpty(t, jf.RelPath)
	assert.FileExists(t, jf.Path)

	loaded, err := v.LoadJob(jf.RelPath)
	require.NoError(t, err)
	assert.Equal(t, "Test Role", loaded.Frontmatter.Title)
	assert.Equal(t, []string{"Go", "Rust"}, loaded.Frontmatter.RequiredSkills)
}

func TestJobFilename(t *testing.T) {
	name := JobFilename("2026-03-25", "acme-corp", "Staff Software Engineer")
	assert.Equal(t, filepath.Join("jobs", "target", "2026-03-25-acme-corp-staff-software-engineer.md"), name)
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"targeting", true},
		{"applied", true},
		{"interviewing", true},
		{"closed", true},
		{"rejected", true},
		{"offer", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			fm := &JobFrontmatter{Status: tt.status}
			err := fm.ValidateStatus()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid job status")
			}
		})
	}
}
