package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"gopkg.in/yaml.v3"

	"github.com/davecgh/go-spew/spew"
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
	BBAddresses []uint32 `yaml:"bb_addresses"`
}

func (fun *SoraFunction) LastAddress() uint32 {
	return fun.Address + fun.Size - 4
}

func (fun *SoraFunction) SetLastAddress(last_addr uint32) {
	fun.Size = last_addr - fun.Address + 4
}

func (fun *SoraFunction) AddBB(bb_addr uint32) {
	for _, ex_bb := range fun.BBAddresses {
		if ex_bb == bb_addr {
			return
		}
	}

	fun.BBAddresses = append(fun.BBAddresses, bb_addr)
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
	yaml SoraYaml

	// HLEModules (yaml)
	Parser       *BBTraceParser
	BBManager    *BasicBlockManager
	FunManager   *FunctionManager
	InstrManager *InstructionManager

	// MemoryDump
	buf unsafe.Pointer

	// UseDef Analyzer

	// SymbolMap
	SymMap *SymbolMap

	EntryAddr uint32

	// flags
	debugMode int
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
	fmt.Println("main Yaml:", filename)
	return nil
}

func (doc *SoraDocument) LoadMemory(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fmt.Println("main Data:", filename)
	doc.buf = bridge.GlobalSetMemoryBase(data, doc.yaml.Memory.Start)

	return nil
}

func newSoraDocument() *SoraDocument {
	doc := &SoraDocument{
		SymMap:    CreateSymbolMap(),
		debugMode: 0,
	}
	bridge.GlobalSetSymbolMap(doc.SymMap.ptr)
	bridge.GlobalSetGetFuncNameFunc(doc.GetHLEFuncName)

	doc.Parser = NewBBTraceParser(doc)
	doc.BBManager = NewBasicBlockManager(doc)
	doc.FunManager = NewFunctionManager(doc)
	doc.InstrManager = NewInstructionManager(doc)

	return doc
}

func NewSoraDocument(path string, load_analyzed bool) (*SoraDocument, error) {
	main_yaml := filepath.Join(path, "Sora.yaml")
	main_data := filepath.Join(path, "SoraMemory.bin")
	bb_data := filepath.Join(path, "SoraBBTrace.rec")

	doc := newSoraDocument()
	doc.Parser.setFilename(bb_data)

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

func (doc *SoraDocument) GetHLE(moduleIndex int, funcIndex int) (*PSPHLEModule, *PSPHLEFunction) {
	if moduleIndex < len(doc.yaml.HLEModules) {
		modl := &doc.yaml.HLEModules[moduleIndex]
		if funcIndex < len(modl.Funcs) {
			fun := &modl.Funcs[funcIndex]
			return modl, fun
		}
		return modl, nil
	}
	return nil, nil
}

func (doc *SoraDocument) GetHLEFuncName(moduleIndex int, funcIndex int) string {
	modl, fun := doc.GetHLE(moduleIndex, funcIndex)

	if modl != nil {
		if fun != nil {
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

func (doc *SoraDocument) ParseDizz(dizz string) (mnemonic string, arguments []*SoraArgument) {
	params := strings.Split(dizz, "\t")
	mnemonic = params[0]

	for _, param := range params[1:] {
		argz := strings.Split(param, ",")
		for _, a := range argz {
			var arg *SoraArgument = NewSoraArgument(strings.TrimSpace(a), doc.SymMap.GetLabelName)
			arguments = append(arguments, arg)
		}
	}

	return
}

func (doc *SoraDocument) ProcessBB(start_addr uint32, last_addr uint32, cb BBYieldFunc) int {
	var bbas BBAnalState
	bbas.Init()
	var prevInstr *SoraInstruction = nil
	bb_exists := doc.BBManager.Get(start_addr)
	if bb_exists != nil {
		//fmt.Printf("WARNING:\toverwrite last_addr because ProcessBB exists on 0x%08x\n", bb_exists.Address)
		last_addr = bb_exists.LastAddress
	}

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

func (doc *SoraDocument) DebugBB(theBB *SoraBasicBlock, mode string) {
	if doc.Parser.CurrentID != 1 {
		return
	}

	fmt.Printf("DEBUG: [%s]\n", mode)
	for addr := theBB.Address; addr <= theBB.LastAddress; addr += 4 {
		instr := doc.InstrManager.Get(addr)
		fmt.Print("\t")
		if instr.Address == theBB.BranchAddress {
			fmt.Print("*")
		} else {
			fmt.Print(" ")
		}
		if instr.Info.HasDelaySlot {
			fmt.Print("D")
		} else {
			fmt.Print(" ")
		}

		fmt.Printf(" 0x%08x: %s\n", instr.Address, instr.Info.Dizz)
	}
}

func (doc *SoraDocument) GetRegName(cat int, index int) string {
	regname := bridge.MIPSDebugInterface_GetRegName(cat, index)
	if regname != nil {
		return *regname
	}
	return ""
}

func (doc *SoraDocument) GetLabelName(addr uint32) string {
	if funTarget := doc.SymMap.GetFunctionStart(uint32(addr)); funTarget != 0 {
		label := doc.SymMap.GetLabelName(funTarget)
		if label != nil {
			ss := *label
			disp := addr - funTarget
			if disp != 0 {
				ss += fmt.Sprintf("__0x%x", disp)
			}

			return ss
		}
	}
	return fmt.Sprintf("loc_0x%08x", addr)
}

func (doc *SoraDocument) GetPrintLines(state BBAnalState) {
	label := doc.GetLabelName(state.BBAddr)

	fmt.Printf("%s:\t// 0x%08x", label, state.BBAddr)

	if state.Visited {
		fmt.Println("\t(v)")
	} else {
		fmt.Println()
	}

	skip_delayslot := uint32(0)

	for _, line := range state.Lines {
		if line.Address == state.BranchAddr {
			fmt.Print("*")
		} else {
			fmt.Print(" ")
		}
		if line.Address == state.LastAddr {
			fmt.Print("_")
		} else {
			fmt.Print(" ")
		}

		fmt.Printf("0x%08x\t%s", line.Address, line.Info.Dizz)

		if skip_delayslot == line.Address {
			skip_delayslot = 0
		} else {
			ss, ok := Code(line, doc)
			if ok == 1 {
				skip_delayslot = line.Address + 4
			}
			if ss != "" {
				fmt.Println("\t; " + ss)
			}
		}

		fmt.Println()
	}
	//fmt.Printf("last 0x%08x, branch 0x%08x\n", state.LastAddr, state.BranchAddr)
	//fmt.Printf("---\n")
}

func (doc *SoraDocument) GetPrintCodes(state BBAnalState) {
	label := doc.GetLabelName(state.BBAddr)

	xref_froms := doc.BBManager.GetEnterRefs(state.BBAddr)
	for _, xref := range xref_froms {
		spew.Dump(xref)
	}

	fmt.Printf("%s:\t// 0x%08x", label, state.BBAddr)

	if state.Visited {
		fmt.Println("\t// visited")
	} else {
		fmt.Println()
	}

	skip_delayslot := uint32(0)

	for _, line := range state.Lines {
		//fmt.Printf("0x%08x\t%s", line.Address, line.Info.Dizz)

		if skip_delayslot != 0 {
			if line.Address != skip_delayslot {
				panic("unsync skip_delayslot")
			}
			skip_delayslot = 0
		} else {
			ss, ok := Code(line, doc)
			if ok == 1 {
				skip_delayslot = line.Address + 4
			}
			if ss != "" {
				fmt.Printf("%s", ss)
			}
			if ok == -1 {
				fmt.Printf("\t0x%08x\t%s", line.Address, line.Info.Dizz)
				panic("error")
			}
		}

		fmt.Println()
	}
}
