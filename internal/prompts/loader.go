package prompts

import (
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateNames lists all available prompt templates.
var TemplateNames = []string{
	"decompose_resume",
	"extract_job",
	"gap_analysis",
	"synthesize_resume",
	"synthesize_cover_letter",
	"coach_question",
}

// Library holds parsed prompt templates, ready for rendering.
type Library struct {
	templates map[string]*template.Template
}

// NewLibrary parses all embedded templates and returns a Library.
func NewLibrary() (*Library, error) {
	lib := &Library{
		templates: make(map[string]*template.Template),
	}

	for _, name := range TemplateNames {
		filename := "templates/" + name + ".tmpl"
		data, err := templateFS.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("prompts: read %s: %w", filename, err)
		}

		tmpl, err := template.New(name).Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("prompts: parse %s: %w", name, err)
		}

		lib.templates[name] = tmpl
	}

	return lib, nil
}

// Get returns the named template, or an error if it doesn't exist.
func (l *Library) Get(name string) (*template.Template, error) {
	tmpl, ok := l.templates[name]
	if !ok {
		return nil, fmt.Errorf("prompts: unknown template %q", name)
	}
	return tmpl, nil
}

// Names returns the list of available template names.
func (l *Library) Names() []string {
	names := make([]string, 0, len(l.templates))
	for name := range l.templates {
		names = append(names, name)
	}
	return names
}
