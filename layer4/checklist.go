package layer4

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed template/checklist.md
var checklistTemplate string

// ToChecklist converts an EvaluationPlan into a Markdown template checklist.
func (e EvaluationPlan) ToChecklist() (string, error) {
	// Parse the template with custom functions
	tmpl, err := template.New("checklist").
		Funcs(template.FuncMap{
			"splitLines": func(s string) []string {
				return strings.Split(s, "\n")
			},
			"ne": func(a, b string) bool {
				return a != b
			},
			"codeBlock": func(s string) string {
				return "`" + s + "`"
			},
		}).
		Parse(checklistTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, e); err != nil {
		return "", err
	}

	return buf.String(), nil
}
