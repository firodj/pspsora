package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/davecgh/go-spew/spew"
)

const (
	KIND_ID    uint16 = 0x4449 // 'I'49 _ 'D'44
	KIND_SZ    uint16 = 0x5A53 // 'S'53 _ 'Z'5A
	KIND_START uint16 = 0x5453 // 'S'53 _ 'T'54
	KIND_NAME  uint16 = 0x4D4E // 'N'4E _ 'M'4D
	KIND_END   uint16 = 0x4445 // 'E'45 _ 'D'44
)

type RefTs int

type BBTraceStackItem struct {
	address uint32
	RA      uint32
	Fun     *SoraFunction
	NodeID  FunGraphNodeID
}

func NewStackItem(bb_init *SoraBasicBlock) *BBTraceStackItem {
	s := &BBTraceStackItem{
		address: bb_init.Address,
	}

	return s
}

func (s *BBTraceStackItem) SetAddress(bb *SoraBasicBlock) {
	s.address = bb.Address
	if s.Fun != nil {
		s.Fun.AddBB(s.address)
	}
}

func (s *BBTraceStackItem) Address() uint32 { return s.address }

type BBTraceThreadState struct {
	ID          uint16
	PC          uint32
	RegSP       int
	Stack       *Queue[*BBTraceStackItem]
	Executing   bool
	FunGraph    *FunGraph
	CallHistory *CallHistory

	Name string
}

type BBTraceParam struct {
	ID     uint16
	Kind   uint32
	PC     uint32
	LastPC uint32
	Nts    RefTs
}

type BBTraceYield func(param BBTraceParam)

type BBTraceParser struct {
	doc       *SoraDocument
	filename  string
	Nts       RefTs
	Fts       RefTs
	CurrentID uint16
	Threads   map[uint16]*BBTraceThreadState
}

func NewBBTraceParser(doc *SoraDocument, filename string) *BBTraceParser {
	bbtrace := &BBTraceParser{
		doc:       doc,
		filename:  filename,
		CurrentID: 0,
		Nts:       0,
		Fts:       0,
	}
	return bbtrace
}

func FindFirstNull(b []byte) int {
	l := 0
	x := l

	if b == nil {
		return -1
	}
	r := len(b)
	if r == 0 {
		return -1
	}

	for {
		if r-1 <= l {
			x = l
			if b[x] != 0 {
				x++
			}
			break
		}

		x = l + ((r - l) / 2)

		if b[x] == 0 {
			r = x
		} else {
			l = x + 1
		}
	}

	return x
}

func (bbtrace *BBTraceParser) EndParsing() {
	for _, thread := range bbtrace.Threads {
		if thread.CallHistory != nil {
			thread.CallHistory.StopAll(thread.Stack.Len(), bbtrace.Nts)
		}
	}
}

func (bbtrace *BBTraceParser) Parse(ctx context.Context, length int) error {
	bin, err := os.Open(bbtrace.filename)
	if err != nil {
		return err
	}
	defer bin.Close()

	bbtrace.Threads = make(map[uint16]*BBTraceThreadState)

	bbtrace.Nts = 1
	bbtrace.Fts = 1
	initial_length := length
	cur_ID := uint16(0)

	buf32 := make([]byte, 4)
	buf16 := make([]byte, 2)

	process := func(ctx context.Context, ch chan error) (stop bool, err error) {
		_, err = bin.Read(buf16)
		if err != nil {
			if err != io.EOF {
				return
			}
			fmt.Println("INFO:\tstop by EOF")
			stop = true
			return
		}

		kind := uint16(binary.LittleEndian.Uint16(buf16))
		if kind != KIND_ID {
			err = fmt.Errorf("ERROR:\tunmatched kind 'ID', found: 0x%x", kind)
			return
		}

		_, err = bin.Read(buf16)
		if err != nil {
			return
		}
		cur_ID = uint16(binary.LittleEndian.Uint16(buf16))

		_, err = bin.Read(buf16)
		if err != nil {
			return
		}
		kind = uint16(binary.LittleEndian.Uint16(buf16))

		if kind != KIND_SZ {
			err = fmt.Errorf("ERROR:\tunmatched kind 'SZ', found: 0x%x", kind)
			return
		}

		_, err = bin.Read(buf32)
		if err != nil {
			return
		}
		size := int(binary.LittleEndian.Uint32(buf32))

		records := make([]byte, size*4)
		_, err = bin.Read(records)
		if err != nil {
			return
		}

		fmt.Printf("INFO:\t[%d] read record size=%d\n", cur_ID, size)
		currentThread := bbtrace.SetCurrentThread(cur_ID)

		if currentThread.CallHistory != nil {
			currentThread.CallHistory.AddMarker(bbtrace.Nts, currentThread.Name)
			currentThread.CallHistory.Fts = bbtrace.Fts
		}

		err = nil
		for i := 0; i < size; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("canceling")
				err = ctx.Err()
			default:
			}

			if stop || err != nil {
				break
			}

			last_kind := kind

			last_pc := uint32(0)
			pc := binary.LittleEndian.Uint32(records[i*4:])

			if (pc & 0xFFFF0000) == 0 {
				kind = uint16(pc & 0xFFFF)

				switch kind {
				case KIND_START:
					i++
					pc = binary.LittleEndian.Uint32(records[i*4:])

					past_pc := bbtrace.SetCurrentThreadPC(pc)
					fmt.Printf("INFO:\t[%d] #(%d/%d) KIND_START pc=0x%08x last_pc=0x%08x\n", cur_ID, i-1, size, pc, past_pc)
				case KIND_NAME:
					i++
					j := i + 8
					str := records[i*4 : j*4]
					i = j - 1
					name := string(str[0:FindFirstNull(str)])
					fmt.Printf("INFO:\t[%d] #(%d/%d) KIND_NAME name=%s\n", cur_ID, i-8, size, name)
					if last_kind == KIND_START {
						currentThread.Name = name

						if currentThread.CallHistory != nil {
							currentThread.CallHistory.AddMarker(bbtrace.Nts, currentThread.Name)
						}

						switch name {
						case "idle0", "idle1", "SceIoAsync":
							currentThread.Executing = false
						}
					} else {
						err = fmt.Errorf("unknown name for what last_kind: 0x%04x", last_kind)
					}
				case KIND_END:
					i++
					end_pc := binary.LittleEndian.Uint32(records[i*4:])
					fmt.Printf("INFO:\t[%d] #(%d/%d) KIND_END end_pc=0x%08x\n", cur_ID, i-1, size, end_pc)
				default:
					err = fmt.Errorf("[%d] unknown kind: 0x%04x", cur_ID, kind)
				}

				continue
			}

			i++
			last_pc = binary.LittleEndian.Uint32(records[i*4:])

			if currentThread.Executing {
				param := BBTraceParam{
					ID:     bbtrace.CurrentID,
					Kind:   0,
					PC:     pc,
					LastPC: last_pc,
					Nts:    bbtrace.Nts,
				}

				//fmt.Printf("DEBUG:\t[%d] #(%d/%d) %d {0x%08x, 0x%08x}\n", cur_ID, i-1, size, param.Nts, param.PC, param.LastPC)

				err = bbtrace.ParsingBB(param)
				if err != nil {
					break
				}
			} else {
				fmt.Printf("DEBUG:\t[%d] #(%d/%d) skip thread %s (0x%08x, 0x%08x)\n", cur_ID, i-1, size, currentThread.Name,
					pc, last_pc)
			}

			bbtrace.Nts++

			if length > 0 {
				length -= 1
				if length == 0 {
					fmt.Printf("INFO:\tstop by length (%d)\n", initial_length)
					stop = true
					break
				}
			}
		}

		if currentThread.CallHistory != nil {
			currentThread.CallHistory.AddMarker(bbtrace.Nts, currentThread.Name)
			bbtrace.Fts = currentThread.CallHistory.Fts
		}

		return
	}

	ch := make(chan error)

	producer := func() {
		for {
			stop, err := process(ctx, ch)

			ch <- err

			if stop || err != nil {
				break
			}
		}
		close(ch)
	}

	consumer := func() (err error) {
		for err = range ch {
			if err != nil {
				return
			}
		}
		return
	}

	go producer()
	defer bbtrace.EndParsing()
	return consumer()
}

func (bbtrace *BBTraceParser) SetCurrentThread(id uint16) *BBTraceThreadState {
	if bbtrace.CurrentID == 0 || bbtrace.CurrentID != id {
		bbtrace.CurrentID = id
		if _, ok := bbtrace.Threads[bbtrace.CurrentID]; !ok {
			bbtrace.Threads[bbtrace.CurrentID] = &BBTraceThreadState{
				ID:          bbtrace.CurrentID,
				RegSP:       0,
				PC:          0,
				Executing:   true,
				FunGraph:    NewFunGraph(),
				CallHistory: nil, //NewCallHistory(),
				Stack:       new(Queue[*BBTraceStackItem]),
			}
		}
	}
	return bbtrace.Threads[bbtrace.CurrentID]
}

func (bbtrace *BBTraceParser) SetCurrentThreadPC(pc uint32) uint32 {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]

	last_pc := currentThread.PC
	currentThread.PC = pc

	return last_pc
}

func (bbtrace *BBTraceParser) ParsingBB(param BBTraceParam) error {
	if param.ID != bbtrace.CurrentID {
		panic("assert failed")
	}
	bbtrace.SetCurrentThreadPC(param.PC)

	theBB, err := bbtrace.EnsureBB(param.PC)
	if err != nil {
		return err
	}

	if param.LastPC == 0 {
		// Usually start thread doesn't have last_pc
		bbtrace.OnEnterFunc(theBB, 0)
		return nil
	}

	err = bbtrace.OnMergingPastToLast(param.LastPC)
	if err != nil {
		return err
	}

	lastBB := bbtrace.doc.BBManager.Get(param.LastPC)
	if lastBB == nil {
		return fmt.Errorf("unable to get last BB 0x%08x at 0x%08x", param.LastPC, theBB.Address)
	}

	brInstr := bbtrace.doc.InstrManager.Get(lastBB.BranchAddress)
	if brInstr == nil {
		return fmt.Errorf("unable to get lat Instruction at 0x%08x", lastBB.BranchAddress)
	}

	bbtrace.doc.BBManager.CreateReference(lastBB.Address, theBB.Address)

	if brInstr.Mnemonic == "jal" || brInstr.Mnemonic == "jalr" {
		ra := brInstr.Address + 4
		if brInstr.Info.HasDelaySlot {
			ra += 4
		}
		bbtrace.OnEnterFunc(theBB, ra)
	} else if brInstr.Mnemonic == "jr" && brInstr.Args[0].Reg == "ra" {
		bbtrace.OnLeaveFunc(theBB)
	} else {
		bbtrace.OnContinueNext(theBB)
	}

	return nil
}

func (bbtrace *BBTraceParser) EnsureBB(bb_addr uint32) (*SoraBasicBlock, error) {
	theBB := bbtrace.doc.BBManager.Get(bb_addr)

	if theBB == nil {
		bbtrace.doc.ProcessBB(bb_addr, 0, bbtrace.OnEachBB)
		theBB = bbtrace.doc.BBManager.Get(bb_addr)

		if theBB == nil {
			err := fmt.Errorf("unable to get BB after creating at: 0x%08x", bb_addr)
			return nil, err
		}
	} else if theBB.Address != bb_addr {
		prevBB, splitBB := bbtrace.doc.BBManager.SplitAt(bb_addr)
		if prevBB != theBB {
			err := fmt.Errorf("unexpected prevBB(0x%08x) != theBB(0x%08x)", prevBB.Address, theBB.Address)
			return nil, err
		}
		theBB = splitBB
		fmt.Printf("INFO:\tsplit bb at 0x%08x from original 0x%08x\n", splitBB.Address, prevBB.Address)
	}

	return theBB, nil
}

func (bbtrace *BBTraceParser) OnEachBB(state BBAnalState) {
	newBB := bbtrace.doc.BBManager.Get(state.BBAddr)

	if newBB == nil {
		newBB = bbtrace.doc.BBManager.Create(state.BBAddr)
		if newBB == nil {
			fmt.Printf("ERROR:\tunable to create BB at: 0x%08x, either already exists?\n", state.BBAddr)
			return
		} else {
			newBB.LastAddress = state.LastAddr
			newBB.BranchAddress = state.BranchAddr
		}
	} else if newBB.Address != state.BBAddr {
		fmt.Printf("ERROR:\tfix me to split bb during OnEachBB at: 0x%08x\n", state.BBAddr)
		return
	}
}

func (bbtrace *BBTraceParser) OnEnterFunc(theBB *SoraBasicBlock, ra uint32) {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]
	parent_ID := FunGraphNodeID(0)

	if currentThread.Stack.Len() > 0 {
		if ra == 0 {
			fmt.Printf("WARNING:\tundefined ra when entering func\n")
		}
		currentThread.Stack.Top().RA = ra
		parent_ID = currentThread.Stack.Top().NodeID

		if currentThread.CallHistory != nil {
			level := currentThread.Stack.Len()
			currentThread.CallHistory.EndBlock(level, bbtrace.Nts)
		}
	}

	theFunc := bbtrace.doc.FunManager.Get(theBB.Address)
	if theFunc == nil {
		fn_start := bbtrace.doc.SymMap.GetFunctionStart(theBB.Address)
		if fn_start != 0 {
			theFunc, _ = bbtrace.doc.FunManager.SplitAt(theBB.Address)

			if theFunc == nil {
				fmt.Printf("ERROR:\tsplit func 0x%08x\n", theBB.Address)
			}
		} else {
			theFunc = bbtrace.doc.FunManager.CreateNewFunction(theBB.Address, theBB.Size())

			if theFunc == nil {
				fmt.Printf("ERROR:\tunable to create func from bb 0x%08x\n", theBB.Address)
			}
		}
	}

	theFunc.AddBB(theBB.Address)

	stack_item := NewStackItem(theBB)
	stack_item.Fun = theFunc

	if currentThread.FunGraph != nil {
		node := currentThread.FunGraph.AddNode(theBB.Address, parent_ID)
		node.Fun = theFunc
		node.Duration++
		stack_item.NodeID = node.ID
	}

	currentThread.Stack.Push(stack_item)

	if currentThread.CallHistory != nil {
		level := currentThread.Stack.Len()
		currentThread.CallHistory.AddBlock(level, bbtrace.Nts, theBB.Address, theFunc.Name)
	}

	//fmt.Printf("INFO:\tenter func bb 0x%08x ra=0x%08x", theBB.Address, ra)
	//if theFunc != nil {
	//	fmt.Printf(" name=%s", theFunc.Name)
	//}
	//fmt.Println()
	bbtrace.doc.DebugBB(theBB, "enter")
}

func (bbtrace *BBTraceParser) OnLeaveFunc(theBB *SoraBasicBlock) {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]

	if currentThread.CallHistory != nil {
		level := currentThread.Stack.Len()
		currentThread.CallHistory.EndBlock(level, bbtrace.Nts)
	}
	_ = currentThread.Stack.Pop()

	if currentThread.Stack.Len() > 0 {
		expected_ra := currentThread.Stack.Top().RA

		if expected_ra != theBB.Address {
			fmt.Printf("WARNING:\tunexpected ra 0x%08x, expecting 0x%08x\n--- callback? ---\n", theBB.Address, expected_ra)

			bbtrace.OnEnterFunc(theBB, expected_ra)
		} else {
			past_bb := currentThread.Stack.Top().Address()
			currentThread.Stack.Top().SetAddress(theBB)
			//fmt.Printf("INFO:\tleave bb 0x%08x\n", past_top.Address())

			bbtrace.doc.BBManager.CreateReference(past_bb, theBB.Address).SetAdjacent(true)

			if currentThread.CallHistory != nil {
				level := currentThread.Stack.Len()
				currentThread.CallHistory.EndBlock(level, bbtrace.Nts)
			}

			bbtrace.doc.DebugBB(theBB, "leave")
		}
	} else {
		myFunc := bbtrace.doc.FunManager.Get(theBB.Address)
		fmt.Printf("INFO:\tend of stack, goto: 0x%08x", theBB.Address)
		if myFunc != nil {
			fmt.Printf(" name: %s", myFunc.Name)
		}
		fmt.Println()

		bbtrace.doc.DebugBB(theBB, "end")
	}
}

func (bbtrace *BBTraceParser) OnContinueNext(theBB *SoraBasicBlock) {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]
	currentThread.Stack.Top().SetAddress(theBB)

	bbtrace.doc.DebugBB(theBB, "continue")
}

func (bbtrace *BBTraceParser) OnMergingPastToLast(last_pc uint32) error {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]

	past_addr := currentThread.Stack.Top().Address()

	bb_visits := make(map[uint32]*BBVisit)

	for n := 0; true; n++ {
		pastBB := bbtrace.doc.BBManager.Get(past_addr)
		//fmt.Printf("INFO:\tmerging past #%d 0x%08x\n", n, past_addr)

		if pastBB == nil {
			return fmt.Errorf("OnMergingPastToLast past BB notexist: 0x%08x towards: 0x%08x", past_addr, last_pc)
		}

		if _, ok := bb_visits[pastBB.Address]; !ok {
			bb_visits[pastBB.Address] = &BBVisit{
				BB: pastBB, Visited: true,
			}
		} else {
			return fmt.Errorf("merging loop at #%d 0x%08x", n, pastBB.Address)
		}

		if n > 0 {
			currentThread.Stack.Top().SetAddress(pastBB)
			bbtrace.doc.DebugBB(pastBB, "merging")
		} else {
			if currentThread.Stack.Top().Address() != pastBB.Address {
				return fmt.Errorf("assert failed, expect stack.Top.Address == pastBB.Address")
			}
		}

		next_addr := pastBB.LastAddress + 4

		if pastBB.BranchAddress != 0 {
			pastBrInstr := bbtrace.doc.InstrManager.Get(pastBB.BranchAddress)

			if pastBrInstr == nil {
				return fmt.Errorf("no branch instr for past BB at 0x%08x", pastBB.BranchAddress)
			}

			if pastBrInstr.Info.IsLikelyBranch {
				if pastBB.BranchAddress == last_pc {
					break
				}
			}

			if pastBB.LastAddress == last_pc {
				break
			}

			if pastBrInstr.Info.IsConditional {
				next_addr = pastBB.LastAddress + 4
				if pastBrInstr.Info.IsBranchToRegister {
					fmt.Printf("WARNING:\tunimplemented conditional register branch for merging\n")
				}
			} else {
				if pastBrInstr.Info.IsBranchToRegister {
					break
				} else if pastBrInstr.Info.BranchTarget != 0 {
					next_addr = pastBrInstr.Info.BranchTarget
				} else {
					spew.Dump(pastBrInstr)
					panic("todo")
				}
			}
		}
		bbtrace.doc.BBManager.CreateReference(pastBB.Address, next_addr).SetAdjacent(true)
		past_addr = next_addr
	}

	return nil
}

func (bbtrace *BBTraceParser) DumpAllFunGraph() {
	for _, thread := range bbtrace.Threads {
		fmt.Printf("Thread #%d %s\n", thread.ID, thread.Name)

		if thread.FunGraph != nil {
			thread.FunGraph.DumpNode(0, 0)
		}
	}
}

func (bbtrace *BBTraceParser) DumpAllCallHistory() {
	for _, thread := range bbtrace.Threads {
		fmt.Printf("Call History Thread #%d %s\n", thread.ID, thread.Name)

		if thread.CallHistory != nil {
			thread.CallHistory.Dump()
		}
	}
}
