package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"gopkg.in/yaml.v3"

	"github.com/firodj/pspsora/bridge"
)
type PSPSegment struct {
	Addr uint32 `yaml:"addr"`
	Size int    `yaml:"size"`
}

type PSPNativeModule struct {
	Name      string       `yaml:"name"`
	Segments  []PSPSegment `yaml:"segments"`
	EntryAddr uint32       `yaml:"entry_addr"`
}

type PSPModule struct {
	NM        PSPNativeModule `yaml:"nm"`
	TextStart uint32          `yaml:"textStart"`
	TextEnd   uint32          `yaml:"textEnd"`
	ModulePtr uint32          `yaml:"modulePtr"`
}

type SoraFunction struct {
	Name        string   `yaml:"name"`
	Address     uint32   `yaml:"address"`
	Size        uint32   `yaml:"size"`
	BBAddresses []uint32 `yaml:bb_addresses"`

}

func (fun *SoraFunction) LastAddress() uint32 {
	return fun.Address + fun.Size - 4
}

type PSPHLEFunction struct {
	Idx     string `yaml:"idx"`
	Nid     uint32 `yaml:"nid"`
	Name    string `yaml:"name"`
	ArgMask string `yaml:"argmask"`
	RetMask string `yaml:"retmask"`
	Flags   uint32 `yaml:"flags"`
}

type PSPHLEModule struct {
	Name  string           `yaml:"name"`
	Funcs []PSPHLEFunction `yaml:"funcs"`
}

type PSPLoadedModule struct {
	Name     string `yaml:"name"`
	Address  uint32 `yaml:"address"`
	Size     uint32 `yaml:"size"`
	IsActive bool   `yaml:"isActive"`
}

type PSPMemory struct {
	Start uint32 `yaml:"start"`
	Size  int    `yaml:"size"`
}

type SoraYaml struct {
	Module        PSPModule         `yaml:"module"`
	Memory        PSPMemory         `yaml:"memory"`
	LoadedModules []PSPLoadedModule `yaml:"loaded_modules"`
	SymFunctions  []SoraFunction    `yaml:"functions"`
	HLEModules    []PSPHLEModule    `yaml:"hle_modules"`
}

type SoraDocument struct {
	yaml     SoraYaml

	// HLEModules (yaml)
	Parser  *BBTraceParser
	BBManager      *BasicBlockManager
	FunManager     *FunctionManager
	InstrManager     *InstructionManager

	// MemoryDump
	buf      unsafe.Pointer

	// UseDef Analyzer

	// SymbolMap
	SymMap   *SymbolMap

	mapAddrToFunc map[uint32]int
	mapNameToFunc map[string][]int

	EntryAddr uint32
}

func (doc *SoraDocument) LoadYaml(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&doc.yaml)
	if err != nil {
		return err
	}

	return nil
}

func (doc *SoraDocument) LoadMemory(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	doc.buf = bridge.GlobalSetMemoryBase(data, doc.yaml.Memory.Start)

	return nil
}

func NewSoraDocument(path string, load_analyzed bool) (*SoraDocument, error) {
	main_yaml := filepath.Join(path, "Sora.yaml")
	main_data := filepath.Join(path, "SoraMemory.bin")
	bb_data   := filepath.Join(path, "SoraBBTrace.rec")

	doc := &SoraDocument{
		SymMap: CreateSymbolMap(),
		mapAddrToFunc: make(map[uint32]int),
		mapNameToFunc: make(map[string][]int),
	}
	bridge.GlobalSetSymbolMap(doc.SymMap.ptr)
	bridge.GlobalSetGetFuncNameFunc(doc.GetHLEFuncName)
	doc.Parser = NewBBTraceParser(doc, bb_data)
	doc.BBManager = NewBasicBlockManager(doc)
	doc.FunManager = NewFunctionManager(doc)
	doc.InstrManager = NewInstructionManager(doc)

	err := doc.LoadYaml(main_yaml)
	if err != nil {
		return nil, err
	}

	err = doc.LoadMemory(main_data)
	if err != nil {
		return nil, err
	}

	for _, modl := range doc.yaml.LoadedModules {
		doc.SymMap.AddModule(modl.Name, modl.Address, uint32(modl.Size))
	}

	for idx := range doc.yaml.SymFunctions {
		fun := &doc.yaml.SymFunctions[idx]
		doc.FunManager.RegisterExistingFunction(fun)
	}

	doc.EntryAddr = doc.yaml.Module.NM.EntryAddr

	return doc, err
}

func (doc *SoraDocument) GetHLEFuncName(moduleIndex int, funcIndex int) string {
	if moduleIndex < len(doc.yaml.HLEModules) {
		modl := &doc.yaml.HLEModules[moduleIndex]
		if funcIndex < len(modl.Funcs) {
			fun := &modl.Funcs[funcIndex]
			return fmt.Sprintf("%s::%s", modl.Name, fun.Name)
		} else {
			return fmt.Sprintf("%s::func%x", modl.Name, funcIndex)
		}
	}
	return fmt.Sprintf("HLE(%x,%x)", moduleIndex, funcIndex)
}

func (doc *SoraDocument) Delete() {
	bridge.FreeAllocatedCString()
	bridge.GlobalSetGetFuncNameFunc(nil)
	bridge.GlobalSetSymbolMap(nil)
	bridge.GlobalSetMemoryBase(nil, 0)
	doc.SymMap.Delete()
}

func (doc *SoraDocument) Disasm(address uint32) *SoraInstruction {
	if !bridge.MemoryIsValidAddress(address) {
		return nil
	}
	instr := doc.InstrManager.Get(address)
	if instr != nil {
		return instr
	}
	instr = doc.InstrManager.Create(address, bridge.MIPSAnalystGetOpcodeInfo(address))
	mnemonic, args := doc.ParseDizz(instr.Info.Dizz)
	instr.Mnemonic = mnemonic
	instr.Args = args
	return instr
}

type SoraArgType string

const (
	ArgNone SoraArgType = ""
	ArgImm SoraArgType = "imm"
	ArgReg SoraArgType = "reg"
	ArgMem SoraArgType = "mem"
)
type SoraArgument struct {
	Type SoraArgType
	Label string
	ValOfs int
	Reg string
	IsCodeLocation bool
}

func NewSoraArgument(opr string, labellookup func(uint32)*string) (arg *SoraArgument) {
	if len(opr) == 2 {
		return &SoraArgument{
			Type: ArgReg,
			Reg: opr,
		}
	}

	arg = &SoraArgument{}

	lookup := false
	if strings.HasPrefix(opr, "->$") {
		opr = "0x" + opr[3:]
		arg.IsCodeLocation = true
		lookup = true
	} else if strings.HasPrefix(opr, "->") {
		opr = opr[2:]
		arg.IsCodeLocation = true
	}

	var imm int
	var rs string

	n, _ := fmt.Sscanf(opr, "%v(%s)", &imm, &rs)
	if n >= 1 {
		arg.Type = ArgImm
		arg.ValOfs = imm
		if n >= 2 {
			arg.Type = ArgMem
			arg.Reg = rs[:len(rs)-1]
		}
	} else {
		arg.Type = ArgReg
		arg.Reg = opr
	}

	if (lookup) {
		if labellookup != nil {
			label := labellookup(uint32(arg.ValOfs))
			if label != nil {
				arg.Label = *label
			}
		}
	}
	return
}

func (doc *SoraDocument) ParseDizz(dizz string) (mnemonic string, arguments []*SoraArgument) {
	params := strings.Split(dizz, "\t")
	mnemonic = params[0]

	for _, param := range params[1:] {
		argz := strings.Split(param, ",")
		for _, a := range argz {
			var arg *SoraArgument = NewSoraArgument(a, doc.SymMap.GetLabelName)
			arguments = append(arguments, arg)
		}
	}

	return
}

func (doc *SoraDocument) ProcessBB(start_addr uint32, last_addr uint32, cb BBYieldFunc) int {
	var bbas BBAnalState
	bbas.Init()
	var prevInstr *SoraInstruction = nil

	for addr := start_addr; last_addr == 0 || addr <= last_addr; addr += 4 {
		bbas.SetBB(addr)

		instr := doc.Disasm(addr)

		bbas.Append(instr)

		if instr.Info.IsBranch {
			bbas.SetBranch(addr)

			if !instr.Info.HasDelaySlot {
				fmt.Printf("WARNING:\tunhandled branch without delay shot\n")
				bbas.Yield(addr, cb)

				if last_addr == 0 && instr.Info.IsConditional {
					break
				}
			}
		}

		if prevInstr != nil && prevInstr.Info.HasDelaySlot {
			bbas.Yield(addr, cb)

			if last_addr == 0 && !prevInstr.Info.IsConditional {
				break
			}
		}

		prevInstr = instr
	}

	bbas.Yield(last_addr, cb)

	return bbas.Count
}