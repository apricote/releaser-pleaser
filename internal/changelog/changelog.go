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

func DefaultTemplate() *template.Template {
	return changelogTemplate
}

type Data struct {
	Commits     map[string][]commitparser.AnalyzedCommit
	Version     string
	VersionLink string
	Prefix      string
	Suffix      string
}

func New(commits map[string][]commitparser.AnalyzedCommit, version, versionLink, prefix, suffix string) Data {
	return Data{
		Commits:     commits,
		Version:     version,
		VersionLink: versionLink,
		Prefix:      prefix,
		Suffix:      suffix,
	}
}

type Formatting struct {
	HideVersionTitle bool
}

func Entry(logger *slog.Logger, tpl *template.Template, data Data, formatting Formatting) (string, error) {
	var changelog bytes.Buffer
	err := tpl.Execute(&changelog, map[string]any{
		"Data":       data,
		"Formatting": formatting,
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
