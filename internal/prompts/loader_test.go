package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLibrary(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)
	assert.NotNil(t, lib)
}

func TestNewLibrary_AllTemplatesLoad(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	for _, name := range TemplateNames {
		tmpl, err := lib.Get(name)
		require.NoError(t, err, "template %q should load", name)
		assert.NotNil(t, tmpl)
	}
}

func TestGet_UnknownTemplate(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	_, err = lib.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

func TestNames(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	names := lib.Names()
	assert.Len(t, names, len(TemplateNames))
	for _, expected := range TemplateNames {
		assert.Contains(t, names, expected)
	}
}

func TestTemplateNames_Complete(t *testing.T) {
	expected := []string{
		"decompose_resume",
		"extract_job",
		"gap_analysis",
		"synthesize_resume",
		"synthesize_cover_letter",
		"coach_question",
	}
	assert.Equal(t, expected, TemplateNames)
}
