package rp

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type ReleasePullRequest struct {
	ID          int
	Title       string
	Description string
	Labels      []string
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
)

const (
	DescriptionLanguagePrefix = "rp-prefix"
	DescriptionLanguageSuffix = "rp-suffix"
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
	descriptionAST := parser.NewParser().Parse(text.NewReader(source))

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

func textFromLines(source []byte, n ast.Node) string {
	content := make([]byte, 0)

	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		content = append(content, line.Value(source)...)
	}

	return string(content)

}
