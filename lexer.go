package lexmd

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type itemType int

const (
	itemSpace itemType = iota
	itemText
	itemBang
	itemLParen
	itemRParen
	itemRBracket
	itemHash
	itemEqual
	itemHyph
	itemStar
	itemLBracket
	itemTilde
	itemCaret
	itemBreak
	itemNewLine
	itemHeading
	itemTab
	itemEOF
	itemError
)

// Heading: ==itemHash, ==itemHash || ==itemSpace,
// Italic: ==itemStar, <=itemStar, ==itemStar
// Bold: ==itemStar, ==itemStar, <=itemRBracket, ==itemStar, ==itemStar

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
	return fmt.Sprintf("type: %v, val: %q, line: %v", i.typ, i.val, i.line)
}

const eof = -1

const (
	SPACE       = ' '
	NEWLINE     = '\n'
	HASH        = '#'
	STAR        = '*'
	BANG        = '!'
	LPAREN      = '('
	RPAREN      = ')'
	LBRACKET    = '['
	RBRACKET    = ']'
	TILDE       = '~'
	CARET       = '^'
	HYPH        = '-'
	EQUAL       = '='
	TAB         = '\t'
	GREATERTHAN = '>'
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
	for ; l.peek() == NEWLINE; l.next() {
		l.line++
	}
	if l.pos-l.start > 1 {
		l.emit(itemBreak)
		return lexBlock
	}
	l.emit(itemNewLine)
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
	i := item{t, l.input[l.start:l.pos], l.line}
	if t == itemBreak || t == itemNewLine {
		i.line--
	}
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
	// 	case HASH:
	// 		return lexHeading
	case NEWLINE:
		return lexNewLines
	default:
		return lexText
	}
}

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

func lexTilde(l *lexer) stateFn {
	// lexer is sitting on an '~'
	l.next()
	l.emit(itemTilde)
	return lexText
}

func lexCaret(l *lexer) stateFn {
	l.next()
	l.emit(itemCaret)
	return lexText
}

func lexHyph(l *lexer) stateFn {
	l.next()
	l.emit(itemHyph)
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

func lexSpace(l *lexer) stateFn {
	l.next()
	l.emit(itemSpace)
	return lexText
}

func lexHash(l *lexer) stateFn {
	l.next()
	l.emit(itemHash)
	return lexText
}

func lexEqual(l *lexer) stateFn {
	l.next()
	l.emit(itemEqual)
	return lexText
}

func lexText(l *lexer) stateFn {
	// should be sitting on text
	// could be at the start of a block
	// after a newline or after invalid styling
	// should emit a text item at newline
	// or at valid inline styling
	for {
		switch l.peek() {
		case NEWLINE:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBlock
		case SPACE:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexSpace
		case STAR:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexStar
		case BANG:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBang
		case LPAREN:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLParen
		case RPAREN:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexRParen
		case LBRACKET:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLBracket
		case RBRACKET:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexRBracket
		case HASH:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexHash
		case TILDE:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexTilde
		case CARET:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexCaret
		case HYPH:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexHyph
		case EQUAL:
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexEqual
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
