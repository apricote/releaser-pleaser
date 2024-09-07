package releasepr

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"regexp"
	"text/template"

	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/markdown"
	"github.com/apricote/releaser-pleaser/internal/versioning"
)

var (
	releasePRTemplate *template.Template
)

//go:embed releasepr.md.tpl
var rawReleasePRTemplate string

func init() {
	var err error
	releasePRTemplate, err = template.New("releasepr").Parse(rawReleasePRTemplate)
	if err != nil {
		log.Fatalf("failed to parse release pr template: %v", err)
	}
}

type ReleasePullRequest struct {
	git.PullRequest
	Labels []Label

	Head          string
	ReleaseCommit *git.Commit
}

func NewReleasePullRequest(head, branch, version, changelogEntry string) (*ReleasePullRequest, error) {
	rp := &ReleasePullRequest{
		Head:   head,
		Labels: []Label{LabelReleasePending},
	}

	rp.SetTitle(branch, version)
	if err := rp.SetDescription(changelogEntry, ReleaseOverrides{}); err != nil {
		return nil, err
	}

	return rp, nil
}

type ReleaseOverrides struct {
	Prefix          string
	Suffix          string
	NextVersionType versioning.NextVersionType
}

const (
	DescriptionLanguagePrefix = "rp-prefix"
	DescriptionLanguageSuffix = "rp-suffix"
)

const (
	MarkdownSectionChangelog = "changelog"
)

const (
	TitleFormat = "chore(%s): release %s"
)

var (
	TitleRegex = regexp.MustCompile("chore(.*): release (.*)")
)

func (pr *ReleasePullRequest) GetOverrides() (ReleaseOverrides, error) {
	overrides := ReleaseOverrides{}
	overrides = pr.parseVersioningFlags(overrides)
	overrides, err := pr.parseDescription(overrides)
	if err != nil {
		return ReleaseOverrides{}, err
	}

	return overrides, nil
}

func (pr *ReleasePullRequest) parseVersioningFlags(overrides ReleaseOverrides) ReleaseOverrides {
	for _, label := range pr.Labels {
		switch label {
		// Versioning
		case LabelNextVersionTypeNormal:
			overrides.NextVersionType = versioning.NextVersionTypeNormal
		case LabelNextVersionTypeRC:
			overrides.NextVersionType = versioning.NextVersionTypeRC
		case LabelNextVersionTypeBeta:
			overrides.NextVersionType = versioning.NextVersionTypeBeta
		case LabelNextVersionTypeAlpha:
			overrides.NextVersionType = versioning.NextVersionTypeAlpha
		case LabelReleasePending, LabelReleaseTagged:
			// These labels have no effect on the versioning.
			break
		}
	}

	return overrides
}

func (pr *ReleasePullRequest) parseDescription(overrides ReleaseOverrides) (ReleaseOverrides, error) {
	source := []byte(pr.Description)

	err := markdown.WalkAST(source,
		markdown.GetCodeBlockText(source, DescriptionLanguagePrefix, &overrides.Prefix, nil),
		markdown.GetCodeBlockText(source, DescriptionLanguageSuffix, &overrides.Suffix, nil),
	)
	if err != nil {
		return ReleaseOverrides{}, err
	}

	return overrides, nil
}

func (pr *ReleasePullRequest) ChangelogText() (string, error) {
	source := []byte(pr.Description)

	var sectionText string
	err := markdown.WalkAST(source, markdown.GetSectionText(source, MarkdownSectionChangelog, &sectionText))
	if err != nil {
		return "", err
	}

	return sectionText, nil

}

func (pr *ReleasePullRequest) SetTitle(branch, version string) {
	pr.Title = fmt.Sprintf(TitleFormat, branch, version)
}

func (pr *ReleasePullRequest) Version() (string, error) {
	matches := TitleRegex.FindStringSubmatch(pr.Title)
	if len(matches) != 3 {
		return "", fmt.Errorf("title has unexpected format")
	}

	return matches[2], nil
}

func (pr *ReleasePullRequest) SetDescription(changelogEntry string, overrides ReleaseOverrides) error {
	var description bytes.Buffer
	err := releasePRTemplate.Execute(&description, map[string]any{
		"Changelog": changelogEntry,
		"Overrides": overrides,
	})
	if err != nil {
		return err
	}

	pr.Description = description.String()

	return nil
}
