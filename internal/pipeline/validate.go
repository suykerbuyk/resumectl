package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseJSONResponse extracts and unmarshals a JSON response from an LLM,
// handling the common case where the LLM wraps JSON in markdown code fences.
func ParseJSONResponse[T any](raw string) (T, error) {
	var zero T

	cleaned := stripCodeFences(raw)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return zero, fmt.Errorf("pipeline: empty JSON response")
	}

	var result T
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return zero, fmt.Errorf("pipeline: parse JSON response: %w", err)
	}

	return result, nil
}

// stripCodeFences removes markdown code fences (```json ... ``` or ``` ... ```)
// from around JSON content.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)

	// Check for opening fence
	if strings.HasPrefix(s, "```") {
		// Find end of first line (the opening fence line)
		idx := strings.Index(s, "\n")
		if idx >= 0 {
			s = s[idx+1:]
		}
		// Remove closing fence
		if lastIdx := strings.LastIndex(s, "```"); lastIdx >= 0 {
			s = s[:lastIdx]
		}
	}

	return strings.TrimSpace(s)
}

// ValidateResumeOutput checks a generated resume for common issues.
// Returns a list of warnings (empty if all checks pass).
func ValidateResumeOutput(content string, experienceCompanies []string) []string {
	var warnings []string

	lower := strings.ToLower(content)

	// Check for summary section
	if !strings.Contains(lower, "summary") && !strings.Contains(lower, "profile") {
		warnings = append(warnings, "resume missing summary/profile section")
	}

	// Check for at least one experience section marker
	hasExperience := strings.Contains(lower, "experience") ||
		strings.Contains(lower, "employment") ||
		strings.Contains(lower, "work history")
	if !hasExperience {
		// Check if any company names appear
		companyFound := false
		for _, company := range experienceCompanies {
			if strings.Contains(lower, strings.ToLower(company)) {
				companyFound = true
				break
			}
		}
		if !companyFound && len(experienceCompanies) > 0 {
			warnings = append(warnings, "resume may be missing experience sections")
		}
	}

	// Check word count for 2-page target (600-800 words)
	words := len(strings.Fields(content))
	if words < 200 {
		warnings = append(warnings, fmt.Sprintf("resume seems short (%d words, target 600-800)", words))
	} else if words > 1500 {
		warnings = append(warnings, fmt.Sprintf("resume may be too long (%d words, target 600-800)", words))
	}

	return warnings
}
