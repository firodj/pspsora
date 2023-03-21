package internal

import (
	"fmt"
	"sort"
)

type FunctionAnalyzer struct {
	doc       *SoraDocument
	fun       *SoraFunction
	bb_visits map[uint32]*BBVisit
	bb_rets   []uint32
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
	var addresses []uint32
	for addr := range anal.bb_visits {
		addresses = append(addresses, addr)
	}
	sort.SliceStable(addresses, func(i, j int) bool {
		return addresses[i] < addresses[j]
	})

	bb_adj := uint32(0)

	for bb_i := range addresses {
		bb_addr := addresses[bb_i]
		anal.doc.BBManager.Get(bb_addr)

		xref_tos := anal.doc.BBManager.GetExitRefs(bb_addr)

		if bb_adj != 0 {
			if bb_adj != bb_addr {
				fmt.Printf("(auto) goto %s\t", anal.doc.GetLabelName(bb_adj))
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
		}
	}

	if len(anal.bb_rets) == 0 {
		fmt.Println("WARNING\tfunction doesnt visit return")
	} else {
		fmt.Print("INFO\tfunction has retruns: ")
		for _, bb_ret := range anal.bb_rets {
			fmt.Printf("0x%08x ", bb_ret)
		}
		fmt.Println()
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
	bb_queues.PushUnique(anal.fun.Address)

	anal.bb_rets = make([]uint32, 0)

	for bb_queues.Len() > 0 {
		cur_addr := bb_queues.Pop()

		if _, ok := anal.bb_visits[cur_addr]; !ok {
			/**
			if cur_addr >= anal.fun.Address && cur_addr <= anal.fun.LastAddress() {
				if bb_notvisit := anal.doc.BBManager.Get(cur_addr); bb_notvisit != nil {
					if bb_notvisit.Address == cur_addr {
						anal.bb_visits[cur_addr] = &BBVisit{
							BB:      bb_notvisit,
							Visited: false,
						}
					} else {
						fmt.Printf("DEBUG:\tunexpected bb 0x%08x for %08x\n", bb_notvisit.Address, cur_addr)
						continue
					}
				}
			} else {
			*/
			fmt.Printf("DEBUG:\tfun %s/0x%08x doesnt include bb 0x08%x\n", anal.fun.Name, anal.fun.Address, cur_addr)
			continue
		}

		cur_visit := anal.bb_visits[cur_addr]
		cur_visit.Visited = true

		brInstr := anal.doc.InstrManager.Get(cur_visit.BB.BranchAddress)
		if brInstr != nil {
			if brInstr.Mnemonic == "jr" && brInstr.Args[0].Reg == "ra" {
				anal.bb_rets = append(anal.bb_rets, cur_addr)
			}
		}

		//is_then, is_else := false, false

		outfrom_bb := anal.doc.BBManager.GetExitRefs(cur_addr)
		for _, xref_to_bb := range outfrom_bb {
			bb_queues.PushUnique(xref_to_bb.To)
			/**
			if xref_to_bb.IsThen {
				is_then = true
			}
			if xref_to_bb.IsElse {
				is_else = true
			}
			*/
		}

		/**
		if is_then && !is_else {
			bb_queues.PushUnique(cur_visit.BB.LastAddress + 4)
		} else if is_else && !is_then {
			if brInstr != nil && brInstr.Info.BranchTarget != 0 {
				bb_queues.PushUnique(brInstr.Info.BranchTarget)
			}
		}
		*/
	}
}
