package lexmd

import (
	"fmt"
	"os"
)

func LexMDFile(md, html string) {
	// open markdown file
	file, _ := os.ReadFile(md)

	// start lexing items is a channel of lexed tokens
	_, items := Lex("testtesttest", string(file))

	// lineBuilder accepts items. groups them into lines
	// and then sends the line on lines channel
	_, blocks := buildBlocks(items)

	// parseNodes accepts lines, parses them into
	// valid nodes and sends them on the node channel
	_, nodes := parseNodes(blocks)

	// accepts nodes and adds them to the doc. sends
	// complete when all channels close.
	doc, complete := buildDoc(nodes)

	<-complete

	final := doc.ToHTML()

	err := os.WriteFile(html, []byte(final), 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func LexMDString(s string) string {
	// start lexing items is a channel of lexed tokens
	_, items := Lex("testtesttest", s)

	// lineBuilder accepts items. groups them into lines
	// and then sends the line on lines channel
	_, blocks := buildBlocks(items)

	// parseNodes accepts lines, parses them into
	// valid nodes and sends them on the node channel
	_, nodes := parseNodes(blocks)

	// accepts nodes and adds them to the doc. sends
	// complete when all channels close.
	doc, complete := buildDoc(nodes)

	<-complete

	html := doc.ToHTML()
	return html
}
