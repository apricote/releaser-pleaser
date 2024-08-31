package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type newLineTransformer struct{}

var _ parser.ASTTransformer = (*newLineTransformer)(nil) // interface compliance

func (t *newLineTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	// No error can happen as they can only come from the walker function
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || node.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}

		switch node.Kind() {
		case ast.KindListItem:
			// Do not add empty lines between every list item
			break
		default:
			// Add empty lines between every other block
			node.SetBlankPreviousLines(true)
		}

		return ast.WalkContinue, nil
	})
}
