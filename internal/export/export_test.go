package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"docx", FormatDOCX},
		{"DOCX", FormatDOCX},
		{"pdf", FormatPDF},
		{"PDF", FormatPDF},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			f, err := ParseFormat(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, f)
		})
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := ParseFormat("html")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format")
}

func TestRenderDOCX_NoPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err == nil {
		t.Skip("pandoc is available")
	}
	err := RenderDOCX("in.md", "out.docx", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pandoc not found")
}

func TestRenderDOCX_WithPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not available")
	}

	tmp := t.TempDir()
	mdPath := filepath.Join(tmp, "test.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Test\n\nHello world.\n"), 0o644))

	outPath := filepath.Join(tmp, "test.docx")
	err := RenderDOCX(mdPath, outPath, "")
	require.NoError(t, err)
	assert.FileExists(t, outPath)
}

func TestRenderDOCX_WithTemplate(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not available")
	}

	tmp := t.TempDir()
	mdPath := filepath.Join(tmp, "test.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Test\n\nContent.\n"), 0o644))

	err := RenderDOCX(mdPath, filepath.Join(tmp, "out.docx"), "/nonexistent/template.docx")
	require.Error(t, err)
}

func TestRenderPDF_NoPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err == nil {
		t.Skip("pandoc is available")
	}
	err := RenderPDF("in.md", "out.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pandoc not found")
}

func TestExport_Dispatcher(t *testing.T) {
	err := Export("in.md", "out.xyz", Format("xyz"), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestExport_DOCX_NoPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err == nil {
		t.Skip("pandoc is available")
	}
	err := Export("in.md", "out.docx", FormatDOCX, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pandoc not found")
}

func TestExport_PDF_NoPandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err == nil {
		t.Skip("pandoc is available")
	}
	err := Export("in.md", "out.pdf", FormatPDF, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pandoc not found")
}
