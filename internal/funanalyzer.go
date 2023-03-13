package internal

import (
	"fmt"
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
	for _, bb_addr := range anal.fun.BBAddresses {
		anal.doc.BBManager.Get(bb_addr)

		fmt.Printf("----\n  xrefs from:\n")
		xref_froms, xref_tos := anal.doc.BBManager.GetRefs(bb_addr)
		for _, xref_from := range xref_froms {
			fmt.Printf("  - %s\n", xref_from)
		}

		anal.doc.ProcessBB(bb_addr, 0, func(bbas BBAnalState) {
			if cur_visit, ok := anal.bb_visits[bbas.BBAddr]; ok {
				bbas.Visited = cur_visit.Visited
			}
			cb(bbas)
		})

		fmt.Printf("  xrefs to:\n")
		for _, xref_to := range xref_tos {
			fmt.Printf("  - %s\n", xref_to)
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
			if bbfun := anal.doc.FunManager.Get(cur_addr); bbfun == nil {
				fmt.Printf("WARNING:\tunknown bb (and not a func): 0x08%x\n", cur_addr)
			} else {
				fmt.Printf("WARNING:\tbb outisde: 0x08%x\n", cur_addr)
			}
			continue
		}

		cur_visit := anal.bb_visits[cur_addr]
		if cur_visit.Visited {
			continue
		}
		cur_visit.Visited = true

		_, outfrom_bb := anal.doc.BBManager.GetRefs(cur_addr)
		for _, xref_to_bb := range outfrom_bb {
			bb_queues.Push(xref_to_bb.To)
		}
	}
}
