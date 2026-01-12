package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

const lexiconTemplate = `---
layout: page
---

# {{.Metadata.Title}}

**Version**: <span class="badge badge-version">{{.Metadata.Version}}</span>

{{.Metadata.Description}}

| Entity | Definition | Context | Source |
|--------|------------|---------|--------|
{{range .Categories}}{{range .Terms}}| **{{.Entity}}** | {{trim (replace "\n" " " .Definition)}} | {{if .Context}}{{.Context}}{{else}} {{end}} | {{if .Source}}{{.Source}}{{else}} {{end}} |
{{end}}{{end}}
`

type Lexicon struct {
	Metadata   `yaml:"metadata"`
	Categories []Category `yaml:"categories"`
}

type Metadata struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type Category struct {
	Name  string `yaml:"name"`
	Terms []Term `yaml:"terms"`
}

type Term struct {
	Entity     string `yaml:"entity"`
	Definition string `yaml:"definition"`
	Context    string `yaml:"context"`
	Source     string `yaml:"source,omitempty"`
}

func (l *Lexicon) ToMarkdownPage(writer io.Writer) error {
	funcMap := template.FuncMap{
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
	}
	tmpl, err := template.New("lexicon").Funcs(funcMap).Parse(lexiconTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err := tmpl.Execute(writer, l); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}
