package prompts

import (
	"bytes"
	"fmt"
	"strings"
)

// Render executes the named template with the given data and returns the
// rendered string. Returns an error if the template doesn't exist or
// rendering fails.
func (l *Library) Render(name string, data any) (string, error) {
	tmpl, err := l.Get(name)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("prompts: render %s: %w", name, err)
	}

	result := buf.String()
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("prompts: render %s produced empty output", name)
	}

	return result, nil
}

// EstimateTokens provides a rough token count estimate for the given text.
// Uses the common ~4 characters per token heuristic.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}
