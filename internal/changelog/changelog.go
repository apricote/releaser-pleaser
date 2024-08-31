package changelog

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"log/slog"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/markdown"
)

var (
	changelogTemplate *template.Template
)

//go:embed changelog.md.tpl
var rawChangelogTemplate string

func init() {
	var err error
	changelogTemplate, err = template.New("changelog").Parse(rawChangelogTemplate)
	if err != nil {
		log.Fatalf("failed to parse changelog template: %v", err)
	}
}

func NewChangelogEntry(logger *slog.Logger, commits []commitparser.AnalyzedCommit, version, link, prefix, suffix string) (string, error) {
	features := make([]commitparser.AnalyzedCommit, 0)
	fixes := make([]commitparser.AnalyzedCommit, 0)

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
		"Prefix":      prefix,
		"Suffix":      suffix,
	})
	if err != nil {
		return "", err
	}

	formatted, err := markdown.Format(changelog.String())
	if err != nil {
		logger.Warn("failed to format changelog entry, using unformatted", "error", err)
		return changelog.String(), nil
	}

	return formatted, nil
}
