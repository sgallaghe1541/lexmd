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
	nodeAltHeading1
	nodeAltHeading2
	nodeText
	nodeUList
	nodeUListItem
	nodeOList
	nodeOListItem
	nodeBold
	nodeItalic
	nodeStrikeThrough
	nodeSubScript
	nodeSuperScript
	nodeLink
	nodeImage
	nodeBreak
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
	myTokens  []item
	tokens    []item
}

type nodeBuilder func([]item) (*node, []item)

func buildHashHeading(items []item) (*node, []item) {
	hashCount := checkHashHeading(items)
	if hashCount == 0 {
		return nil, items
	}

	n := &node{
		typ:  nodeHeading,
		flat: true,
	}

	n.myTokens = items[:hashCount+1]
	n.tokens = items[hashCount+1:]

	n.openTag = fmt.Sprintf("<h%d>", hashCount)
	n.closeTag = fmt.Sprintf("</h%d>", hashCount)

	return n, make([]item, 0)
}

func buildParagraph(items []item) (*node, []item) {
	n := &node{
		typ:      nodeParagraph,
		flat:     true,
		openTag:  "<p>",
		closeTag: "</p>",
		tokens:   items,
	}
	return n, make([]item, 0)
}

func buildUList(items []item) (*node, []item) {
	n, _ := buildUListItem(items)
	if n == nil {
		return nil, make([]item, 0)
	}

	l := &node{
		typ:      nodeUList,
		openTag:  "<ul>",
		closeTag: "</ul>",
	}

	l.addNode(n)

	return l, make([]item, 0)
}

func buildOList() *node {
	return &node{
		typ:      nodeOList,
		openTag:  "<ol>",
		closeTag: "</ol>",
	}
}

func buildBoldItalic(items []item) (*node, []item) {
	tokens := []itemType{itemStar, itemStar, itemStar}

	//not enough tokens to make a valid node
	// if len(tokens)*2 >= len(items) {
	// 	_, items := demote(items, []int{0})
	// 	return nil, items
	// }
	openStart, openStop := find(items, tokens)
	// should always be 0,2

	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	// starting tokens are valid but followed by space
	if items[openStop+1].typ == itemSpace {
		err, items := demote(items, []int{openStart, openStop - 1, openStop})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	closeStart, closeStop := find(items[openStop+1:], tokens)
	// find failed to find ending tokens
	if closeStart == -1 {
		items[openStart].typ = itemText
		return nil, items
	}

	closeStart += len(tokens)
	closeStop += len(tokens)

	// close tokens are valid but preceded by a space
	if items[closeStart-1].typ == itemSpace {
		err, items := demote(items, []int{openStart, openStop - 1, openStop, closeStart, closeStop - 1, closeStop})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	// yes, its janky. only making a bold node and then
	// passing it the rest of the tokens to make a child
	// italic node
	n := &node{
		typ:      nodeBold,
		openTag:  "<strong>",
		closeTag: "</strong>",
		flat:     true,
	}

	myTokens := []item{items[openStart], items[openStart+1], items[closeStop-1], items[closeStop]}
	n.myTokens = myTokens
	n.tokens = items[openStop : closeStart+1]

	if closeStop >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[closeStop+1:]
}

func buildBold(items []item) (*node, []item) {
	tokens := []itemType{itemStar, itemStar}

	openStart, openStop := find(items, tokens)
	// should always be 0,1

	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	// starting tokens are valid but followed by space
	if items[openStop+1].typ == itemSpace {
		err, items := demote(items, []int{openStart, openStop})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	closeStart, closeStop := find(items[openStop+1:], tokens)
	// find failed to find ending tokens
	if closeStart == -1 {
		items[openStart].typ = itemText
		return nil, items
	}

	closeStart += len(tokens)
	closeStop += len(tokens)

	// close tokens are valid but preceded by a space
	if items[closeStart-1].typ == itemSpace {
		err, items := demote(items, []int{openStart, openStop, closeStart, closeStop})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	n := &node{
		typ:      nodeBold,
		openTag:  "<strong>",
		closeTag: "</strong>",
		flat:     true,
	}

	myTokens := []item{items[openStart], items[openStop], items[closeStart], items[closeStop]}
	n.myTokens = myTokens
	n.tokens = items[openStop+1 : closeStart]

	if closeStop >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[closeStop+1:]
}

func buildItalic(items []item) (*node, []item) {
	tokens := []itemType{itemStar}

	//not enough tokens to make a valid node
	if len(tokens)*2 >= len(items) {
		_, items := demote(items, []int{0})
		return nil, items
	}

	openStart, openStop := find(items, tokens)
	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	// starting tokens are valid but followed by space
	if items[openStop+1].typ == itemSpace {
		err, items := demote(items, []int{openStart})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	closeStart, closeStop := find(items[openStop+1:], tokens)
	// find failed to find ending tokens
	if closeStart == -1 {
		items[openStart].typ = itemText
		return nil, items
	}

	closeStart += len(tokens)
	closeStop += len(tokens)

	// close tokens are valid but preceded by a space
	if items[closeStart-1].typ == itemSpace {
		err, items := demote(items, []int{openStart, closeStart})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	n := &node{
		typ:      nodeItalic,
		openTag:  "<em>",
		closeTag: "</em>",
		flat:     true,
	}

	myTokens := []item{items[openStart], items[closeStart]}
	n.myTokens = myTokens
	n.tokens = items[openStop+1 : closeStart]

	if closeStop >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[closeStop+1:]
}

func buildUListItem(items []item) (*node, []item) {
	tokens := []itemType{itemHyph, itemSpace}

	openStart, openStop := find(items, tokens)
	// should always be 0,1

	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	n := &node{
		typ:       nodeUListItem,
		openTag:   "<li>",
		closeTag:  "</li>",
		flat:      true,
		nestLevel: 1,
	}

	myTokens := []item{items[openStart], items[openStop]}
	n.myTokens = myTokens
	n.tokens = items[openStop+1:]

	return n, make([]item, 0)
}

func buildAltHeading1(items []item) (*node, []item) {
	if len(items) < 2 {
		return nil, items
	}

	if !checkAll(items, itemEqual) {
		return nil, items
	}

	// find failed to find starting tokens
	// at the beginning of items
	n := &node{
		typ:  nodeAltHeading1,
		flat: true,
	}

	n.tokens = items

	return n, make([]item, 0)
}

func buildStrikeThrough(items []item) (*node, []item) {
	tokens := []itemType{itemTilde, itemTilde}

	openStart, openStop := find(items, tokens)
	// should always be 0,1

	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	// starting tokens are valid but followed by space
	if items[openStop+1].typ == itemSpace {
		err, items := demote(items, []int{openStart, openStop})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	closeStart, closeStop := find(items[openStop+1:], tokens)
	// find failed to find ending tokens
	if closeStart == -1 {
		items[openStart].typ = itemText
		return nil, items
	}

	closeStart += len(tokens)
	closeStop += len(tokens)

	// close tokens are valid but preceded by a space
	if items[closeStart-1].typ == itemSpace {
		err, items := demote(items, []int{openStart, openStop, closeStart, closeStop})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	n := &node{
		typ:      nodeStrikeThrough,
		openTag:  "<s>",
		closeTag: "</s>",
		flat:     true,
	}

	myTokens := []item{items[openStart], items[openStop], items[closeStart], items[closeStop]}
	n.myTokens = myTokens
	n.tokens = items[openStop+1 : closeStart]

	if closeStop >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[closeStop+1:]
}

func buildSubScript(items []item) (*node, []item) {
	tokens := []itemType{itemTilde}

	//not enough tokens to make a valid node
	if len(tokens)*2 >= len(items) {
		_, items := demote(items, []int{0})
		return nil, items
	}

	openStart, openStop := find(items, tokens)
	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	// starting tokens are valid but followed by space
	if items[openStop+1].typ == itemSpace {
		err, items := demote(items, []int{openStart})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	closeStart, closeStop := find(items[openStop+1:], tokens)
	// find failed to find ending tokens
	if closeStart == -1 {
		items[openStart].typ = itemText
		return nil, items
	}

	closeStart += len(tokens)
	closeStop += len(tokens)

	// close tokens are valid but preceded by a space
	if items[closeStart-1].typ == itemSpace {
		err, items := demote(items, []int{openStart, closeStart})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	n := &node{
		typ:      nodeSubScript,
		openTag:  "<sub>",
		closeTag: "</sub>",
		flat:     true,
	}

	myTokens := []item{items[openStart], items[closeStart]}
	n.myTokens = myTokens
	n.tokens = items[openStop+1 : closeStart]

	if closeStop >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[closeStop+1:]
}

func buildSuperScript(items []item) (*node, []item) {
	tokens := []itemType{itemCaret}

	//not enough tokens to make a valid node
	if len(tokens)*2 >= len(items) {
		_, items := demote(items, []int{0})
		return nil, items
	}

	openStart, openStop := find(items, tokens)
	// find failed to find starting tokens
	// at the beginning of items
	if openStart != 0 {
		return nil, items
	}

	// starting tokens are valid but followed by space
	if items[openStop+1].typ == itemSpace {
		err, items := demote(items, []int{openStart})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	closeStart, closeStop := find(items[openStop+1:], tokens)
	// find failed to find ending tokens
	if closeStart == -1 {
		items[openStart].typ = itemText
		return nil, items
	}

	closeStart += len(tokens)
	closeStop += len(tokens)

	// close tokens are valid but preceded by a space
	if items[closeStart-1].typ == itemSpace {
		err, items := demote(items, []int{openStart, closeStart})
		if err != nil {
			fmt.Println(err)
		}
		return nil, items
	}

	n := &node{
		typ:      nodeSuperScript,
		openTag:  "<sup>",
		closeTag: "</sup>",
		flat:     true,
	}

	myTokens := []item{items[openStart], items[closeStart]}
	n.myTokens = myTokens
	n.tokens = items[openStop+1 : closeStart]

	if closeStop >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[closeStop+1:]
}

func buildLink(items []item) (*node, []item) {
	if len(items) == 0 {
		return nil, items
	}
	if items[0].typ != itemLBracket {
		return nil, items
	}
	if len(items) < 6 {
		_, items = demote(items, []int{0})
		return nil, items
	}

	rBracket, _ := find(items, []itemType{itemRBracket})
	if rBracket == -1 {
		_, items = demote(items, []int{0})
		return nil, items
	}

	if len(items) <= rBracket+1 {
		_, items = demote(items, []int{0, rBracket})
		return nil, items
	}
	if items[rBracket+1].typ != itemLParen {
		_, items = demote(items, []int{0, rBracket})
		return nil, items
	}

	rParen, _ := find(items, []itemType{itemRParen})

	if rParen == -1 {
		_, items = demote(items, []int{0, rBracket, rBracket + 1})
		return nil, items
	}

	link := ""
	for _, item := range items[rBracket+2 : rParen] {
		link += item.val
	}

	n := &node{
		typ:      nodeLink,
		openTag:  fmt.Sprintf("<a href=\"%s\" target=\"_blank\">", link),
		closeTag: "</a>",
		tokens:   items[1:rBracket],
		myTokens: items,
		flat:     true,
	}

	if rParen >= len(items)-1 {
		return n, make([]item, 0)
	}
	return n, items[rParen+1:]
}

func buildImage(items []item) (*node, []item) {
	return nil, items
}

func buildText(items []item) (*node, []item) {
	if len(items) == 0 {
		return nil, items
	}

	n := &node{
		typ: nodeText,
		val: "",
	}

	i := 0
	for _, item := range items {
		if item.typ >= itemStar {
			break
		}
		n.addText(item.val)
		i++
	}
	if i == 0 {
		return nil, items
	}
	//this is a *test

	if len(items) == i {
		n.myTokens = items
		return n, make([]item, 0)
	}

	n.myTokens = items[:i]
	return n, items[i:]
}

var validBuilders = map[nodeType][]nodeBuilder{
	nodeHeading: []nodeBuilder{buildText},
	nodeParagraph: []nodeBuilder{
		buildBoldItalic,
		buildBold,
		buildItalic,
		buildStrikeThrough,
		buildSubScript,
		buildSuperScript,
		buildLink,
		buildText,
	},
	nodeUListItem: []nodeBuilder{
		buildBoldItalic,
		buildBold,
		buildItalic,
		buildStrikeThrough,
		buildSubScript,
		buildSuperScript,
		buildText,
	},
	nodeBold: []nodeBuilder{
		buildItalic,
		buildStrikeThrough,
		buildSubScript,
		buildSuperScript,
		buildText,
	},
	nodeItalic: []nodeBuilder{
		buildBold,
		buildStrikeThrough,
		buildSubScript,
		buildSuperScript,
		buildText,
	},
	nodeStrikeThrough: []nodeBuilder{
		buildBold,
		buildItalic,
		buildSubScript,
		buildSuperScript,
		buildText,
	},
	nodeSubScript: []nodeBuilder{
		buildItalic,
		buildStrikeThrough,
		buildText,
	},
	nodeSuperScript: []nodeBuilder{
		buildItalic,
		buildStrikeThrough,
		buildText,
	},
	nodeLink: []nodeBuilder{
		buildText,
	},
	nodeOListItem: []nodeBuilder{
		buildBoldItalic,
		buildBold,
		buildItalic,
		buildText,
	},
}

func (n *node) findChildren() {
	builders := validBuilders[n.typ]
	if len(builders) == 0 {
		return
	}
	for {
		if len(n.tokens) == 0 {
			break
		}
		for _, b := range builders {
			node, items := b(n.tokens)
			if node != nil {
				// node.findChildren()
				n.addNode(node)
				n.tokens = items
			}
		}
	}
}

type nodeParser struct {
	lastNodeTyp nodeType
	currentNode *node
	blocks      chan block
	nodes       chan *node
}

func ParseNodes(blocks chan block) (*nodeParser, chan *node) {
	p := &nodeParser{
		lastNodeTyp: nodeDoc,
		blocks:      blocks,
		nodes:       make(chan *node),
	}
	go p.run()

	return p, p.nodes
}

func (p *nodeParser) run() {
	for {
		b, more := <-p.blocks
		if !more {
			break
		}
		node := buildBlockNode(b)
		if node != nil {
			p.nodes <- node
		}
	}
	close(p.nodes)
}

func buildBlockNode(b block) *node {
	// build a node from each line
	// if they are all the same, create the
	// appropriate parent and merge all the nodes.
	builders := []nodeBuilder{
		buildHashHeading,
		buildUList,
		buildParagraph,
	}
	if len(b.lines) < 1 {
		return nil
	}

	var n *node

	for _, builder := range builders {
		n, _ = builder(b.lines[0].items)
		if n != nil {
			break
		}
	}

	switch n.typ {
	case nodeUList:
		for i := 1; i < len(b.lines); i++ {
			c, _ := buildUListItem(b.lines[i].items)
			n.addNode(c)
		}
	default:
		for i := 1; i < len(b.lines); i++ {
			n.tokens = append(n.tokens, b.lines[i].items...)
		}
	}

	return n
}

// func buildLineNode(items []item) (*node, []item) {
// 	// i don't think we'll get here
// 	// but just in case
// 	if len(items) == 0 {
// 		return &node{typ: nodeBreak}, items
// 	}
//
// 	if items[0].typ == itemBreak || items[0].typ == itemEOF {
// 		return &node{typ: nodeBreak}, items
// 	}
// 	// add additional checks for lists
// 	// block quotes, etc.
//
// 	builders := []nodeBuilder{
// 		buildHashHeading,
// 		buildAltHeading1,
// 		buildUListItem,
// 		buildParagraph, //needs to be the last builder
// 	}
// 	for _, b := range builders {
// 		node, items := b(items)
// 		if node != nil {
// 			return node, items
// 		}
// 	}
// 	return nil, items
// }

func checkHashHeading(items []item) int {
	// check if valid hash heading exists
	// return hashcount if true or
	// 0 if false
	hashCount := 0
	for _, item := range items {
		if item.typ != itemHash {
			break
		}
		hashCount++
	}

	if hashCount < 1 || hashCount > 6 {
		return 0
	}

	if len(items) < hashCount+2 {
		return 0
	}

	if items[hashCount].typ == itemSpace {
		return hashCount
	}

	return 0
}

func find(items []item, tokens []itemType) (int, int) {
	if len(tokens) >= len(items) {
		return -1, -1
	}
	check := len(items) - len(tokens)
	for i := 0; i <= check; i++ {
		for j := 0; j < len(tokens); j++ {
			if items[i+j].typ != tokens[j] {
				break
			}
			if j == len(tokens)-1 {
				return i, i + len(tokens) - 1
			}
		}
	}
	return -1, -1
}

func checkAll(items []item, t itemType) bool {
	for _, item := range items {
		if item.typ != t {
			return false
		}
	}
	return true
}

func demote(items []item, indices []int) (error, []item) {
	for _, index := range indices {
		if index >= len(items) {
			err := fmt.Errorf("Index %d is outside bounds of items %d", index, len(items))
			return err, items
		}
		items[index].typ = itemText
	}
	return nil, items
}

func demoteAltHeading(n *node) {
	if n.typ != nodeAltHeading1 && n.typ != nodeAltHeading2 {
		return
	}
	n.typ = nodeParagraph
	n.openTag = "<p>"
	n.closeTag = "</p>"
	n.flat = true

	for _, item := range n.tokens {
		item.typ = itemText
	}

	return
}

func mergeAltHeading(p *node, h *node) {
	if p.typ != nodeParagraph {
		fmt.Println("We shouldn't be trying to make this a heading")
	}
	// check children. they all need to be text
	for _, n := range p.nodes {
		if n.typ != nodeText {
			return
		}
	}

	p.typ = h.typ
	if h.typ == nodeAltHeading1 {
		p.openTag = "<h1>"
		p.closeTag = "</h1>"
	} else {
		p.openTag = "<h2>"
		p.closeTag = "</h2>"
	}
	p.tokens = append(p.tokens, h.tokens...)

	return
}

func (n *node) addNode(node *node) {
	node.parent = n
	node.findChildren()
	n.nodes = append(n.nodes, node)
}

func (n *node) addText(t string) {
	n.val += t
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
	// html.WriteString(ws)
	html.WriteString(n.closeTag)
	return html.String()
}

type docBuilder struct {
	doc           *node
	nodes         chan *node
	currentParent *node
	complete      chan bool
}

func buildDoc(nodes chan *node) (*node, chan bool) {
	b := &docBuilder{
		doc: &node{
			parent: nil,
			typ:    nodeDoc,
			flat:   false,
			nodes:  make([]*node, 0),
		},
		nodes:    nodes,
		complete: make(chan bool),
	}

	b.currentParent = b.doc

	go b.run()

	return b.doc, b.complete
}

func (b *docBuilder) run() {
	for {
		node, more := <-b.nodes
		if !more {
			break
		}

		// 	switch node.typ {
		// 	case nodeBreak:
		// 		b.currentParent = b.doc
		// 	case nodeParagraph:
		// 		if b.currentParent.typ == nodeParagraph {
		// 			node.findChildren()
		// 			for _, n := range node.nodes {
		// 				b.currentParent.nodes = append(b.currentParent.nodes, n)
		// 			}
		// 		} else {
		// 			if b.currentParent != b.doc {
		// 				b.currentParent = b.doc
		// 			}
		// 			b.addToDoc(node)
		// 			b.currentParent = b.currentParent.nodes[len(b.currentParent.nodes)-1]
		// 		}
		// 	case nodeHeading:
		// 		b.currentParent = b.doc
		// 		b.addToDoc(node)
		// 	case nodeAltHeading1:
		// 		if b.currentParent == b.doc {
		// 			if b.currentParent.nodes[len(b.currentParent.nodes)-1].typ == nodeParagraph {
		// 				mergeAltHeading(b.currentParent.nodes[len(b.currentParent.nodes)-1], node)
		// 			} else {
		// 				b.currentParent = b.doc
		// 				demoteAltHeading(node)
		// 				b.addToDoc(node)
		// 			}
		// 		}
		// 	case nodeUListItem:
		// 		if b.currentParent.typ != nodeUList {
		// 			b.currentParent = b.doc
		// 			b.addToDoc(buildUList())
		// 			b.currentParent = b.currentParent.nodes[len(b.currentParent.nodes)-1]
		// 		}
		// 		b.addToDoc(node)
		// 	}
		b.addToDoc(node)
	}
	b.complete <- true
}

func (b *docBuilder) addToDoc(n *node) {
	b.currentParent.addNode(n)
}
