package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemHeading
	itemText
)

type item struct {
	typ  itemType
	val  string
	line int
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	// if len(i.val) > 10 {
	// 	return fmt.Sprintf("type: %v, val: %.20q", i.typ, i.val)
	// }
	return fmt.Sprintf("type: %v, val: %q", i.typ, i.val)
}

const eof = -1

const (
	NewLine        = "\n"
	InLineElements = "*_"
)

type stateFn func(*lexer) stateFn

type lexer struct {
	name      string
	input     string
	start     int
	pos       int
	startline int
	line      int
	width     int
	items     chan item
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

func (l *lexer) run() {
	for state := lexBlock; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) accept(valid string) bool {
	// consumes next rune if it is in valid set
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	// consumes runes while valid.
	for {
		if !strings.ContainsRune(valid, l.next()) {
			break
		}
	}
	// when loop breaks lexer is sitting on invalid rune
	l.backup()
}

func (l *lexer) emit(t itemType) {
	i := item{t, l.input[l.start:l.pos], l.startline}
	l.items <- i
	l.start = l.pos
	l.startline = l.line

	// for debugging purposes
	// maybe add a flag for turning on and off
	msg := fmt.Sprintf("Sending %s to parser. \n Currently at %d.", i.String(), l.pos)
	fmt.Println(msg)
}

func lexBlock(l *lexer) stateFn {
	// check for EOF
	if l.pos >= len(l.input) {
		l.width = 0
		l.emit(itemEOF)
		return nil
	}
	switch r := l.input[l.pos]; r {
	case '#':
		return lexHeading
	default:
		return lexText
	}
}

func lexHeading(l *lexer) stateFn {
	// lexer is sitting on a '#' at the start of a new block
	// a space would be valid or up to 5 more #
	// first validate heading
	l.acceptRun("#")
	if l.pos-l.start > 5 {
		// not valid start for heading
		return lexText
	}
	if !l.accept(" ") {
		// not valid start for heading ex:"##adfa"
		return lexText
	}
	// should be in a valid heading at this point
	// absorb text until '\n'
	for {
		if strings.HasPrefix(l.input[l.pos:], NewLine) {
			// lexer is sitting on '\n'
			l.acceptRun("\n\r")
			// lexer should be sitting on the first character
			// of a new block
			if l.pos > l.start {
				l.emit(itemHeading)
				return lexBlock
			}
		}
		if l.next() == eof {
			break
		}
	}
	// reached EOF
	if l.pos > l.start {
		l.emit(itemHeading)
	}
	l.emit(itemEOF)
	return nil
}

func lexText(l *lexer) stateFn {
	// should be sitting on text
	// could be at the start of a block
	// after a newline or after invalid styling
	// should emit a text item at newline
	// or at valid inline styling
	for {
		r := l.next()
		// if strings.ContainsRune(InLineElements, r) {
		// 	// validate inline style first?
		// 	return lexInLineStyles
		// }
		if strings.ContainsRune(NewLine, r) {
			l.acceptRun("\n\r")
			l.emit(itemText)
			return lexBlock
		}
		if r == eof {
			break
		}
	}

	// reached EOF
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func lexInLineStyles(l *lexer) stateFn {
	return nil
}
