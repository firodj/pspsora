package internal

import "github.com/firodj/pspsora/binarysearchtree"

type BasicBlockManager struct {
	doc *SoraDocument

	basicBlocks binarysearchtree.AVLTree[uint32, *SoraBasicBlock]
	basicBlockRefs []SoraBBRef
}

func NewBasicBlockManager(doc *SoraDocument) *BasicBlockManager {
	return &BasicBlockManager{
		doc: doc,
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
