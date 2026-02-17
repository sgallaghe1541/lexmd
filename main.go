package main

import (
	"fmt"
	"os"
)

func main() {
	// open markdown file
	file, _ := os.ReadFile("test.md")

	// start lexing items is a channel of lexed tokens
	_, items := Lex("testtesttest", string(file))

	// lineBuilder accepts items. groups them into lines
	// and then sends the line on lines channel
	_, blocks := buildBlocks(items)

	// parseNodes accepts lines, parses them into
	// valid nodes and sends them on the node channel
	_, nodes := ParseNodes(blocks)

	// accepts nodes and adds them to the doc. sends
	// complete when all channels close.
	doc, complete := buildDoc(nodes)

	<-complete

	html := doc.ToHTML()
	fmt.Print(html)

	err := os.WriteFile("test.html", []byte(html), 0644)
	if err != nil {
		fmt.Println(err)
	}

	// for {
	// 	line, more := <-lines
	// 	if !more {
	// 		break
	// 	}
	// 	fmt.Println(line)
	// }
}
