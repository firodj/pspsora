package internal

import (
	"fmt"

	"github.com/firodj/ppsspp/disasm/pspdisasm/models"
)

// generics Queue
type Queue[T any] struct {
	elements []T
}

func (q *Queue[T]) Push(element T) {
	q.elements = append(q.elements, element)
}

func (q *Queue[T]) Len() int {
	return len(q.elements)
}

func (q *Queue[T]) Pop() T {
	element := q.elements[0]
	q.elements = q.elements[1:]
	return element
}

type FunctionAnalyzer struct {
	doc *SoraDocument
	idx int
}

func NewFunctionAnalyzer(doc *SoraDocument, idx int) *FunctionAnalyzer {
	return &FunctionAnalyzer{
		doc: doc,
		idx: idx,
	}
}

type BBVisit struct {
	BB      *SoraBasicBlock
	Visited bool
}

func (anal *FunctionAnalyzer) Process() {
	fun := anal.doc.GetFunctionByIndex(anal.idx)
	if fun == nil {
		fmt.Printf("ERROR:\tno func for index:%d\n", anal.idx)
		return
	}

	bb_visits := make(map[uint32]*BBVisit)
	for _, bb_addr := range fun.BBAddresses {
		bb_visits[bb_addr] = &BBVisit{
			BB: anal.doc.BBGet(bb_addr),
			Visited: false,
		}
	}

	var bb_queues Queue[uint32]
	bb_queues.Push(fun.Address)

	for bb_queues.Len() > 0 {
		cur_addr := bb_queues.Pop()

		if _, ok := bb_visits[cur_addr]; !ok {
			if _, bbfun := anal.doc.GetFunctionByAddress(cur_addr); bbfun == nil {
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
	anal.ProcessBB(fun.Address, fun.LastAddress(), func (bbas BBAnalState) {
		fmt.Printf("bb:0x%08x br:0x%08x last_addr:0x%08x\n", bbas.BBAddr, bbas.BranchAddr, bbas.LastAddr)
		for _, line := range bbas.Lines {
			fmt.Printf("\t0x%08x %s\n", line.Address, line.Dizz)
		}
	})
}


func (anal *FunctionAnalyzer) ProcessBB(start_addr uint32, last_addr uint32, cb BBYieldFunc) int {
	var bbas BBAnalState
	bbas.Init()
	var prevInstr *models.MipsOpcode = nil

	for addr := start_addr; last_addr == 0 || addr <= last_addr; addr += 4 {
		bbas.SetBB(addr)

		instr := anal.doc.Disasm(addr)

		bbas.Append(instr)

		if instr.IsBranch {
			bbas.SetBranch(addr)

			if !instr.HasDelaySlot {
				fmt.Printf("WARNING:\tunhandled branch without delay shot\n")
				bbas.Yield(addr, cb)

				if last_addr == 0 && instr.IsConditional {
					break
				}
			}
		}

		if prevInstr != nil && prevInstr.HasDelaySlot {
			bbas.Yield(addr, cb)

			if last_addr == 0 && !prevInstr.IsConditional {
				break
			}
		}

		prevInstr = instr
	}

	bbas.Yield(last_addr, cb)

	return bbas.Count
}