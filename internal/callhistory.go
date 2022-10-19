package internal

import (
	"fmt"

	"github.com/firodj/pspsora/binarysearchtree"
)

type BlockGraph struct {
	Address      uint32
	Start, Stop  RefTs
	Fts, FtsStop RefTs
	Text         string
}

func (b *BlockGraph) End(n RefTs) {

}

type StackGraph struct {
	Level       int
	blockGraphs binarysearchtree.AVLTree[RefTs, *BlockGraph]
}

func (s *StackGraph) Size() int {
	return s.blockGraphs.Size()
}

func (s *StackGraph) Last() *BlockGraph {
	it := s.blockGraphs.Max()
	if it.End() {
		return nil
	}
	return it.Value()
}

func (s *StackGraph) Add(n RefTs, addr uint32, text string) *BlockGraph {
	it := s.blockGraphs.Search(n)
	if !it.End() {
		return nil
	}

	b := &BlockGraph{
		Address: addr,
		Start:   n,
		Stop:    n + 1,
		Fts:     0,
		FtsStop: 0,
		Text:    text,
	}

	s.blockGraphs.Insert(n, b)
	return b
}

type CallHistory struct {
	Fts         RefTs
	stackGraphs []*StackGraph
}

func NewCallHistory() *CallHistory {
	c := &CallHistory{
		stackGraphs: []*StackGraph{
			{}, // marker stack
		},
	}
	return c
}

func (c *CallHistory) StackAt(level int) *StackGraph {
	if level < 1 {
		return nil
	}
	for c.MaxLevel() < level {
		s := &StackGraph{}
		c.stackGraphs = append(c.stackGraphs, s)
	}
	return c.stackGraphs[level]
}

func (c *CallHistory) MaxLevel() int {
	return len(c.stackGraphs) - 1
}

func (c *CallHistory) StopAll(maxlevel int, n RefTs) {
	for i := maxlevel; i >= 1; i-- {
		last_block := c.stackGraphs[i].Last()
		if last_block != nil {
			last_block.Stop = n
			last_block.FtsStop = c.Fts
		}
	}
}

func (c *CallHistory) AddMarker(n RefTs, text string) *BlockGraph {
	if text == "" {
		return nil
	}

	marker := c.stackGraphs[0]
	last_block := marker.Last()
	if last_block != nil {
		last_block.Stop = n
		last_block.FtsStop = c.Fts
		if last_block.Text == text {
			return last_block
		}
	}
	b := marker.Add(n, 0, text)
	b.Fts = c.Fts

	return b
}

func (c *CallHistory) AddBlock(level int, n RefTs, addr uint32, text string) *BlockGraph {
	stack_graph := c.StackAt(level)
	if stack_graph == nil {
		return nil
	}

	block := stack_graph.Add(n, addr, text)
	block.Fts = c.Fts
	c.Fts++
	return block
}

func (c *CallHistory) EndBlock(level int, n RefTs) {
	stack_graph := c.StackAt(level)
	if stack_graph == nil {
		return
	}

	last_block := stack_graph.Last()
	if last_block != nil {
		last_block.Stop = n
		last_block.FtsStop = c.Fts
	}
}

func (c *CallHistory) Dump() {
	for i := 0; i < len(c.stackGraphs); i++ {
		s := c.stackGraphs[i]
		fmt.Println("level", i)
		for it := s.blockGraphs.Min(); !it.End(); it = it.Next() {
			b := it.Value()
			fmt.Printf("block [%d-%d] %s %x (%d-%d)\n", b.Start, b.Stop, b.Text, b.Address, b.Fts, b.FtsStop)
		}
	}
}
