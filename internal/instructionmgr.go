package internal

import (
	"github.com/firodj/pspsora/binarysearchtree"
	"github.com/firodj/pspsora/models"
)

type SoraInstruction struct {
	Info     models.MipsOpcode
	Address  uint32
	Mnemonic string
	Args     []*SoraArgument
}

type InstructionManager struct {
	doc *SoraDocument

	instructions binarysearchtree.AVLTree[uint32, *SoraInstruction]
}

func NewInstructionManager(doc *SoraDocument) *InstructionManager {
	return &InstructionManager{
		doc: doc,
	}
}

func (mgr *InstructionManager) Create(addr uint32, info *models.MipsOpcode) *SoraInstruction {
	if mgr.Get(addr) != nil {
		return nil
	}
	instr := &SoraInstruction{
		Info:    *info,
		Address: addr,
	}
	mgr.instructions.Insert(addr, instr)
	return instr
}

func (mgr *InstructionManager) Get(addr uint32) *SoraInstruction {
	it := mgr.instructions.Search(addr)
	if it.End() {
		return nil
	}
	return it.Value()
}

func (instr *SoraInstruction) GetSyscallInfo() (int, int) {
	if instr.Mnemonic == "syscall" {
		// Syscalls look like this: 00-- ---- ---- xxxx vvvv vv00 1100
		callno := (instr.Info.Encoded >> 6) & 0xFFFFF
		funcnum := callno & 0xFFF
		modulenum := (callno & 0xFF000) >> 12
		return int(modulenum), int(funcnum)
	}
	return -1, -1
}
