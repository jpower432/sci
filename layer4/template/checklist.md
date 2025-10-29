# Evaluation Checklist
{{- if .Metadata.Id}}

**Plan ID:** {{.Metadata.Id}}
{{- end}}
{{- if .Metadata.Version}}

**Version:** {{.Metadata.Version}}
{{- end}}
{{- if .Metadata.Author.Name}}

**Author:** {{.Metadata.Author.Name}}
{{- end}}
{{- if .Metadata.MappingReferences}}

## Mapping References
{{- range .Metadata.MappingReferences}}

- **{{.Title}}**{{if .Version}} ({{.Version}}){{end}}{{if .Description}}: {{.Description}}{{end}}{{if .Url}} - [Link]({{.Url}}){{end}}
{{- end}}
{{- end}}
{{- if .Plans}}

---
{{- range $index, $plan := .Plans}}
{{- if $index}}

{{- end}}

## Control: {{if $plan.Control.EntryId}}{{$plan.Control.EntryId}}{{else}}Unknown Control{{end}}
{{- if $plan.Control.Remarks}}

*{{$plan.Control.Remarks}}*
{{- end}}
{{- if not $plan.Assessments}}

*No assessments defined for this control.*
{{- else}}
{{- range $plan.Assessments}}

### Requirement: {{if .Requirement.EntryId}}{{.Requirement.EntryId}}{{else}}Unknown Requirement{{end}}
{{- if .Requirement.Remarks}}

*{{.Requirement.Remarks}}*
{{- end}}
{{- if not .Procedures}}

*No procedures defined for this requirement.*
{{- else}}
{{- range .Procedures}}

- [ ] {{if .Name}}**{{.Name}}**{{else if .Id}}**{{.Id}}**{{else}}**Procedure**{{end}}{{if and .Id .Name (ne .Id .Name)}} ({{codeBlock .Id}}){{end}}
{{- if .Description}}
{{- range (splitLines .Description)}}
  {{.}}
{{- end}}
{{- end}}
{{- if .Documentation}}
  ?? [Documentation]({{.Documentation}})
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}