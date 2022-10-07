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
	Addr uint32
	RA   uint32
	Fun *SoraFunction
	// FNTreeNodeID
}

type BBTraceThreadState struct {
	ID uint16
	PC uint32
	RegSP int
	Stack []BBTraceStackItem
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

			// last_pc := uint32(0)
			pc := binary.LittleEndian.Uint32(records[i * 4:])

			if ((pc & 0xFFFF0000) == 0) {
				kind = uint16(pc & 0xFFFF)

				if (kind == KIND_START) {
					i++
					pc = binary.LittleEndian.Uint32(records[i * 4:])

					fmt.Printf("INFO:\tstart 0x%08x\n", pc)
					bbtrace.SetCurrentThreadPC(cur_ID, pc)
				} else if (kind == KIND_NAME) {
					i++
					j := i+32
					str := records[i * 4:j]
					i = j
					name := string(str[0:FindFirstNull(str)])
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

			if bbtrace.Threads[bbtrace.CurrentID].Executing {
				param := BBTraceParam{
					ID: bbtrace.CurrentID,
					Kind: 0,
					PC: pc,
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
	last_pc := bbtrace.SetCurrentThreadPC(param.ID, param.PC)

	fmt.Printf("#%d {0x%08x, 0x%08x}\n", param.Nts, param.PC, param, last_pc)

	if last_pc == 0 {
		// Usually start thread doesn't have last_pc
		fmt.Println("TODO: OnEnterFunc bb, 0")
		return nil
	}

	fmt.Println("TODO: OnMergingPastToLast param.last_pc)")


	return nil
}

func (bbtrace *BBTraceParser) EnsureBB(bb_addr uint32) (*SoraBasicBlock, error) {
	bb := bbtrace.doc.GetBB(bb_addr)
}