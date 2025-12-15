package main

import (
	"strings"
)

type Node struct {
	inLine    bool
	nestLevel int
	openTag   string
	closeTag  string
	nodes     []Node
}

func (n *Node) ToHTML() string {
	ws := ""
	if n.nestLevel > 0 {
		for i := 1; i <= n.nestLevel; i++ {
			ws += "  "
		}
	}
	var html strings.Builder
	if !n.inLine {
		html.WriteString("\n")
	}
	html.WriteString(ws)
	html.WriteString(n.openTag)
	for _, node := range n.nodes {
		html.WriteString(node.ToHTML())
	}
	if !n.inLine {
		html.WriteString("\n")
	}
	html.WriteString(ws)
	html.WriteString(n.closeTag)
	return html.String()
}
