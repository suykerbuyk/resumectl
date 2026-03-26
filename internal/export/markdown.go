package export

import (
	"fmt"
	"strings"

	"github.com/jsuykerbuyk/resumectl/internal/vault"
)

// AssembleMarkdown combines a resume body with contact info into a
// canonical markdown document with normalized headings.
func AssembleMarkdown(resume *vault.ResumeFile, contact *vault.ContactInfo) (string, error) {
	if strings.TrimSpace(resume.Body) == "" {
		return "", fmt.Errorf("export: resume body is empty")
	}

	var b strings.Builder

	// Contact header
	if contact != nil {
		b.WriteString("# " + contact.Name + "\n\n")

		var details []string
		if contact.Email != "" {
			details = append(details, contact.Email)
		}
		if contact.Phone != "" {
			details = append(details, contact.Phone)
		}
		if contact.Location != "" {
			details = append(details, contact.Location)
		}
		if len(details) > 0 {
			b.WriteString(strings.Join(details, " | ") + "\n")
		}

		var links []string
		if contact.LinkedIn != "" {
			links = append(links, contact.LinkedIn)
		}
		if contact.GitHub != "" {
			links = append(links, contact.GitHub)
		}
		if contact.Website != "" {
			links = append(links, contact.Website)
		}
		if len(links) > 0 {
			b.WriteString(strings.Join(links, " | ") + "\n")
		}

		b.WriteString("\n---\n\n")
	}

	// Normalize headings: ensure no h1 in body (reserve h1 for name)
	body := normalizeHeadings(resume.Body)
	b.WriteString(body)

	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}

	return b.String(), nil
}

// normalizeHeadings ensures the body uses h2+ headings (no h1).
// If the body contains h1 headings, they are demoted to h2.
func normalizeHeadings(body string) string {
	lines := strings.Split(body, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			// Demote h1 to h2
			line = strings.Replace(line, "# ", "## ", 1)
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
