package markdown

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	exast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	rpexast "github.com/apricote/releaser-pleaser/internal/markdown/extensions/ast"
)

type blockState struct {
	node  ast.Node
	fresh bool
}

type listState struct {
	marker  byte
	ordered bool
	index   int
}

type Renderer struct {
	listStack   []listState
	openBlocks  []blockState
	prefixStack []string
	prefix      []byte
	atNewline   bool
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer() renderer.NodeRenderer {
	r := &Renderer{}

	return r
}

func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// default registrations
	// blocks
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)

	// GFM Extensions
	// Tables
	reg.Register(exast.KindTable, r.renderTable)
	reg.Register(exast.KindTableHeader, r.renderTableHeader)
	reg.Register(exast.KindTableRow, r.renderTableRow)
	reg.Register(exast.KindTableCell, r.renderTableCell)
	// Strikethrough
	reg.Register(exast.KindStrikethrough, r.renderStrikethrough)
	// Checkbox
	reg.Register(exast.KindTaskCheckBox, r.renderTaskCheckBox)

	// releaser-pleaser Extensions
	// Section
	reg.Register(rpexast.KindSection, r.renderSection)
}

func (r *Renderer) write(w io.Writer, buf []byte) (int, error) {
	written := 0
	for len(buf) > 0 {
		if r.atNewline {
			if err := r.beginLine(w); err != nil {
				return 0, fmt.Errorf(": %w", err)
			}
		}

		atNewline := false
		newline := bytes.IndexByte(buf, '\n')
		if newline == -1 {
			newline = len(buf) - 1
		} else {
			atNewline = true
		}

		n, err := w.Write(buf[:newline+1])
		written += n
		r.atNewline = n > 0 && atNewline && n == newline+1
		if len(r.openBlocks) != 0 {
			r.openBlocks[len(r.openBlocks)-1].fresh = false
		}
		if err != nil {
			return written, fmt.Errorf(": %w", err)
		}
		buf = buf[n:]
	}
	return written, nil
}

func (r *Renderer) beginLine(w io.Writer) error {
	if len(r.openBlocks) != 0 {
		current := r.openBlocks[len(r.openBlocks)-1]
		if current.node.Kind() == ast.KindParagraph && !current.fresh {
			return nil
		}
	}

	n, err := w.Write(r.prefix)
	if n != 0 {
		r.atNewline = r.prefix[len(r.prefix)-1] == '\n'
	}
	if err != nil {
		return fmt.Errorf(": %w", err)
	}
	return nil
}

func (r *Renderer) writeLines(w util.BufWriter, source []byte, lines *text.Segments) error {
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		if _, err := r.write(w, line.Value(source)); err != nil {
			return fmt.Errorf(": %w", err)
		}
	}
	return nil
}

func (r *Renderer) writeByte(w io.Writer, c byte) error {
	if _, err := r.write(w, []byte{c}); err != nil {
		return fmt.Errorf(": %w", err)
	}
	return nil
}

// WriteString writes a string to an io.Writer, ensuring that appropriate indentation and prefixes are added at the
// beginning of each line.
func (r *Renderer) writeString(w io.Writer, s string) (int, error) {
	n, err := r.write(w, []byte(s))
	if err != nil {
		return n, fmt.Errorf(": %w", err)
	}
	return n, nil
}

// PushIndent adds the specified amount of indentation to the current line prefix.
func (r *Renderer) pushIndent(amount int) {
	r.pushPrefix(strings.Repeat(" ", amount))
}

// PushPrefix adds the specified string to the current line prefix.
func (r *Renderer) pushPrefix(prefix string) {
	r.prefixStack = append(r.prefixStack, prefix)
	r.prefix = append(r.prefix, []byte(prefix)...)
}

// PopPrefix removes the last piece added by a call to PushIndent or PushPrefix from the current line prefix.
func (r *Renderer) popPrefix() {
	r.prefix = r.prefix[:len(r.prefix)-len(r.prefixStack[len(r.prefixStack)-1])]
	r.prefixStack = r.prefixStack[:len(r.prefixStack)-1]
}

// OpenBlock ensures that each block begins on a new line, and that blank lines are inserted before blocks as
// indicated by node.HasPreviousBlankLines.
func (r *Renderer) openBlock(w util.BufWriter, _ []byte, node ast.Node) error {
	r.openBlocks = append(r.openBlocks, blockState{
		node:  node,
		fresh: true,
	})

	hasBlankPreviousLines := node.HasBlankPreviousLines()

	// FIXME: standard goldmark table parser doesn't recognize Blank Previous Lines so we'll always add one
	if node.Kind() == exast.KindTable {
		hasBlankPreviousLines = true
	}

	// Work around the fact that the first child of a node notices the same set of preceding blank lines as its parent.
	if p := node.Parent(); p != nil && p.FirstChild() == node {
		if p.Kind() == ast.KindDocument || p.Kind() == ast.KindListItem || p.HasBlankPreviousLines() {
			hasBlankPreviousLines = false
		}
	}

	if hasBlankPreviousLines {
		if err := r.writeByte(w, '\n'); err != nil {
			return fmt.Errorf(": %w", err)
		}
	}

	r.openBlocks[len(r.openBlocks)-1].fresh = true

	return nil
}

// CloseBlock marks the current block as closed.
func (r *Renderer) closeBlock(w io.Writer) error {
	if !r.atNewline {
		if err := r.writeByte(w, '\n'); err != nil {
			return fmt.Errorf(": %w", err)
		}
	}

	r.openBlocks = r.openBlocks[:len(r.openBlocks)-1]
	return nil
}

// RenderDocument renders an *ast.Document node to the given BufWriter.
func (r *Renderer) renderDocument(_ util.BufWriter, _ []byte, _ ast.Node, _ bool) (ast.WalkStatus, error) {
	r.listStack, r.prefixStack, r.prefix, r.atNewline = nil, nil, nil, false
	return ast.WalkContinue, nil
}

// RenderHeading renders an *ast.Heading node to the given BufWriter.
func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if _, err := r.writeString(w, strings.Repeat("#", node.(*ast.Heading).Level)); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if err := r.writeByte(w, ' '); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if err := r.writeByte(w, '\n'); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderBlockquote renders an *ast.Blockquote node to the given BufWriter.
func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if _, err := r.writeString(w, "> "); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		r.pushPrefix("> ")
	} else {
		r.popPrefix()

		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderCodeBlock renders an *ast.CodeBlock node to the given BufWriter.
func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		r.popPrefix()
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
		return ast.WalkContinue, nil
	}

	if err := r.openBlock(w, source, node); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	// // Each line of a code block needs to be aligned at the same offset, and a code block must start with at least four
	// // spaces. To achieve this, we unconditionally add four spaces to the first line of the code block and indent the
	// // rest as necessary.
	// if _, err := r.writeString(w, "    "); err != nil {
	// 	return ast.WalkStop, fmt.Errorf(": %w", err)
	// }

	r.pushIndent(4)
	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	return ast.WalkContinue, nil
}

// RenderFencedCodeBlock renders an *ast.FencedCodeBlock node to the given BufWriter.
func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
		return ast.WalkContinue, nil
	}

	if err := r.openBlock(w, source, node); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	code := node.(*ast.FencedCodeBlock)

	// Write the start of the fenced code block.
	fence := []byte("```")
	if _, err := r.write(w, fence); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	language := code.Language(source)
	if _, err := r.write(w, language); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	if err := r.writeByte(w, '\n'); err != nil {
		return ast.WalkStop, nil
	}

	// Write the contents of the fenced code block.
	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	// Write the end of the fenced code block.
	if err := r.beginLine(w); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	if _, err := r.write(w, fence); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	if err := r.writeByte(w, '\n'); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	return ast.WalkContinue, nil
}

// RenderHTMLBlock renders an *ast.HTMLBlock node to the given BufWriter.
func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
		return ast.WalkContinue, nil
	}

	if err := r.openBlock(w, source, node); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	// Write the contents of the HTML block.
	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	// Write the closure line, if any.
	html := node.(*ast.HTMLBlock)
	if html.HasClosure() {
		if _, err := r.write(w, html.ClosureLine.Value(source)); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderList renders an *ast.List node to the given BufWriter.
func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		list := node.(*ast.List)
		r.listStack = append(r.listStack, listState{
			marker:  list.Marker,
			ordered: list.IsOrdered(),
			index:   list.Start,
		})
	} else {
		r.listStack = r.listStack[:len(r.listStack)-1]
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderListItem renders an *ast.ListItem node to the given BufWriter.
func (r *Renderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		markerWidth := 2 // marker + space

		state := &r.listStack[len(r.listStack)-1]
		if state.ordered {
			width, err := r.writeString(w, strconv.FormatInt(int64(state.index), 10))
			if err != nil {
				return ast.WalkStop, fmt.Errorf(": %w", err)
			}
			state.index++
			markerWidth += width // marker, space, and digits
		}

		if _, err := r.write(w, []byte{state.marker, ' '}); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		r.pushIndent(markerWidth)
	} else {
		r.popPrefix()
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderParagraph renders an *ast.Paragraph node to the given BufWriter.
func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		// A paragraph that follows another paragraph or a blockquote must be preceded by a blank line.
		if !node.HasBlankPreviousLines() {
			if prev := node.PreviousSibling(); prev != nil && (prev.Kind() == ast.KindParagraph || prev.Kind() == ast.KindBlockquote) {
				if err := r.writeByte(w, '\n'); err != nil {
					return ast.WalkStop, fmt.Errorf(": %w", err)
				}
			}
		}

		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderTextBlock renders an *ast.TextBlock node to the given BufWriter.
func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderThematicBreak renders an *ast.ThematicBreak node to the given BufWriter.
func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
		return ast.WalkContinue, nil
	}

	if err := r.openBlock(w, source, node); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	// TODO: this prints an extra no line
	if _, err := r.writeString(w, "--------"); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	return ast.WalkContinue, nil
}

// RenderAutoLink renders an *ast.AutoLink node to the given BufWriter.
func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	if err := r.writeByte(w, '<'); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	if _, err := r.write(w, node.(*ast.AutoLink).Label(source)); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	if err := r.writeByte(w, '>'); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) shouldPadCodeSpan(source []byte, node *ast.CodeSpan) bool {
	c := node.FirstChild()
	if c == nil {
		return false
	}

	segment := c.(*ast.Text).Segment
	text := segment.Value(source)

	var firstChar byte
	if len(text) > 0 {
		firstChar = text[0]
	}

	allWhitespace := true
	for {
		if util.FirstNonSpacePosition(text) != -1 {
			allWhitespace = false
			break
		}
		c = c.NextSibling()
		if c == nil {
			break
		}
		segment = c.(*ast.Text).Segment
		text = segment.Value(source)
	}
	if allWhitespace {
		return false
	}

	var lastChar byte
	if len(text) > 0 {
		lastChar = text[len(text)-1]
	}

	return firstChar == '`' || firstChar == ' ' || lastChar == '`' || lastChar == ' '
}

// RenderCodeSpan renders an *ast.CodeSpan node to the given BufWriter.
func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	code := node.(*ast.CodeSpan)
	delimiter := []byte{'`'}
	pad := r.shouldPadCodeSpan(source, code)

	if _, err := r.write(w, delimiter); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	if pad {
		if err := r.writeByte(w, ' '); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		text := c.(*ast.Text).Segment
		if _, err := r.write(w, text.Value(source)); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}
	if pad {
		if err := r.writeByte(w, ' '); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}
	if _, err := r.write(w, delimiter); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	return ast.WalkSkipChildren, nil
}

// RenderEmphasis renders an *ast.Emphasis node to the given BufWriter.
func (r *Renderer) renderEmphasis(w util.BufWriter, _ []byte, node ast.Node, _ bool) (ast.WalkStatus, error) {
	em := node.(*ast.Emphasis)
	if _, err := r.writeString(w, strings.Repeat("*", em.Level)); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) escapeLinkDest(dest []byte) []byte {
	requiresEscaping := false
	for _, c := range dest {
		if c <= 32 || c == '(' || c == ')' || c == 127 {
			requiresEscaping = true
			break
		}
	}
	if !requiresEscaping {
		return dest
	}

	escaped := make([]byte, 0, len(dest)+2)
	escaped = append(escaped, '<')
	for _, c := range dest {
		if c == '<' || c == '>' {
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, c)
	}
	escaped = append(escaped, '>')
	return escaped
}

func (r *Renderer) linkTitleDelimiter(title []byte) byte {
	for i, c := range title {
		if c == '"' && (i == 0 || title[i-1] != '\\') {
			return '\''
		}
	}
	return '"'
}

func (r *Renderer) renderLinkOrImage(w util.BufWriter, open string, dest, title []byte, enter bool) error {
	if enter {
		if _, err := r.writeString(w, open); err != nil {
			return fmt.Errorf(": %w", err)
		}
	} else {
		if _, err := r.writeString(w, "]("); err != nil {
			return fmt.Errorf(": %w", err)
		}

		if _, err := r.write(w, r.escapeLinkDest(dest)); err != nil {
			return fmt.Errorf(": %w", err)
		}
		if len(title) != 0 {
			delimiter := r.linkTitleDelimiter(title)
			if _, err := fmt.Fprintf(w, ` %c%s%c`, delimiter, string(title), delimiter); err != nil {
				return fmt.Errorf(": %w", err)
			}
		}

		if err := r.writeByte(w, ')'); err != nil {
			return fmt.Errorf(": %w", err)
		}
	}
	return nil
}

// RenderImage renders an *ast.Image node to the given BufWriter.
func (r *Renderer) renderImage(w util.BufWriter, _ []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	img := node.(*ast.Image)
	if err := r.renderLinkOrImage(w, "![", img.Destination, img.Title, enter); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	return ast.WalkContinue, nil
}

// RenderLink renders an *ast.Link node to the given BufWriter.
func (r *Renderer) renderLink(w util.BufWriter, _ []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	link := node.(*ast.Link)
	if err := r.renderLinkOrImage(w, "[", link.Destination, link.Title, enter); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	return ast.WalkContinue, nil
}

// RenderRawHTML renders an *ast.RawHTML node to the given BufWriter.
func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkSkipChildren, nil
	}

	raw := node.(*ast.RawHTML)
	for i := 0; i < raw.Segments.Len(); i++ {
		segment := raw.Segments.At(i)
		if _, err := r.write(w, segment.Value(source)); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkSkipChildren, nil
}

// RenderText renders an *ast.Text node to the given BufWriter.
func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	text := node.(*ast.Text)
	value := text.Segment.Value(source)

	if _, err := r.write(w, value); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	switch {
	case text.HardLineBreak():
		if _, err := r.writeString(w, "\\\n"); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	case text.SoftLineBreak():
		if err := r.writeByte(w, '\n'); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

// RenderString renders an *ast.String node to the given BufWriter.
func (r *Renderer) renderString(w util.BufWriter, _ []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	str := node.(*ast.String)
	if _, err := r.write(w, str.Value); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderTable(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableHeader(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if _, err := r.writeString(w, "| "); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if _, err := r.writeString(w, " |\n|"); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		for x := 0; x < node.ChildCount(); x++ { // use as column count
			if _, err := r.writeString(w, " --- |"); err != nil {
				return ast.WalkStop, fmt.Errorf(": %w", err)
			}
		}

		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableRow(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.openBlock(w, source, node); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if _, err := r.writeString(w, "| "); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	} else {
		if _, err := r.writeString(w, " |"); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}

		if err := r.closeBlock(w); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableCell(w util.BufWriter, _ []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if node.NextSibling() != nil {
			if _, err := r.writeString(w, " | "); err != nil {
				return ast.WalkStop, fmt.Errorf(": %w", err)
			}
		}
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderStrikethrough(w util.BufWriter, _ []byte, _ ast.Node, _ bool) (ast.WalkStatus, error) {
	if _, err := r.writeString(w, "~~"); err != nil {
		return ast.WalkStop, fmt.Errorf(": %w", err)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTaskCheckBox(w util.BufWriter, _ []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		var fill byte = ' '
		if task := node.(*exast.TaskCheckBox); task.IsChecked {
			fill = 'x'
		}

		if _, err := r.write(w, []byte{'[', fill, ']', ' '}); err != nil {
			return ast.WalkStop, fmt.Errorf(": %w", err)
		}
	}

	return ast.WalkContinue, nil
}
