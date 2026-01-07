package main

import (
	"fmt"
	"os"
)

func main() {
	file, _ := os.ReadFile("test.md")
	_, items := Lex("testtesttest", string(file))
	_, lines := buildLines(items)
	doc, complete := buildDoc(lines)

	<-complete

	fmt.Print(doc.ToHTML())

	// for {
	// 	line, more := <-lines
	// 	if !more {
	// 		break
	// 	}
	// 	fmt.Println(line)
	// }
}
