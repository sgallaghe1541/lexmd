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
	itemBreak
	itemHeading
	itemText
	itemStar
	itemBang
	itemLParen
	itemRParen
	itemLBracket
	itemRBracket
)

type item struct {
	typ    itemType
	val    string
	line   int
	nested int
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
	return fmt.Sprintf("type: %v, val: %q, line: %v, nested: %v", i.typ, i.val, i.line, i.nested)
}

const eof = -1

const (
	NewLine     = '\n'
	Hash        = '#'
	Star        = '*'
	Bang        = '!'
	LParen      = '('
	RParen      = ')'
	LBracket    = '['
	RBracket    = ']'
	Hyph        = '-'
	Equal       = '='
	Tab         = '\t'
	GreaterThan = '>'
)

type stateFn func(*lexer) stateFn

type lexer struct {
	name      string
	input     string
	start     int
	pos       int
	startline int
	line      int
	nested    int
	width     int
	items     chan item
}

func Lex(name, input string) (*lexer, chan item) {
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

func (l *lexer) peek() (r rune) {
	r, _ = utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
	l.width = 0
}

func (l *lexer) accept(valid string) bool {
	// consumes next rune if it is in valid set
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func lexNewLines(l *lexer) stateFn {
	for ; l.peek() == NewLine; l.next() {
		l.line++
	}
	if l.pos-l.start > 1 {
		l.emit(itemBreak)
		return lexBlock
	}
	l.ignore()
	return lexBlock
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
	i := item{t, l.input[l.start:l.pos], l.line, l.nested}
	l.items <- i
	l.start = l.pos
	l.startline = l.line

	// for debugging purposes
	// maybe add a flag for turning on and off
	// msg := fmt.Sprintf("Sending %s to parser. \n Currently at %d.", i.String(), l.pos)
	// fmt.Println(msg)
}

func (l *lexer) checkEOF() bool {
	if l.pos >= len(l.input) {
		l.width = 0
		l.emit(itemEOF)
		return true
	}
	return false
}

func lexBlock(l *lexer) stateFn {
	// check for EOF
	if l.checkEOF() {
		return nil
	}
	switch r := l.input[l.pos]; r {
	case Hash:
		return lexHeading
	case NewLine:
		return lexNewLines
	default:
		return lexText
	}
}

// func lexInLineStyles(l *lexer) stateFn {
// 	// check for EOF
// 	if l.checkEOF() {
// 		return nil
// 	}
//
// 	switch r := l.input[l.pos]; r {
// 	case '*':
// 		return lexStars
// 	default:
// 		return lexText
// 	}
// }

func lexStar(l *lexer) stateFn {
	// lexer is sitting on an '*'
	// advance lexer and emit itemStar
	l.next()
	l.emit(itemStar)
	return lexText
}

func lexBang(l *lexer) stateFn {
	// lexer is sitting on an '!'
	// advance lexer and emit itemBang
	l.next()
	l.emit(itemBang)
	return lexText
}

func lexLParen(l *lexer) stateFn {
	l.next()
	l.emit(itemLParen)
	return lexText
}

func lexRParen(l *lexer) stateFn {
	l.next()
	l.emit(itemRParen)
	return lexText
}

func lexLBracket(l *lexer) stateFn {
	l.next()
	l.emit(itemLBracket)
	return lexText
}

func lexRBracket(l *lexer) stateFn {
	l.next()
	l.emit(itemRBracket)
	return lexText
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
		if l.peek() == NewLine {
			// lexer is sitting on '\n'
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
		switch l.peek() {
		case NewLine:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBlock
		case Star:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexStar
		case Bang:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBang
		case LParen:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLParen
		case RParen:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexRParen
		case LBracket:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLBracket
		case RBracket:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexRBracket
		}

		if l.next() == eof {
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
