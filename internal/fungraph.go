package internal

import (
	"fmt"
	"strings"

	"github.com/firodj/pspsora/binarysearchtree"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type FunGraphNodeID int

type FunGraphNode struct {
	ID       FunGraphNodeID
	ParentID FunGraphNodeID
	Subs     *orderedmap.OrderedMap[uint32, FunGraphNodeID]

	Address  uint32
	Fun      *SoraFunction
	Count    int64
	Duration int64
}

type FunGraph struct {
	root  *FunGraphNode
	nodes *binarysearchtree.AVLTree[FunGraphNodeID, *FunGraphNode]
	index FunGraphNodeID
}

func NewFunGraph() *FunGraph {
	g := &FunGraph{
		index: 0,
		root: &FunGraphNode{
			ParentID: -1,
			Subs:     orderedmap.New[uint32, FunGraphNodeID](),
		},
		nodes: new(binarysearchtree.AVLTree[FunGraphNodeID, *FunGraphNode]),
	}
	g.nodes.Insert(g.root.ID, g.root)
	g.index++

	return g
}

func (g *FunGraph) At(id FunGraphNodeID) *FunGraphNode {
	it := g.nodes.Search(id)
	if it.End() {
		return nil
	}
	return it.Value()
}

func (g *FunGraph) AddNode(func_addr uint32, parent_ID FunGraphNodeID) *FunGraphNode {
	items := g.At(parent_ID).Subs

	if item_ID, ok := items.Get(func_addr); ok {
		node := g.At(item_ID)
		node.Count++
		return node
	}

	node := &FunGraphNode{
		ID:       g.index,
		ParentID: parent_ID,
		Subs:     orderedmap.New[uint32, FunGraphNodeID](),
		Count:    1,
	}

	g.nodes.Insert(node.ID, node)
	items.Set(func_addr, node.ID)
	g.index++

	return node
}

func (g *FunGraph) DumpNode(level int, parent_ID FunGraphNodeID) {
	items := g.At(parent_ID).Subs

	for pair := items.Oldest(); pair != nil; pair = pair.Next() {
		node := g.At(pair.Value)
		fmt.Print(strings.Repeat(" ", level))
		fmt.Print("+ ")

		if node.Fun != nil {
			fmt.Printf("%s (%d)", node.Fun.Name, node.Duration)
		} else {
			fmt.Printf("0x%08x (%d)", node.Address, node.Duration)
		}

		fmt.Println()
		g.DumpNode(level+1, node.ID)
	}

}
