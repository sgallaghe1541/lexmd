package main

import (
	"fmt"
)

func main() {
	input := `##### Valid Heading
    `
	lexer, tokens := lex("test", input)
	fmt.Printf("%s \n", lexer.name)

	for {
		t, more := <-tokens
		if !more {
			break
		}
		fmt.Print(t)
	}
}
