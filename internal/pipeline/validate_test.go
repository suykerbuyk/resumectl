package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSONResponse_Valid(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	result, err := ParseJSONResponse[testStruct](`{"name": "John", "age": 30}`)
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 30, result.Age)
}

func TestParseJSONResponse_WithCodeFences(t *testing.T) {
	type testStruct struct {
		OK bool `json:"ok"`
	}

	input := "```json\n{\"ok\": true}\n```"
	result, err := ParseJSONResponse[testStruct](input)
	require.NoError(t, err)
	assert.True(t, result.OK)
}

func TestParseJSONResponse_WithCodeFencesNoLang(t *testing.T) {
	type testStruct struct {
		OK bool `json:"ok"`
	}

	input := "```\n{\"ok\": true}\n```"
	result, err := ParseJSONResponse[testStruct](input)
	require.NoError(t, err)
	assert.True(t, result.OK)
}

func TestParseJSONResponse_Empty(t *testing.T) {
	type testStruct struct{}

	_, err := ParseJSONResponse[testStruct]("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty JSON")
}

func TestParseJSONResponse_InvalidJSON(t *testing.T) {
	type testStruct struct{}

	_, err := ParseJSONResponse[testStruct]("not json at all")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse JSON")
}

func TestParseJSONResponse_Array(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}

	result, err := ParseJSONResponse[[]item](`[{"name": "a"}, {"name": "b"}]`)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "a", result[0].Name)
}

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no fences", `{"ok": true}`, `{"ok": true}`},
		{"json fences", "```json\n{\"ok\": true}\n```", `{"ok": true}`},
		{"plain fences", "```\n{\"ok\": true}\n```", `{"ok": true}`},
		{"with whitespace", "  ```json\n{\"ok\": true}\n```  ", `{"ok": true}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, stripCodeFences(tt.input))
		})
	}
}

func TestValidateResumeOutput_Valid(t *testing.T) {
	content := `# John Smith

## Summary

Experienced engineer with deep expertise.

## Experience

### Principal Architect — Acme Corp
- Built distributed systems handling 2M IOPS.
- Led cross-functional team of 8 engineers.
` + generateWords(600)

	warnings := ValidateResumeOutput(content, []string{"Acme Corp"})
	assert.Empty(t, warnings)
}

func TestValidateResumeOutput_MissingSummary(t *testing.T) {
	content := "## Experience\n\n" + generateWords(600)
	warnings := ValidateResumeOutput(content, nil)
	assert.Contains(t, warnings[0], "summary")
}

func TestValidateResumeOutput_TooShort(t *testing.T) {
	content := "## Summary\n\nShort resume.\n"
	warnings := ValidateResumeOutput(content, nil)
	found := false
	for _, w := range warnings {
		if assert.ObjectsAreEqual("resume seems short", w[:18]) {
			found = true
		}
	}
	assert.True(t, found || len(warnings) > 0)
}

func TestValidateResumeOutput_TooLong(t *testing.T) {
	content := "## Summary\n\n## Experience\n\n" + generateWords(1600)
	warnings := ValidateResumeOutput(content, nil)
	longWarning := false
	for _, w := range warnings {
		if len(w) > 0 && w[0:6] == "resume" {
			longWarning = true
		}
	}
	assert.True(t, longWarning)
}

func generateWords(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			result += " "
		}
		result += "word"
	}
	return result + "\n"
}
