package main

import (
	"fmt"
	"os"
)

func main() {
	file, _ := os.ReadFile("test.md")
	_, tokens := Lex("testtesttest", string(file))
	_, blocks := parseBlocks(tokens)
	for {
		block, more := <-blocks
		if !more {
			break
		}
		fmt.Println(block)
	}
}
