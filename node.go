package main

import (
	"fmt"
	"strings"
)

type nodeType int

const (
	nodeDoc nodeType = iota
	nodeParagraph
	nodeHeading
	nodeText
	nodeOrderedList
	nodeUnorderedList
	nodeBold
	nodeItalic
	nodeLink
	nodeImage
)

type node struct {
	parent    *node
	typ       nodeType
	flat      bool
	val       string
	nestLevel int
	openTag   string
	closeTag  string
	nodes     []*node
}

func (n *node) addNode(node *node) {
	node.parent = n
	n.nodes = append(n.nodes, node)
}

func (n *node) addText(t string) {
	n.val += t
}

func (n *node) mergeNode(sibling *node) {
	for _, node := range sibling.nodes {
		n.addNode(node)
	}
}

func (n *node) ToHTML() string {
	ws := ""
	if n.nestLevel > 0 {
		for i := 1; i <= n.nestLevel; i++ {
			ws += "  "
		}
	}

	var html strings.Builder

	if n.typ == nodeDoc {
		for i, node := range n.nodes {
			html.WriteString(node.ToHTML())
			if i < len(n.nodes) {
				html.WriteString("\n")
			}
		}
		return html.String()
	}

	if n.typ == nodeText {
		html.WriteString(n.val)
		return html.String()
	}

	if !n.flat {
		html.WriteString(ws)
		html.WriteString(n.openTag)
		for _, node := range n.nodes {
			html.WriteString("\n")
			html.WriteString(node.ToHTML())
		}
		html.WriteString("\n")
		html.WriteString(ws)
		html.WriteString(n.closeTag)
		return html.String()
	}

	html.WriteString(ws)
	html.WriteString(n.openTag)
	for _, node := range n.nodes {
		html.WriteString(node.ToHTML())
	}
	html.WriteString(ws)
	html.WriteString(n.closeTag)
	return html.String()
}

type docBuilder struct {
	doc           *node
	lines         chan line
	currentParent *node
	complete      chan bool
}

func buildDoc(lines chan line) (*node, chan bool) {
	b := &docBuilder{
		doc: &node{
			parent: nil,
			typ:    nodeDoc,
			flat:   false,
			nodes:  make([]*node, 0),
		},
		lines:    lines,
		complete: make(chan bool),
	}

	b.currentParent = b.doc

	go b.run()

	return b.doc, b.complete
}

func (b *docBuilder) run() {
	for {
		line, more := <-b.lines
		if !more {
			break
		}
		//parse line into nodes
		node := buildNode(line)
		//add nodes to doc or determine correct parent
		if node != nil {
			b.addToDoc(node)
		}
	}
	b.complete <- true
}

func (b *docBuilder) addToDoc(n *node) {
	b.currentParent.addNode(n)
}

func buildNode(l line) *node {
	// check for empty line
	if len(l.items) == 0 {
		return nil
	}

	// check for nesting
	tabCount := 0
	for i := 0; l.items[i].typ == itemTab; i++ {
		tabCount += i
	}

	switch l.items[tabCount].typ {
	case itemHeading:
		return buildHeading(l, tabCount)
	case itemBreak:
		return nil
	case itemEOF:
		return nil
	default:
		return buildParagraph(l, tabCount)
	}
}

func buildHeading(l line, nestLevel int) *node {
	// need to add handling for if a line with more than one item was passed
	header := strings.SplitN(l.items[0].val, " ", 2)
	weight := len(header[0])
	val := header[1]

	text := &node{
		typ:  nodeText,
		flat: true,
		val:  val,
	}

	n := &node{
		typ:       nodeHeading,
		flat:      true,
		nestLevel: nestLevel,
		openTag:   fmt.Sprintf("<h%v>", weight),
		closeTag:  fmt.Sprintf("</h%v>", weight),
		val:       val,
	}

	n.addNode(text)
	return n
}

func buildParagraph(l line, nestLevel int) *node {
	p := &node{
		typ:       nodeParagraph,
		flat:      true,
		nestLevel: nestLevel,
		openTag:   "<p>",
		closeTag:  "</p>",
	}

	for {
		i, node := parseInline(l.items, 0)
		p.addNode(node)
		if i >= len(l.items) {
			break
		}
	}
	return p
}

func parseInline(tokens []item, start int) (int, *node) {
	var node *node

	switch tokens[start].typ {
	case itemStar:
	default:
		node = buildText(tokens[start])
		return start + 1, node
	}
}

func buildText(t item) *node {
	return &node{
		typ:  nodeText,
		flat: true,
	}
}
