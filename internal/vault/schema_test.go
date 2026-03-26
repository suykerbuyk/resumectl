package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testFrontmatter struct {
	Name  string   `yaml:"name"`
	Tags  []string `yaml:"tags,omitempty"`
	Count int      `yaml:"count,omitempty"`
}

func TestParseFrontmatter_Valid(t *testing.T) {
	input := []byte("---\nname: test\ntags:\n  - a\n  - b\ncount: 42\n---\n\nBody content here.\n")
	fm, body, err := ParseFrontmatter[testFrontmatter](input)
	require.NoError(t, err)
	assert.Equal(t, "test", fm.Name)
	assert.Equal(t, []string{"a", "b"}, fm.Tags)
	assert.Equal(t, 42, fm.Count)
	assert.Equal(t, "Body content here.\n", body)
}

func TestParseFrontmatter_EmptyBody(t *testing.T) {
	input := []byte("---\nname: test\n---\n")
	fm, body, err := ParseFrontmatter[testFrontmatter](input)
	require.NoError(t, err)
	assert.Equal(t, "test", fm.Name)
	assert.Equal(t, "", body)
}

func TestParseFrontmatter_MissingOpenDelimiter(t *testing.T) {
	input := []byte("name: test\n---\n")
	_, _, err := ParseFrontmatter[testFrontmatter](input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing opening frontmatter delimiter")
}

func TestParseFrontmatter_MissingCloseDelimiter(t *testing.T) {
	input := []byte("---\nname: test\n")
	_, _, err := ParseFrontmatter[testFrontmatter](input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing closing frontmatter delimiter")
}

func TestParseFrontmatter_InvalidYAML(t *testing.T) {
	input := []byte("---\n: :\n  invalid: {{{\n---\n")
	_, _, err := ParseFrontmatter[testFrontmatter](input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse frontmatter YAML")
}

func TestParseFrontmatter_UnknownFieldsIgnored(t *testing.T) {
	input := []byte("---\nname: test\nunknown_field: value\nextra: 123\n---\n\nBody.\n")
	fm, body, err := ParseFrontmatter[testFrontmatter](input)
	require.NoError(t, err)
	assert.Equal(t, "test", fm.Name)
	assert.Equal(t, "Body.\n", body)
}

func TestRenderFile(t *testing.T) {
	fm := testFrontmatter{Name: "hello", Tags: []string{"x", "y"}, Count: 7}
	data, err := RenderFile(fm, "Some body text.\n")
	require.NoError(t, err)

	assert.Contains(t, string(data), "---\n")
	assert.Contains(t, string(data), "name: hello")
	assert.Contains(t, string(data), "Some body text.")
}

func TestRenderFile_EmptyBody(t *testing.T) {
	fm := testFrontmatter{Name: "test"}
	data, err := RenderFile(fm, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "name: test")
	// Should end with closing delimiter
	assert.Contains(t, content, "---\n")
}

func TestParseRenderRoundtrip(t *testing.T) {
	original := testFrontmatter{Name: "roundtrip", Tags: []string{"a", "b"}, Count: 99}
	body := "This is the body.\n"

	data, err := RenderFile(original, body)
	require.NoError(t, err)

	parsed, parsedBody, err := ParseFrontmatter[testFrontmatter](data)
	require.NoError(t, err)

	assert.Equal(t, original.Name, parsed.Name)
	assert.Equal(t, original.Tags, parsed.Tags)
	assert.Equal(t, original.Count, parsed.Count)
	assert.Equal(t, body, parsedBody)
}

func TestRenderFile_TrailingNewline(t *testing.T) {
	fm := testFrontmatter{Name: "test"}
	data, err := RenderFile(fm, "no trailing newline")
	require.NoError(t, err)
	assert.True(t, data[len(data)-1] == '\n', "rendered file should end with newline")
}
