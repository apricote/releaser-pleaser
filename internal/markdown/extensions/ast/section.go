package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// A Section struct represents a section of elements.
type Section struct {
	gast.BaseBlock
	Name string
}

// Dump implements Node.Dump.
func (n *Section) Dump(source []byte, level int) {
	m := map[string]string{
		"Name": n.Name,
	}
	gast.DumpHelper(n, source, level, m, nil)
}

// KindSection is a NodeKind of the Section node.
var KindSection = gast.NewNodeKind("Section")

// Kind implements Node.Kind.
func (n *Section) Kind() gast.NodeKind {
	return KindSection
}

// NewSection returns a new Section node.
func NewSection(name string) *Section {
	return &Section{Name: name}
}
