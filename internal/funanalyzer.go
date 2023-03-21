package internal

import (
	"fmt"
	"sort"
)

type FunctionAnalyzer struct {
	doc       *SoraDocument
	fun       *SoraFunction
	bb_visits map[uint32]*BBVisit
}

func NewFunctionAnalyzer(doc *SoraDocument, fun *SoraFunction) *FunctionAnalyzer {
	return &FunctionAnalyzer{
		doc: doc,
		fun: fun,
	}
}

type BBVisit struct {
	BB      *SoraBasicBlock
	Visited bool
}

func (anal *FunctionAnalyzer) Debug(cb BBYieldFunc) {
	addresses := anal.fun.BBAddresses
	sort.SliceStable(addresses, func(i, j int) bool {
		return i < j
	})

	bb_adj := uint32(0)

	for bb_i := range addresses {
		bb_addr := addresses[bb_i]
		anal.doc.BBManager.Get(bb_addr)

		xref_tos := anal.doc.BBManager.GetExitRefs(bb_addr)

		if bb_adj != 0 {
			if bb_adj != bb_addr {
				fmt.Printf("goto %s\t", anal.doc.GetLabelName(bb_adj))
			}
			bb_adj = 0
		}

		anal.doc.ProcessBB(bb_addr, 0, func(bbas BBAnalState) {
			if cur_visit, ok := anal.bb_visits[bbas.BBAddr]; ok {
				bbas.Visited = cur_visit.Visited
			}
			cb(bbas)
		})

		for _, xref_to := range xref_tos {
			if xref_to.IsAdjacent {
				bb_adj = xref_to.To
			}
			//fmt.Printf("  - %s\n", xref_to)
		}
	}
}

func (anal *FunctionAnalyzer) Process() {
	if anal.bb_visits != nil {
		fmt.Printf("WARNING:\function already processed\n")
		return
	}
	anal.bb_visits = make(map[uint32]*BBVisit)

	for _, bb_addr := range anal.fun.BBAddresses {
		anal.bb_visits[bb_addr] = &BBVisit{
			BB:      anal.doc.BBManager.Get(bb_addr),
			Visited: false,
		}
	}

	var bb_queues Queue[uint32]
	bb_queues.Push(anal.fun.Address)

	for bb_queues.Len() > 0 {
		cur_addr := bb_queues.Pop()

		if _, ok := anal.bb_visits[cur_addr]; !ok {
			fmt.Printf("DEBUG:\tfun %s/0x%08x doesnt have bb 0x08%x\n", anal.fun.Name, anal.fun.Address, cur_addr)
			continue
		}

		cur_visit := anal.bb_visits[cur_addr]
		if cur_visit.Visited {
			continue
		}
		cur_visit.Visited = true

		outfrom_bb := anal.doc.BBManager.GetExitRefs(cur_addr)
		for _, xref_to_bb := range outfrom_bb {
			bb_queues.Push(xref_to_bb.To)
		}
	}
}
