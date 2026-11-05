{{define "entry" -}}
- {{ if .BreakingChange}}**BREAKING**: {{end}}{{ if .Scope }}**{{.Scope}}**: {{end}}{{.Description}} ([{{.ShortHash}}]({{.URL}}))
{{ end }}

{{- if not .Formatting.HideVersionTitle }}
## [{{.Data.Version}}]({{.Data.VersionLink}})
{{ if .Data.CompareURL }}
[Compare to previous version]({{.Data.CompareURL}})
{{ end -}}
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

{{- $hasOther := false -}}
{{- range $type, $commits := .Data.Commits -}}
{{- if and (ne $type "feat") (ne $type "fix") $commits -}}
{{- if not $hasOther }}
### Other

{{ $hasOther = true -}}
{{- end -}}
{{ range $commits -}}{{template "entry" .}}{{end}}
{{- end -}}
{{- end -}}

{{- if .Data.Suffix }}
{{ .Data.Suffix }}
{{ end }}
