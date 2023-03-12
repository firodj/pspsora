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

func (bb *SoraBasicBlock) Size() uint32 {
	return bb.LastAddress - bb.Address + 4
}

type BBRefKey struct {
	From uint32 `yaml:"from"`
	To   uint32 `yaml:"to"`
}

type SoraBBRef struct {
	BBRefKey

	IsDynamic  bool // TODO: immediate or by reg/mem/ptr
	IsAdjacent bool // next/prev
	IsLinked   bool // call/linked
	IsVisited  bool // TODO: by bbtrace
}

func (ref *SoraBBRef) SetAdjacent(v bool) *SoraBBRef {
	ref.IsAdjacent = true
	return ref
}

func (ref *SoraBBRef) SetLinked(v bool) *SoraBBRef {
	ref.IsLinked = true
	return ref
}

func (ref *SoraBBRef) String() string {
	s := fmt.Sprintf("0x%0x -> 0x%x", ref.From, ref.To)
	if ref.IsAdjacent {
		s = s + " a"
	}
	if ref.IsLinked {
		s = s + " l"
	}
	return s
}

type BasicBlockManager struct {
	doc *SoraDocument

	basicBlocks binarysearchtree.AVLTree[uint32, *SoraBasicBlock]
	refs        map[BBRefKey]*SoraBBRef
	refsToBB    map[uint32][]uint32
	refsFromBB  map[uint32][]uint32
}

func NewBasicBlockManager(doc *SoraDocument) *BasicBlockManager {
	return &BasicBlockManager{
		doc:        doc,
		refs:       make(map[BBRefKey]*SoraBBRef),
		refsToBB:   make(map[uint32][]uint32),
		refsFromBB: make(map[uint32][]uint32),
	}
}

func (bbmanager *BasicBlockManager) Get(addr uint32) (bb *SoraBasicBlock) {
	if addr == 0 {
		return nil
	}

	f, c := bbmanager.basicBlocks.FloorCeil(addr)

	if !c.End() {
		bb = c.Value()

		if addr != bb.Address {
			if !f.End() {
				bb = f.Value()
			} else {
				bb = nil
			}
		}
	} else if !f.End() {
		bb = f.Value()
	}

	if bb != nil && addr > bb.LastAddress {
		bb = nil
	}

	if bb != nil && bb.Address > addr {
		bbmanager.basicBlocks.String()
		err := fmt.Errorf("found=%d query=%d", bb.Address, addr)
		panic(err)
	}

	return
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
	bbref := bbmanager.GetRef(from_addr, to_addr)
	if bbref == nil {
		key := BBRefKey{
			From: from_addr,
			To:   to_addr,
		}

		bbref = &SoraBBRef{
			BBRefKey: key,
		}
		bbmanager.refsToBB[to_addr] = append(bbmanager.refsToBB[to_addr], from_addr)
		bbmanager.refsFromBB[from_addr] = append(bbmanager.refsFromBB[from_addr], to_addr)
		bbmanager.refs[key] = bbref
	}
	return bbref
}

func (bbmanager *BasicBlockManager) GetRef(from_addr, to_addr uint32) *SoraBBRef {
	key := BBRefKey{
		From: from_addr,
		To:   to_addr,
	}

	if bbref, ok := bbmanager.refs[key]; ok {
		return bbref
	}

	return nil
}

func (bbmanager *BasicBlockManager) GetRefs(addr uint32) (xref_froms []*SoraBBRef, xref_tos []*SoraBBRef) {
	xref_froms = make([]*SoraBBRef, 0)
	xref_tos = make([]*SoraBBRef, 0)

	if from_bbs_, ok_in := bbmanager.refsToBB[addr]; ok_in {
		for _, from_bb := range from_bbs_ {
			if bbref := bbmanager.GetRef(from_bb, addr); bbref != nil {
				xref_froms = append(xref_froms, bbref)
			}
		}
	}

	if to_bbs, ok_out := bbmanager.refsFromBB[addr]; ok_out {
		for _, to_bb := range to_bbs {
			if bbref := bbmanager.GetRef(addr, to_bb); bbref != nil {
				xref_tos = append(xref_tos, bbref)
			}
		}
	}

	return
}

func (bbmanager *BasicBlockManager) SplitAt(split_addr uint32) (prev_bb, split_bb *SoraBasicBlock) {
	prev_bb = bbmanager.Get(split_addr)
	if prev_bb == nil {
		return
	} else if prev_bb.Address == split_addr {
		split_bb = prev_bb
		return
	} else if prev_bb.Address > split_addr {
		fmt.Printf("ERROR:\tunable to split non exist bb 0x%08x\n", split_addr)
		return nil, nil
	}

	last_addr := prev_bb.LastAddress
	if prev_bb.LastAddress >= split_addr {
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
