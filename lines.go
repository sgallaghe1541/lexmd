package main

import (
	"fmt"
)

type line struct {
	num   int
	items []item
}

func (l line) String() string {
	return fmt.Sprintf("Line Number: %v, Items: %v", l.num, l.items)
}

type lineParser struct {
	lineNumber  int
	currentLine []item
	items       chan item
	lines       chan line
}

func buildLines(items chan item) (*lineParser, chan line) {
	p := &lineParser{
		lineNumber:  0,
		currentLine: make([]item, 0),
		items:       items,
		lines:       make(chan line),
	}

	go p.run()

	return p, p.lines
}

func (p *lineParser) run() {
	for {
		i, more := <-p.items
		if !more {
			break
		}
		if i.line != p.lineNumber {
			p.sendLine()
			p.lineNumber = i.line
		}
		p.currentLine = append(p.currentLine, i)
	}
	p.sendLine()
	close(p.lines)
}

func (p *lineParser) sendLine() {
	l := line{
		num:   p.lineNumber,
		items: p.currentLine,
	}
	p.lines <- l
	p.currentLine = make([]item, 0)
}
