package internal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	KIND_ID uint16 = 0x4449 // 49'I' 44'D'
	KIND_SZ uint16 = 0x5A53 // 53'S' 5A'Z'
  KIND_START uint16 = 0x5453 // 53'S' 54'T'
	KIND_NAME uint16 = 0x4D4E // 4E'N', 4D'M'
)

type RefTs int

type BBTraceStackItem struct {
	Address uint32
	RA   uint32
	Fun *SoraFunction
	// FNTreeNodeID
}

type BBTraceThreadState struct {
	ID uint16
	PC uint32
	RegSP int
	Stack Queue[*BBTraceStackItem]
	Executing bool
	// FlameGraph
	// FNHierarchy

	Name string
}

type BBTraceParam struct {
	ID uint16
	Kind uint32
	PC uint32
	LastPC uint32
	Nts RefTs
}

type BBTraceYield func (param BBTraceParam)

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
		doc: doc,
		filename: filename,
		CurrentID: 0,
		Nts: 0,
		Fts: 0,
	}
	return bbtrace
}

func FindFirstNull(b []byte) int  {
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
			l = x+1
		}
	}

	return x
}

func (bbtrace *BBTraceParser) Parse(cb BBTraceYield, length int) error {
	bin, err := os.Open(bbtrace.filename)
	if err != nil {
		return err
	}
	defer bin.Close()

	bbtrace.Threads = make(map[uint16]*BBTraceThreadState)

	ok := true
	bbtrace.Nts = 1
	bbtrace.Fts = 1
	if length < 1 {
		length = 1
	}
	var cur_ID uint16 = 0

	buf32 := make([]byte, 4)
	buf16 := make([]byte, 2)

	for ok {

		_, err := bin.Read(buf16)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		kind := uint16(binary.LittleEndian.Uint16(buf16))
		if kind != KIND_ID {
			return fmt.Errorf("ERROR:\tunmatched kind 'ID', found: 0x%x", kind)
		}

		_, err = bin.Read(buf16)
		if err != nil {
			return err
		}
		cur_ID = uint16(binary.LittleEndian.Uint16(buf16))
		//fmt.Println(cur_ID)

		_, err = bin.Read(buf16)
		if err != nil {
			return err
		}
		kind = uint16(binary.LittleEndian.Uint16(buf16))

		if kind != KIND_SZ {
			return fmt.Errorf("ERROR:\tunmatched kind 'SZ', found: 0x%x", kind)
		}

		_, err = bin.Read(buf32)
		if err != nil {
			return err
		}
		size := int(binary.LittleEndian.Uint32(buf32))

		records := make([]byte, size * 4)
		_, err = bin.Read(records)
		if err != nil {
			return err
		}

		for i := 0; i<size; i++ {
			last_kind := kind

			last_pc := uint32(0)
			pc := binary.LittleEndian.Uint32(records[i * 4:])
			fmt.Printf("i = %d\n", i)

			if ((pc & 0xFFFF0000) == 0) {
				kind = uint16(pc & 0xFFFF)

				if (kind == KIND_START) {
					i++
					pc = binary.LittleEndian.Uint32(records[i * 4:])

					fmt.Printf("INFO:\tKIND_START pc=0x%08x\n", pc)
					bbtrace.SetCurrentThreadPC(cur_ID, pc)
				} else if (kind == KIND_NAME) {
					i++
					j := i+8
					str := records[i * 4:j * 4]
					i = j-1
					name := string(str[0:FindFirstNull(str)])
					fmt.Printf("INFO:\tKIND_NAME name=%s\n", name)
					if last_kind == KIND_START {
						bbtrace.Threads[bbtrace.CurrentID].Name = name
					} else {
						fmt.Printf("ERROR:\tunknown name for what last_kind: 0x%04x\n", last_kind)
						break
					}

					switch name {
					case "idle0", "idle1", "SceIoAsync":
						bbtrace.Threads[bbtrace.CurrentID].Executing = false
					}
				} else {
					fmt.Printf("ERROR:\tunknown kind: 0x%04x\n", kind)
					break
				}

				continue
			}

			i++
			last_pc = binary.LittleEndian.Uint32(records[i * 4:])

			if bbtrace.Threads[bbtrace.CurrentID].Executing {
				param := BBTraceParam{
					ID: bbtrace.CurrentID,
					Kind: 0,
					PC: pc,
					LastPC: last_pc,
					Nts: bbtrace.Nts,
				}

				bbtrace.ParsingBB(param)
			}

			bbtrace.Nts++
			length -= 1
			if (length == 0) {
				ok = false
				break
			}

			// threads_[cur_ID].flame_graph.AddMarker(nts_, threads_[cur_ID].name());
			// fts_ = threads_[cur_ID].flame_graph.fts_;
		}

		return nil
	}

	return nil
}

func (bbtrace *BBTraceParser) SetCurrentThread(id uint16) {
	if bbtrace.CurrentID == 0 || bbtrace.CurrentID != id {
		bbtrace.CurrentID = id
		if _, ok := bbtrace.Threads[bbtrace.CurrentID]; !ok {
			bbtrace.Threads[bbtrace.CurrentID] = &BBTraceThreadState{
				ID: bbtrace.CurrentID,
			  RegSP: 0,
			  PC: 0,
			  Executing: true,
			}
		}
	}
}

func (bbtrace *BBTraceParser) SetCurrentThreadPC(id uint16, pc uint32) uint32 {
	bbtrace.SetCurrentThread(id)

	last_pc := bbtrace.Threads[bbtrace.CurrentID].PC
	bbtrace.Threads[bbtrace.CurrentID].PC = pc

	return last_pc
}

func (bbtrace *BBTraceParser) ParsingBB(param BBTraceParam) error {
	bbtrace.SetCurrentThreadPC(param.ID, param.PC)

	fmt.Printf("#%d {0x%08x, 0x%08x}\n", param.Nts, param.PC, param.LastPC)

	theBB, err := bbtrace.EnsureBB(param.PC)
	if err != nil {
		return err
	}
	//spew.Dump(bb)

	if param.LastPC == 0 {
		// Usually start thread doesn't have last_pc
		bbtrace.OnEnterFunc(theBB, 0)
		return nil
	}

	bbtrace.OnMergingPastToLast(param.LastPC)

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

	if theBB == nil  {
		bbtrace.doc.ProcessBB(bb_addr, 0, bbtrace.OnEachBB)
		theBB = bbtrace.doc.BBManager.Get(bb_addr)

		if theBB == nil {
			err := fmt.Errorf("ERROR:\tunable to get BB after creating at: 0x%08x", bb_addr)
			return nil, err
		}
	} else if (theBB.Address != bb_addr) {
		prevBB, splitBB := bbtrace.doc.BBManager.SplitAt(bb_addr)
		if prevBB != theBB {
			err := fmt.Errorf("ERROR:\tunexpected prevBB(0x%08x) != theBB(0x%08x)", prevBB.Address, theBB.Address)
			return nil, err
		}
		theBB = splitBB
		fmt.Printf("INFO:\tsplit bb at 0x%08x\n", splitBB.Address)
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
	fmt.Printf("OnEnterFunc(0x%08x, ra=0x%08x)\n", theBB.Address, ra)

	if currentThread.Stack.Len() > 0 {
		if ra == 0 {
			fmt.Printf("WARNING:\tundefined ra when entering func\n")
		}
		currentThread.Stack.Top().RA = ra
	}

	stack_item := &BBTraceStackItem{
		Address: theBB.Address,
	}
	currentThread.Stack.Push(stack_item)

	fmt.Printf("INFO:\tenter func bb 0x%08x\n", theBB.Address)
}

func (bbtrace *BBTraceParser) OnLeaveFunc(theBB *SoraBasicBlock) {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]

	currentThread.Stack.Pop()
	if currentThread.Stack.Len() > 0 {
		expected_ra := currentThread.Stack.Top().RA

		if expected_ra != theBB.Address {
			fmt.Printf("WARNING:\tunexpected ra 0x%08x, expecting 0x%08x\n--- callback? ---\n", theBB.Address, expected_ra)
			bbtrace.OnEnterFunc(theBB, expected_ra)
		} else {
			past_bb := currentThread.Stack.Top().Address
			currentThread.Stack.Top().Address = theBB.Address
			// FUNC

			bbtrace.doc.BBManager.CreateReference(past_bb, theBB.Address).SetAdjacent(true)

			// FLAMEGRAPH
		}
	} else {
		// FUNC
	}
}

func (bbtrace *BBTraceParser) OnContinueNext(theBB *SoraBasicBlock) {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]
	currentThread.Stack.Top().Address = theBB.Address
	// FUNC
}

func (bbtrace *BBTraceParser) OnMergingPastToLast(last_pc uint32) error {
	currentThread := bbtrace.Threads[bbtrace.CurrentID]

	past_addr := currentThread.Stack.Top().Address

	for n := 0; true; n++ {
		pastBB := bbtrace.doc.BBManager.Get(past_addr)

		if pastBB == nil {
			return fmt.Errorf("OnMergingPastToLast past BB notexist: 0x%08x towards: 0x%08x", past_addr, last_pc)
		}

		if n > 0 {
			// FUNC
			currentThread.Stack.Top().Address = pastBB.Address
		} else {
			if currentThread.Stack.Top().Address != pastBB.Address {
				return fmt.Errorf("assert failed stack.Top.Address != pastBB.Address")
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
			} else if pastBrInstr.Info.IsBranchToRegister {
				break
			}

			bbtrace.doc.BBManager.CreateReference(pastBB.Address, next_addr).SetAdjacent(true)
			past_addr = next_addr
		}
	}

	return nil
}