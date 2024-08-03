package markdown

import (
	markdownfmt "github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"

	"github.com/apricote/releaser-pleaser/internal/markdown/extensions"
)

func NewParser() parser.Parser {
	p := goldmark.DefaultParser()
	p.AddOptions(parser.WithBlockParsers(util.Prioritized(extensions.NewSectionParser(), 0)))
	return p
}

func New() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extensions.Section),
		goldmark.WithRenderer(markdownfmt.NewRenderer()),
	)
}
