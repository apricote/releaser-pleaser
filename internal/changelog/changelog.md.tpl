{{define "entry" -}}
- {{ if .Scope }}**{{.Scope}}**: {{end}}{{.Description}}
{{ end }}

{{- if not .Formatting.HideVersionTitle }}
## [{{.Data.Version}}]({{.Data.VersionLink}})
{{ end -}}
{{- if .Data.Prefix }}
{{ .Data.Prefix }}
{{ end -}}
{{- with .Data.Commits.feat }}
### Features

{{ range . -}}{{template "entry" .}}{{end}}
{{- end -}}
{{- with .Data.Commits.fix }}
### Bug Fixes

{{ range . -}}{{template "entry" .}}{{end}}
{{- end -}}

{{- if .Data.Suffix }}
{{ .Data.Suffix }}
{{ end }}
