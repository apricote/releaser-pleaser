package rp

import (
	"bytes"
	"html/template"
	"log"

	"github.com/go-git/go-git/v5"
)

var (
	changelogTemplate *template.Template
)

func init() {
	var err error
	changelogTemplate, err = template.New("changelog").Parse(`## [{{.Version}}]({{.VersionLink}})
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
`,
	)
	if err != nil {
		log.Fatalf("failed to parse changelog template: %v", err)
	}
}

func UpdateChangelog(wt *git.Worktree, commits []AnalyzedCommit) error {
	return nil
}

func formatChangelog(commits []AnalyzedCommit, version, link string) (string, error) {
	features := make([]AnalyzedCommit, 0)
	fixes := make([]AnalyzedCommit, 0)

	for _, commit := range commits {
		switch commit.Type {
		case "feat":
			features = append(features, commit)
		case "fix":
			fixes = append(fixes, commit)
		}
	}

	var changelog bytes.Buffer
	err := changelogTemplate.Execute(&changelog, map[string]any{
		"Features":    features,
		"Fixes":       fixes,
		"Version":     version,
		"VersionLink": link,
	})
	if err != nil {
		return "", err
	}

	return changelog.String(), nil

}
