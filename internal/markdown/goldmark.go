package markdown

import (
	"bytes"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"

	"github.com/apricote/releaser-pleaser/internal/markdown/extensions"
)

func New() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extensions.Section),
		goldmark.WithParserOptions(parser.WithASTTransformers(
			util.Prioritized(&newLineTransformer{}, 1),
		)),
		goldmark.WithRenderer(markdown.NewRenderer()),
	)
}

// Format the Markdown document in a style mimicking Prettier. This is done for compatibility with other tools
// users might have installed in their IDE. This does not guarantee that the output matches Prettier exactly.
func Format(input string) (string, error) {
	var buf bytes.Buffer
	buf.Grow(len(input))

	err := New().Convert([]byte(input), &buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
