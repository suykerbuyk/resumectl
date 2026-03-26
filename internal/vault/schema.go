package vault

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

// ParseFrontmatter splits a markdown file with YAML frontmatter into the
// typed frontmatter struct and the remaining body text. The file must begin
// with a "---" line, followed by YAML, followed by another "---" line.
func ParseFrontmatter[T any](data []byte) (T, string, error) {
	var zero T

	content := string(data)
	if len(content) < 3 || content[:3] != frontmatterDelimiter {
		return zero, "", fmt.Errorf("vault: missing opening frontmatter delimiter")
	}

	// Find the closing delimiter
	rest := content[3:]
	// Skip the newline after opening ---
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	closingIdx := bytes.Index([]byte(rest), []byte("\n"+frontmatterDelimiter))
	if closingIdx == -1 {
		return zero, "", fmt.Errorf("vault: missing closing frontmatter delimiter")
	}

	yamlContent := rest[:closingIdx]
	body := rest[closingIdx+1+len(frontmatterDelimiter):]

	// Strip leading newline(s) from body
	for len(body) > 0 && (body[0] == '\n' || body[0] == '\r') {
		body = body[1:]
	}

	var fm T
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return zero, "", fmt.Errorf("vault: parse frontmatter YAML: %w", err)
	}

	return fm, body, nil
}

// RenderFile serializes a frontmatter struct and body into a complete
// markdown file with YAML frontmatter delimiters.
func RenderFile(frontmatter any, body string) ([]byte, error) {
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("vault: marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(frontmatterDelimiter + "\n")
	buf.Write(yamlData)
	buf.WriteString(frontmatterDelimiter + "\n")
	if body != "" {
		buf.WriteString("\n")
		buf.WriteString(body)
		// Ensure trailing newline
		if body[len(body)-1] != '\n' {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}
