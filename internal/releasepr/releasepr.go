package releasepr

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/markdown"
	ast2 "github.com/apricote/releaser-pleaser/internal/markdown/extensions/ast"
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

// ReleasePullRequest
//
// TODO: Reuse [git.PullRequest]
type ReleasePullRequest struct {
	ID          int
	Title       string
	Description string
	Labels      []Label

	Head          string
	ReleaseCommit *git.Commit
}

// Label is the string identifier of a pull/merge request label on the forge.
type Label string

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
	LabelNextVersionTypeNormal Label = "rp-next-version::normal"
	LabelNextVersionTypeRC     Label = "rp-next-version::rc"
	LabelNextVersionTypeBeta   Label = "rp-next-version::beta"
	LabelNextVersionTypeAlpha  Label = "rp-next-version::alpha"

	LabelReleasePending Label = "rp-release::pending"
	LabelReleaseTagged  Label = "rp-release::tagged"
)

var KnownLabels = []Label{
	LabelNextVersionTypeNormal,
	LabelNextVersionTypeRC,
	LabelNextVersionTypeBeta,
	LabelNextVersionTypeAlpha,

	LabelReleasePending,
	LabelReleaseTagged,
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
	descriptionAST := markdown.New().Parser().Parse(text.NewReader(source))

	err := ast.Walk(descriptionAST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Type() != ast.TypeBlock || n.Kind() != ast.KindFencedCodeBlock {
			return ast.WalkContinue, nil
		}

		codeBlock, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkStop, fmt.Errorf("node has unexpected type: %T", n)
		}

		switch string(codeBlock.Language(source)) {
		case DescriptionLanguagePrefix:
			overrides.Prefix = textFromLines(source, codeBlock)
		case DescriptionLanguageSuffix:
			overrides.Suffix = textFromLines(source, codeBlock)
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return ReleaseOverrides{}, err
	}

	return overrides, nil
}

func (pr *ReleasePullRequest) ChangelogText() (string, error) {
	source := []byte(pr.Description)
	gm := markdown.New()
	descriptionAST := gm.Parser().Parse(text.NewReader(source))

	var section *ast2.Section

	err := ast.Walk(descriptionAST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Type() != ast.TypeBlock || n.Kind() != ast2.KindSection {
			return ast.WalkContinue, nil
		}

		anySection, ok := n.(*ast2.Section)
		if !ok {
			return ast.WalkStop, fmt.Errorf("node has unexpected type: %T", n)
		}

		if anySection.Name != MarkdownSectionChangelog {
			return ast.WalkContinue, nil
		}

		section = anySection
		return ast.WalkStop, nil
	})
	if err != nil {
		return "", err
	}

	if section == nil {
		return "", nil
	}

	outputBuffer := new(bytes.Buffer)
	err = gm.Renderer().Render(outputBuffer, source, section)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

func textFromLines(source []byte, n ast.Node) string {
	content := make([]byte, 0)

	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		content = append(content, line.Value(source)...)
	}

	return strings.TrimSpace(string(content))
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
