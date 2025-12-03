package layer3

// markdownTemplate is the default template for generating markdown checklist output.
// This template is used internally by ToMarkdownChecklist().
const markdownTemplate = `{{if .PolicyId}}# Implementation Guidance: {{.PolicyId}}

{{end}}{{if .Author}}**Author:** {{.Author}}{{if .AuthorVersion}} (v{{.AuthorVersion}}){{end}}

{{end}}{{range $index, $section := .Sections}}{{if $index}}
---

{{end}}## {{$section.ControlName}}

{{if $section.ControlReference}}**Control:** {{$section.ControlReference}}

{{end}}{{if eq (len $section.Items) 0}}- [ ] No assessments defined
{{else}}{{range $section.Items}}- [ ] {{if .RequirementId}}**{{.RequirementId}}**: {{end}}{{.Requirement}}{{if and .Description (ne .Description .Requirement)}} - {{.Description}}{{end}}
{{if .Documentation}}    > [Documentation]({{.Documentation}})
{{end}}{{end}}{{end}}{{end}}`
