package main

import (
	"fmt"
)

type blockType int

const (
	blockEmpty blockType = iota
	blockHeading
	blockP
	blockUL
	blockOL
	blockQuote
	blockNot
)

type block struct {
	typ    blockType
	nested int
	items  []item
}

func (b *block) String() string {
	return fmt.Sprintf("Block type: %v, items: %s", b.typ, b.items)
}

func (b *block) addItem(i item) {
	b.items = append(b.items, i)
}

type blockParser struct {
	items   chan item
	blocks  chan *block
	current *block
}

func (p *blockParser) run() {
	for {
		item, more := <-p.items
		if !more {
			break
		}
		switch item.typ {
		case itemBreak:
			p.checktoSendBlock()
		case itemHeading:
			p.checktoSendBlock()
			p.current = newBlock(blockHeading, item)
			p.sendBlock()
		case itemEOF:
			break
		default:
			if p.current == nil {
				p.current = newBlock(blockP, item)
			} else {
				p.current.addItem(item)
			}
		}
	}
	p.checktoSendBlock()
	close(p.blocks)
}

func (p *blockParser) checktoSendBlock() {
	if p.current != nil {
		p.sendBlock()
	}
}

func (p *blockParser) sendBlock() {
	p.blocks <- p.current
	p.current = nil
}

func newBlock(t blockType, i item) *block {
	return &block{
		typ:    t,
		nested: i.nested,
		items:  []item{i},
	}
}

func parseBlocks(items chan item) (*blockParser, chan *block) {
	p := &blockParser{
		items:   items,
		blocks:  make(chan *block),
		current: nil,
	}

	go p.run()

	return p, p.blocks
}
