package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsuykerbuyk/resumectl/internal/config"
	"github.com/jsuykerbuyk/resumectl/internal/llm"
	"github.com/jsuykerbuyk/resumectl/internal/prompts"
	"github.com/jsuykerbuyk/resumectl/internal/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "llm-responses", name))
	require.NoError(t, err)
	return string(data)
}

func setupTestDeps(t *testing.T, responses ...string) string {
	t.Helper()
	vaultRoot := setupTestVault(t)

	c := config.DefaultConfig()
	c.Vault.Path = vaultRoot
	c.Vault.Git.AutoCommit = false
	cfg = c

	v, err := vault.Open(vaultRoot, c)
	require.NoError(t, err)

	lib, err := prompts.NewLibrary()
	require.NoError(t, err)

	mock := llm.NewMockProvider(responses...)
	providers := map[string]llm.Provider{"mock": mock}
	taskMap := map[string]string{
		"decompose":         "mock",
		"extract_job":       "mock",
		"gap_analysis":      "mock",
		"synthesize_resume": "mock",
		"cover_letter":      "mock",
		"coach":             "mock",
	}

	deps = &Dependencies{
		Vault:   v,
		Router:  llm.NewRouter(providers, taskMap),
		Prompts: lib,
	}

	return vaultRoot
}

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func resetGlobals(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		cfg = nil
		deps = nil
		dryRun = false
		providerOvr = ""
	})
}

// --- Command tests ---

func TestAllCommandsRegistered(t *testing.T) {
	root := NewRootCmd()
	commands := make(map[string]bool)
	for _, cmd := range root.Commands() {
		commands[cmd.Name()] = true
	}

	for _, name := range []string{"version", "vault", "ingest", "parse-job", "match", "build", "coach", "export"} {
		assert.True(t, commands[name], "command %q should be registered", name)
	}
}

func TestAllCommandsHaveHelp(t *testing.T) {
	for _, name := range []string{"ingest", "parse-job", "match", "build", "coach", "export"} {
		t.Run(name, func(t *testing.T) {
			resetGlobals(t)
			out, err := runCmd(t, name, "--help")
			require.NoError(t, err)
			assert.NotEmpty(t, out)
		})
	}
}

func TestParseJobCmd(t *testing.T) {
	resetGlobals(t)
	fixture := loadFixture(t, "parsejob_response.json")
	vaultRoot := setupTestDeps(t, fixture)

	// Create a source file
	srcFile := filepath.Join(vaultRoot, "posting.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("We are hiring a Staff Engineer."), 0o644))

	out, err := runCmd(t, "parse-job", srcFile)
	require.NoError(t, err)
	assert.Contains(t, out, "Created")
	assert.Contains(t, out, "Acme Corp")
}

func TestParseJobCmd_DryRun(t *testing.T) {
	resetGlobals(t)
	fixture := loadFixture(t, "parsejob_response.json")
	vaultRoot := setupTestDeps(t, fixture)

	srcFile := filepath.Join(vaultRoot, "posting.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("Hiring."), 0o644))

	out, err := runCmd(t, "parse-job", "--dry-run", srcFile)
	require.NoError(t, err)
	assert.Contains(t, out, "[dry-run]")
	assert.Contains(t, out, "Acme Corp")
}

func TestMatchCmd(t *testing.T) {
	resetGlobals(t)
	fixture := loadFixture(t, "gap_analysis_response.json")
	setupTestDeps(t, fixture)

	out, err := runCmd(t, "match", "jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)
	assert.Contains(t, out, "Gap Analysis")
	assert.Contains(t, out, "78/100")
	assert.Contains(t, out, "Go")
}

func TestMatchCmd_JSON(t *testing.T) {
	resetGlobals(t)
	fixture := loadFixture(t, "gap_analysis_response.json")
	setupTestDeps(t, fixture)

	out, err := runCmd(t, "match", "--format", "json", "jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)
	assert.Contains(t, out, `"overall_score"`)
	assert.Contains(t, out, `"strong_matches"`)
}

func TestBuildCmd(t *testing.T) {
	resetGlobals(t)
	gapFixture := loadFixture(t, "gap_analysis_response.json")
	resumeFixture := loadFixture(t, "synthesize_resume_response.md")
	setupTestDeps(t, gapFixture, resumeFixture)

	out, err := runCmd(t, "build", "jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)
	assert.Contains(t, out, "Created")
}

func TestBuildCmd_DryRun(t *testing.T) {
	resetGlobals(t)
	gapFixture := loadFixture(t, "gap_analysis_response.json")
	resumeFixture := loadFixture(t, "synthesize_resume_response.md")
	setupTestDeps(t, gapFixture, resumeFixture)

	out, err := runCmd(t, "build", "--dry-run", "jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)
	assert.Contains(t, out, "[dry-run]")
	assert.Contains(t, out, "John Smith")
}

func TestIngestCmd(t *testing.T) {
	resetGlobals(t)
	// New role that won't be a duplicate
	fixture := `[{
		"role": "New Engineer",
		"company": "New Corp",
		"company_slug": "new-corp",
		"start": "2025-06",
		"end": "present",
		"current": true,
		"tags": ["new"],
		"skills": ["Go"],
		"domain": "software",
		"summary": "New role.",
		"contributions": ["Did stuff."],
		"technologies": ["Go"]
	}]`
	vaultRoot := setupTestDeps(t, fixture)

	srcFile := filepath.Join(vaultRoot, "resume.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("Old resume text."), 0o644))

	out, err := runCmd(t, "ingest", srcFile)
	require.NoError(t, err)
	assert.Contains(t, out, "Created")
	assert.Contains(t, out, "Ingested 1")
}

func TestIngestCmd_DryRun(t *testing.T) {
	resetGlobals(t)
	fixture := `[{"role": "Test", "company": "Test Co", "company_slug": "test-co",
		"start": "2025-01", "tags": [], "skills": [], "domain": "software",
		"summary": "Test.", "contributions": [], "technologies": []}]`
	vaultRoot := setupTestDeps(t, fixture)

	srcFile := filepath.Join(vaultRoot, "resume.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("Text."), 0o644))

	out, err := runCmd(t, "ingest", "--dry-run", srcFile)
	require.NoError(t, err)
	assert.Contains(t, out, "[dry-run]")
}

func TestExportCmd_DryRun(t *testing.T) {
	resetGlobals(t)
	vaultRoot := setupTestDeps(t)

	// Create a resume file in the test vault
	resumeDir := filepath.Join(vaultRoot, "resumes", "generated")
	require.NoError(t, os.MkdirAll(resumeDir, 0o755))
	resumeContent := "---\njob_file: jobs/target/test.md\ngenerated: \"2026-03-26\"\nmodel: test\nstatus: draft\nversion: 1\n---\n\n## Summary\n\nTest resume content.\n"
	require.NoError(t, os.WriteFile(filepath.Join(resumeDir, "test-resume.md"), []byte(resumeContent), 0o644))

	out, err := runCmd(t, "export", "--dry-run", "resumes/generated/test-resume.md")
	require.NoError(t, err)
	assert.Contains(t, out, "[dry-run]")
}

func TestCoachCmd_NoGaps(t *testing.T) {
	resetGlobals(t)
	// Gap analysis returns zero gaps
	noGaps := `{"overall_score": 0.95, "narrative": "Great match.", "strong_matches": [], "partial_matches": [], "gaps": []}`
	setupTestDeps(t, noGaps)

	out, err := runCmd(t, "coach", "jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)
	assert.Contains(t, out, "No gaps found")
}

func TestGetProvider_Override(t *testing.T) {
	resetGlobals(t)
	setupTestDeps(t)
	providerOvr = "mock"

	p, err := getProvider(llm.TaskDecompose)
	require.NoError(t, err)
	assert.Equal(t, "mock", p.Name())
}

func TestGetProvider_OverrideUnknown(t *testing.T) {
	resetGlobals(t)
	setupTestDeps(t)
	providerOvr = "nonexistent"

	_, err := getProvider(llm.TaskDecompose)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown provider")
}

func TestGetProvider_FromRouter(t *testing.T) {
	resetGlobals(t)
	setupTestDeps(t)
	providerOvr = ""

	p, err := getProvider(llm.TaskDecompose)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestInitDeps_AlreadyInitialized(t *testing.T) {
	resetGlobals(t)
	setupTestDeps(t)

	// initDeps should be a no-op since deps is already set
	err := initDeps()
	require.NoError(t, err)
}

func TestExtractSourceText_Plaintext(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(f, []byte("plain text content"), 0o644))

	text, err := extractSourceText(f)
	require.NoError(t, err)
	assert.Equal(t, "plain text content", text)
}

func TestExtractSourceText_Markdown(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.md")
	require.NoError(t, os.WriteFile(f, []byte("# Heading\n\nMarkdown content."), 0o644))

	text, err := extractSourceText(f)
	require.NoError(t, err)
	assert.Contains(t, text, "Markdown content")
}

func TestExtractSourceText_HTML(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.html")
	require.NoError(t, os.WriteFile(f, []byte("<html><body><p>HTML content</p></body></html>"), 0o644))

	text, err := extractSourceText(f)
	require.NoError(t, err)
	assert.Contains(t, text, "HTML content")
	assert.NotContains(t, text, "<html>")
}

func TestExtractSourceText_URL(t *testing.T) {
	_, err := extractSourceText("https://example.com/job")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL ingestion not supported")
}

func TestExtractSourceText_MissingFile(t *testing.T) {
	_, err := extractSourceText("/nonexistent/file.txt")
	require.Error(t, err)
}

func TestFetchJobText_Stdin(t *testing.T) {
	// We can't easily test stdin without piping, but we can test the file path
	tmp := t.TempDir()
	f := filepath.Join(tmp, "job.txt")
	require.NoError(t, os.WriteFile(f, []byte("Job posting text."), 0o644))

	cfg = config.DefaultConfig()
	text, url, err := fetchJobText(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, "Job posting text.", text)
	assert.Empty(t, url)
}

func TestMatchCmd_InvalidJobFile(t *testing.T) {
	resetGlobals(t)
	setupTestDeps(t)

	_, err := runCmd(t, "match", "nonexistent.md")
	require.Error(t, err)
}

func TestBuildCmd_WithCoverLetter(t *testing.T) {
	resetGlobals(t)
	gapFixture := loadFixture(t, "gap_analysis_response.json")
	resumeFixture := loadFixture(t, "synthesize_resume_response.md")
	coverLetter := "Dear Hiring Manager, I am excited to apply."
	setupTestDeps(t, gapFixture, resumeFixture, coverLetter)

	out, err := runCmd(t, "build", "--cover-letter", "jobs/target/2026-03-25-acme-staff-engineer.md")
	require.NoError(t, err)
	assert.Contains(t, out, "Cover Letter")
}
