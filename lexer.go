package lexmd

import (
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

const eof = -1

const (
    NewLine = "\n"
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

func (l *lexer) next() (r rune) {
	if l.pos > len(l.input) {
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
	for strings.ContainsRune(valid, l.next()) {
	}
	// when loop breaks lexer is sitting on invalid rune
	l.backup()
}

func (l *lexer) absorbNewLines() {
	// helper function to absorb contiguous verticle whitespce
    for strings.ContainsRune("\n\r", l.next()) {
    }
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos], l.startline}
	l.start = l.pos
	l.startline = l.line
}

func lexBlock(l *lexer) stateFn {
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
            l.absorbNewLines()
            // lexer should be sitting on the first character
            // of a new block
            if l.pos > l.start {
                l.emit(itemHeading)
                return lexBlock
            }
        if l.next() == eof { break }
    }
    // reached EOF
    if l.pos > l.start {
        l.emit(itemHeading)
    }
    l.emit(itemEOF)
    return nil
}

func lexText(l *lexer) stateFn {
	return nil
}
