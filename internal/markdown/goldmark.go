package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	"github.com/apricote/releaser-pleaser/internal/markdown/extensions"
	"github.com/apricote/releaser-pleaser/internal/markdown/renderer/markdown"
)

func New() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extensions.Section),
		goldmark.WithRenderer(renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(markdown.NewRenderer(), 1)))),
	)
}
