package internal

import (
	"fmt"

	"github.com/firodj/pspsora/binarysearchtree"
)

type SoraBasicBlock struct {
	Address       uint32 `yaml:"address"`
	LastAddress   uint32 `yaml:"last_address"`
	BranchAddress uint32 `yaml:"branch_address"`
}

type BBRefKey struct {
	From  uint32 `yaml:"from"`
	To    uint32 `yaml:"to"`
}

type SoraBBRef struct {
	BBRefKey

	IsDynamic  bool // immediate or by reg/mem/ptr
	IsAdjacent bool // next/prev
	IsLinked   bool // call/linked
	IsVisited  bool // by bbtrace
}

func (ref *SoraBBRef) SetAdjacent(v bool) *SoraBBRef {
	ref.IsAdjacent = true
	return ref
}

type BasicBlockManager struct {
	doc *SoraDocument

	basicBlocks binarysearchtree.AVLTree[uint32, *SoraBasicBlock]
	refs           map[BBRefKey]*SoraBBRef
	refsToBB       map[uint32][]uint32
	refsFromBB		 map[uint32][]uint32
}

func NewBasicBlockManager(doc *SoraDocument) *BasicBlockManager {
	return &BasicBlockManager{
		doc: doc,
		refs: make(map[BBRefKey]*SoraBBRef),
		refsToBB: make(map[uint32][]uint32),
		refsFromBB: make(map[uint32][]uint32),
	}
}

func (bbmanager *BasicBlockManager) Get(addr uint32) (bb *SoraBasicBlock) {
	if addr == 0 {
		return nil
	}

	it := bbmanager.basicBlocks.LowerBound(addr)

	if !it.End() {
		bb = it.Value()

		if addr != bb.Address {
			it := it.Prev()
			if !it.End() {
				bb = it.Value()
			} else {
				bb = nil
			}
		}
	} else {
		it = bbmanager.basicBlocks.Max()
		if !it.End() {
			bb = it.Value()
		}
	}

	if bb != nil {
		if addr <= bb.LastAddress {
			return bb
		}
	}

	return nil
}

func (bbmanager *BasicBlockManager) Create(addr uint32) *SoraBasicBlock {
	bb := bbmanager.Get(addr)
	if bb != nil {
		return nil
	}

	bb = &SoraBasicBlock{
		Address: addr,
	}

	bbmanager.basicBlocks.Insert(addr, bb)
	return bb
}

func (bbmanager *BasicBlockManager) CreateReference(from_addr, to_addr uint32) *SoraBBRef {
	key := BBRefKey{
		From: from_addr,
		To: to_addr,
	}

	if _, ok := bbmanager.refs[key]; ok {
		return bbmanager.refs[key]
	}

	bbref :=  &SoraBBRef{
		BBRefKey: key,
	}
	bbmanager.refsToBB[to_addr] = append(bbmanager.refsToBB[to_addr], from_addr)
	bbmanager.refsFromBB[from_addr] =  append(bbmanager.refsFromBB[from_addr], to_addr)
	bbmanager.refs[key] = bbref;

	return bbref
}

func (bbmanager *BasicBlockManager) SplitAt(split_addr uint32) (prev_bb, split_bb *SoraBasicBlock) {
	prev_bb = bbmanager.Get(split_addr)
	if prev_bb == nil {
		return
	}
	if prev_bb.Address == split_addr {
		return
	}

	last_addr := prev_bb.LastAddress
	if prev_bb.LastAddress  >= split_addr {
		prev_bb.LastAddress = split_addr - 4
	}

	split_bb = bbmanager.Create(split_addr)
	if split_bb == nil {
		prev_bb.LastAddress = last_addr
		fmt.Printf("ERROR:\tunable to create splitted bb at: 0x%08x, possibly exists?\n", split_addr)
		return
	}

	if prev_bb.BranchAddress >= split_bb.Address {
		split_bb.BranchAddress = prev_bb.BranchAddress
		prev_bb.BranchAddress = 0
	}

	split_bb.LastAddress = last_addr
	bbmanager.CreateReference(prev_bb.Address, split_bb.Address).SetAdjacent(true)
	return
}
