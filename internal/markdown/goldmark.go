package markdown

import (
	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"

	"github.com/apricote/releaser-pleaser/internal/markdown/extensions"
)

func New() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extensions.Section),
		goldmark.WithRenderer(markdown.NewRenderer()),
	)
}
