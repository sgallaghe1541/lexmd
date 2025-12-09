package main

import (
	"fmt"
	"os"

	"github.com/sgallaghe1541/lexmd/lexer"
)

func main() {
	file, _ := os.ReadFile("test.md")
	_, tokens := lexer.Lex("testtesttest", string(file))
	for {
		token, more := <-tokens
		fmt.Println(token)
		if !more {
			break
		}
	}
}
