package main

import (
	"fmt"
)

func main() {
	input := `##### Valid Heading

testing text input in a single line. and mulitple newlines




testing a multi-line string.
it should emit two string tokens, I think...
# Another Valid Heading

###### Invalid Heading. this should be text.
`
	lexer, tokens := lex("test", input)
	fmt.Printf("%s \n", lexer.name)
	fmt.Printf("Length of input: %d \n", len(lexer.input))

	for {
		_, more := <-tokens
		if !more {
			break
		}
	}
}
