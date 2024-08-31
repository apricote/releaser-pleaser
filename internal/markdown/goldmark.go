package markdown

import (
	"bytes"
	"strings"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/apricote/releaser-pleaser/internal/markdown/extensions"
	"github.com/apricote/releaser-pleaser/internal/markdown/extensions/ast"
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

func GetCodeBlockText(source []byte, language string, output *string) gast.Walker {
	return func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		if n.Kind() != gast.KindFencedCodeBlock {
			return gast.WalkContinue, nil
		}

		codeBlock := n.(*gast.FencedCodeBlock)

		if string(codeBlock.Language(source)) != language {
			return gast.WalkContinue, nil
		}

		*output = textFromLines(source, codeBlock)
		// Stop looking after we find the first result
		return gast.WalkStop, nil
	}
}

func GetSectionText(source []byte, name string, output *string) gast.Walker {
	return func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		if n.Kind() != ast.KindSection {
			return gast.WalkContinue, nil
		}

		section := n.(*ast.Section)

		if section.Name != name {
			return gast.WalkContinue, nil
		}

		// Do not show section markings in output, we only care about the content
		section.HideInOutput()

		// Found the right section
		outputBuffer := new(bytes.Buffer)
		err := New().Renderer().Render(outputBuffer, source, section)
		if err != nil {
			return gast.WalkStop, err
		}

		*output = outputBuffer.String()
		// Stop looking after we find the first result
		return gast.WalkStop, nil
	}
}

func textFromLines(source []byte, n gast.Node) string {
	content := make([]byte, 0)

	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		content = append(content, line.Value(source)...)
	}

	return strings.TrimSpace(string(content))
}

func WalkAST(source []byte, walkers ...gast.Walker) (err error) {
	doc := New().Parser().Parse(text.NewReader(source))

	for _, walker := range walkers {
		err = gast.Walk(doc, walker)
		if err != nil {
			return err
		}
	}

	return nil
}
