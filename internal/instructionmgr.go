package internal

import (
	"github.com/firodj/bst"
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

	instructions bst.AVLTree[uint32, *SoraInstruction]
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

func (instr *SoraInstruction) IsSyscall() bool {
	return instr.Mnemonic == "syscall"
}

func (instr *SoraInstruction) GetSyscallNumber() (int, int) {
	if !instr.IsSyscall() {
		return -1, -1
	}

	callno := instr.Args[0].ValOfs
	funcnum := callno & 0xFFF
	modulenum := (callno & 0xFF000) >> 12
	return int(modulenum), int(funcnum)
}
