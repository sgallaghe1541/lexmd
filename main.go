package main

import (
	"fmt"
)

func main() {
	input := `##### Valid Heading

testing text input in a single line.

# Another Valid Heading

###### Invalid Heading. this should be text.
`
	lexer, tokens := lex("test", input)
	fmt.Printf("%s \n", lexer.name)

	for {
		_, more := <-tokens
		if !more {
			break
		}
	}
}
