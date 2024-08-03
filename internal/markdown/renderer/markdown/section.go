package markdown

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"

	"github.com/apricote/releaser-pleaser/internal/markdown/extensions"
	rpexast "github.com/apricote/releaser-pleaser/internal/markdown/extensions/ast"
)

func (r *Renderer) renderSection(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	n := node.(*rpexast.Section)

	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if _, err := r.writeString(w, fmt.Sprintf(extensions.SectionStartFormat, n.Name)+"\n"); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if _, err := r.writeString(w, "\n"+fmt.Sprintf(extensions.SectionEndFormat, n.Name)); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}
