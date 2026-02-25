package lexmd

import (
	"fmt"
)

type line struct {
	num   int
	items []item
}

type block struct {
	num   int
	lines []line
}

func (l line) String() string {
	return fmt.Sprintf("Line Number: %v, Items: %v", l.num, l.items)
}

type blockBuilder struct {
	lineNumber   int
	blockNumber  int
	currentBlock block
	currentLine  []item
	items        chan item
	blocks       chan block
}

func buildBlocks(items chan item) (*blockBuilder, chan block) {
	p := &blockBuilder{
		lineNumber:   0,
		blockNumber:  0,
		currentBlock: block{num: 0, lines: make([]line, 0)},
		currentLine:  make([]item, 0),
		items:        items,
		blocks:       make(chan block),
	}

	go p.run()

	return p, p.blocks
}

func (p *blockBuilder) run() {
	for {
		i, more := <-p.items
		if !more {
			break
		}
		switch i.typ {
		case itemBreak, itemEOF:
			p.addLine()
			p.sendBlock()
		default:
			if i.line != p.lineNumber {
				p.addLine()
				p.lineNumber = i.line
			}
			p.currentLine = append(p.currentLine, i)
		}
	}
	p.sendBlock()
	close(p.blocks)
}

func (p *blockBuilder) addLine() {
	if len(p.currentLine) == 0 {
		return
	}
	l := line{
		num:   p.lineNumber,
		items: p.currentLine,
	}
	p.currentBlock.lines = append(p.currentBlock.lines, l)
	p.currentLine = make([]item, 0)
}

func (p *blockBuilder) sendBlock() {
	p.blocks <- p.currentBlock
	p.blockNumber++
	p.currentBlock = block{num: p.blockNumber, lines: make([]line, 0)}
}
