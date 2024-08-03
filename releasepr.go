package rp

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"text/template"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/apricote/releaser-pleaser/internal/markdown"
	east "github.com/apricote/releaser-pleaser/internal/markdown/extensions/ast"
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
	ID          int
	Title       string
	Description string
	Labels      []string

	Head string
}

func NewReleasePullRequest(head, branch, version, changelogEntry string) (*ReleasePullRequest, error) {
	rp := &ReleasePullRequest{
		Head:   head,
		Labels: []string{LabelReleasePending},
	}

	rp.SetTitle(branch, version)
	if err := rp.SetDescription(changelogEntry); err != nil {
		return nil, err
	}

	return rp, nil
}

type ReleaseOverrides struct {
	Prefix string
	Suffix string
	// TODO: Doing the changelog for normal releases after previews requires to know about this while fetching the changesets
	NextVersionType NextVersionType
}

type NextVersionType int

const (
	NextVersionTypeUndefined NextVersionType = iota
	NextVersionTypeNormal
	NextVersionTypeRC
	NextVersionTypeBeta
	NextVersionTypeAlpha
)

func (n NextVersionType) String() string {
	switch n {
	case NextVersionTypeUndefined:
		return "undefined"
	case NextVersionTypeNormal:
		return "normal"
	case NextVersionTypeRC:
		return "rc"
	case NextVersionTypeBeta:
		return "beta"
	case NextVersionTypeAlpha:
		return "alpha"
	default:
		return ""
	}
}

// PR Labels
const (
	LabelNextVersionTypeNormal = "rp-next-version::normal"
	LabelNextVersionTypeRC     = "rp-next-version::rc"
	LabelNextVersionTypeBeta   = "rp-next-version::beta"
	LabelNextVersionTypeAlpha  = "rp-next-version::alpha"

	LabelReleasePending = "rp-release::pending"
	LabelReleaseTagged  = "rp-release::tagged"
)

const (
	DescriptionLanguagePrefix = "rp-prefix"
	DescriptionLanguageSuffix = "rp-suffix"
)

const (
	MarkdownSectionOverrides = "overrides"
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
			overrides.NextVersionType = NextVersionTypeNormal
		case LabelNextVersionTypeRC:
			overrides.NextVersionType = NextVersionTypeRC
		case LabelNextVersionTypeBeta:
			overrides.NextVersionType = NextVersionTypeBeta
		case LabelNextVersionTypeAlpha:
			overrides.NextVersionType = NextVersionTypeAlpha
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

func (pr *ReleasePullRequest) getCurrentOverridesText() (string, error) {
	source := []byte(pr.Description)
	gm := markdown.New()
	descriptionAST := gm.Parser().Parse(text.NewReader(source))

	var section *east.Section

	err := ast.Walk(descriptionAST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Type() != ast.TypeBlock || n.Kind() != east.KindSection {
			return ast.WalkContinue, nil
		}

		anySection, ok := n.(*east.Section)
		if !ok {
			return ast.WalkStop, fmt.Errorf("node has unexpected type: %T", n)
		}

		if anySection.Name != MarkdownSectionOverrides {
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

	return string(content)
}

func (pr *ReleasePullRequest) SetTitle(branch, version string) {
	pr.Title = fmt.Sprintf("chore(%s): release %s", branch, version)
}

func (pr *ReleasePullRequest) SetDescription(changelogEntry string) error {
	overrides, err := pr.getCurrentOverridesText()
	if err != nil {
		return err
	}

	var description bytes.Buffer
	err = releasePRTemplate.Execute(&description, map[string]any{
		"Changelog": changelogEntry,
		"Overrides": overrides,
	})
	if err != nil {
		return err
	}

	pr.Description = description.String()

	return nil
}
