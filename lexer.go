package lexmd

import (
	"fmt"
	"strings"
)

type itemType int

const (
	itemError itemType = iota
	itemNewLine
	itemDoubleNewLine
	itemSpace
	itemTab
	itemGrave
	itemHash
	itemDoubleHash
	itemTripleHash
	itemCarrot
	itemDash
	itemTilde
	itemStar
	itemDoubleStar
	itemX
	itemBang
	itemLBrack
	itemRBrack
	itemLParen
	itemRParen
	itemLBrace
	itemRBrace
	itemGreater
	itemColon
	itemDoubleEqual
	itemText
	itemEOF
)

const puncToEscape = "!\"#$%&'()*+,-./:;<=>?@[]\\^_`{|}~"

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan item
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}
