## [{{.Version}}]({{.VersionLink}})
{{- if .Prefix }}
{{ .Prefix }}
{{ end -}}
{{- if (gt (len .Features) 0) }}
### Features

{{ range .Features -}}
- {{ if .Scope }}**{{.Scope}}**: {{end}}{{.Description}}
{{ end -}}
{{- end -}}
{{- if (gt (len .Fixes) 0) }}
### Bug Fixes

{{ range .Fixes -}}
- {{ if .Scope }}**{{.Scope}}**: {{end}}{{.Description}}
{{ end -}}
{{- end -}}

{{- if .Suffix }}
{{ .Suffix }}
{{ end -}}
