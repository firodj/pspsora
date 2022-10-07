package internal

import (
	"fmt"

	"github.com/firodj/ppsspp/disasm/pspdisasm/models"
)

type BBYieldFunc func(state BBAnalState)

type BBAnalState struct {
	BBAddr uint32
	BranchAddr uint32
	LastAddr uint32
	Lines []models.MipsOpcode
  Count int
}

func (bbas *BBAnalState) Init() {
	bbas.Reset()
	bbas.Count = 0
}

func (bbas *BBAnalState) Reset() {
	bbas.BBAddr = 0
	bbas.BranchAddr = 0
	bbas.Lines = nil
}

func (bbas *BBAnalState) SetBB(addr uint32) {
	if bbas.BBAddr == 0 {
		bbas.BBAddr = addr
	}
}

func (bbas *BBAnalState) SetBranch(addr uint32) {
	if bbas.BranchAddr != 0 {
		fmt.Printf("WARNING:\tSetBranch already set\n")
	}
	bbas.BranchAddr = addr
}

func (bbas *BBAnalState) Append(instr *models.MipsOpcode) {
	bbas.Lines = append(bbas.Lines, *instr)
}

func (bbas *BBAnalState) Yield(last_addr uint32, cb BBYieldFunc) {
	if len(bbas.Lines) == 0 {
		return
	}
	bbas.LastAddr = last_addr
	cb(*bbas)

	bbas.Reset()
	bbas.Count += 1
}
