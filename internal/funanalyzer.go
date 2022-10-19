package internal

import (
	"fmt"
)

type FunctionAnalyzer struct {
	doc *SoraDocument
	fun *SoraFunction
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

func (anal *FunctionAnalyzer) Process() {
	bb_visits := make(map[uint32]*BBVisit)
	for _, bb_addr := range anal.fun.BBAddresses {
		bb_visits[bb_addr] = &BBVisit{
			BB:      anal.doc.BBManager.Get(bb_addr),
			Visited: false,
		}
	}

	var bb_queues Queue[uint32]
	bb_queues.Push(anal.fun.Address)

	for bb_queues.Len() > 0 {
		cur_addr := bb_queues.Pop()

		if _, ok := bb_visits[cur_addr]; !ok {
			if bbfun := anal.doc.FunManager.Get(cur_addr); bbfun == nil {
				fmt.Printf("WARNING:\tunknown bb and not a func: 0x08%x\n", cur_addr)
			}
			continue
		}

		cur_visit := bb_visits[cur_addr]
		if cur_visit.Visited {
			continue
		}
		cur_visit.Visited = true

		fmt.Println("---")

	}

	// TODO:
	/*
		anal.ProcessBB(fun.Address, fun.LastAddress(), func (bbas BBAnalState) {
			fmt.Printf("bb:0x%08x br:0x%08x last_addr:0x%08x\n", bbas.BBAddr, bbas.BranchAddr, bbas.LastAddr)
			for _, line := range bbas.Lines {
				fmt.Printf("\t0x%08x %s\n", line.Address, line.Dizz)
			}
		})
	*/
}
