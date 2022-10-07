package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"gopkg.in/yaml.v3"

	"github.com/firodj/ppsspp/disasm/pspdisasm/binarysearchtree"
	"github.com/firodj/ppsspp/disasm/pspdisasm/bridge"
	"github.com/firodj/ppsspp/disasm/pspdisasm/models"
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

type SoraBasicBlock struct {
	Adress        uint32 `yaml:"address"`
	LastAddress   uint32 `yaml:"last_address"`
	BranchAddress uint32 `yaml:"branch_address"`
}

type SoraBBRef struct {
	From  uint32 `yaml:"from"`
	To    uint32 `yaml:"to"`
	Flags uint32 `yaml:flags"`

	IsDynamic  bool // immediate or by reg/mem/ptr
	IsAdjacent bool // next/prev
	IsLinked   bool // call/linked
	IsVisited  bool // by bbtrace
}

type SoraYaml struct {
	Module        PSPModule         `yaml:"module"`
	Memory        PSPMemory         `yaml:"memory"`
	LoadedModules []PSPLoadedModule `yaml:"loaded_modules"`
	SymFunctions  []SoraFunction    `yaml:"functions"`
	HLEModules    []PSPHLEModule    `yaml:"hle_modules"`
}

// TODO: will move to other persist storage
type SoraAnalyzed struct {
	BasicBlocks    []SoraBasicBlock `yaml:"basic_blocks"`
	BasicBlockRefs []SoraBBRef      `yaml:"basic_block_refs"`
	Functions      []SoraFunction   `yaml:"functions"`
}

type SoraDocument struct {
	yaml     SoraYaml
	analyzed SoraAnalyzed
	// HLEModules (yaml)
	bbtraceparser *BBTraceParser
	basicBlocks binarysearchtree.ItemBinarySearchTree[*SoraBasicBlock]

	// MemoryDump
	buf      unsafe.Pointer

	// UseDef Analyzer

	// SymbolMap
	symmap   *SymbolMap

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
	//anal_yaml := filepath.Join(path, "SoraAnalyzed.yaml")

	doc := &SoraDocument{
		symmap: CreateSymbolMap(),
		mapAddrToFunc: make(map[uint32]int),
		mapNameToFunc: make(map[string][]int),
	}
	bridge.GlobalSetSymbolMap(doc.symmap.ptr)
	bridge.GlobalSetGetFuncNameFunc(doc.GetHLEFuncName)
	doc.bbtraceparser = NewBBTraceParser(doc, bb_data)

	err := doc.LoadYaml(main_yaml)
	if err != nil {
		return nil, err
	}

	err = doc.LoadMemory(main_data)
	if err != nil {
		return nil, err
	}

	for _, modl := range doc.yaml.LoadedModules {
		doc.symmap.AddModule(modl.Name, modl.Address, uint32(modl.Size))
	}

	for idx := range doc.yaml.SymFunctions {
		fun := &doc.yaml.SymFunctions[idx]
		doc.RegisterExistingFunction(fun)
	}

	doc.EntryAddr = doc.yaml.Module.NM.EntryAddr

	return doc, err
}

func (doc *SoraDocument) GetLabelName(addr uint32) *string {
	return doc.symmap.GetLabelName(addr)
}

func (doc *SoraDocument) RegisterNameFunction(idx int) {
	fun := &doc.yaml.SymFunctions[idx]

	if _, ok := doc.mapNameToFunc[fun.Name]; !ok {
		doc.mapNameToFunc[fun.Name] = make([]int, 0)
	}

	for _, exidx := range doc.mapNameToFunc[fun.Name] {
		if exidx == idx {
			return
		}
	}

	doc.mapNameToFunc[fun.Name] = append(doc.mapNameToFunc[fun.Name], idx)
}

// RegisterExistingFunction got fun from yaml.SymFunctions and store  into analyzed.Functions
func (doc *SoraDocument) RegisterExistingFunction(fun *SoraFunction)  {
	doc.symmap.AddFunction(fun.Name, fun.Address, fun.Size, -1)
	doc.CreateNewFunction(fun.Address, fun.Size)
}

func (doc *SoraDocument) CreateNewFunction(addr uint32, size uint32) int {
	if _, ok := doc.mapAddrToFunc[addr]; ok {
		fmt.Printf("WARNING:\tduplicate address CreateNewFunction addr:0x%08x\n", addr);
		return -1
	}
	name := doc.symmap.GetLabelName(addr)
	if name == nil {
		name = new(string)
		*name = fmt.Sprintf("z_un_%08x", addr)
	}

	idx := len(doc.analyzed.Functions)

	doc.analyzed.Functions = append(doc.analyzed.Functions, SoraFunction{
		Address: addr,
		Name: *name,
		Size: size,
	})

	fun := &doc.analyzed.Functions[idx]

  doc.symmap.AddFunction(fun.Name, fun.Address, fun.Size, -1)

	doc.mapAddrToFunc[fun.Address] = idx

	doc.RegisterNameFunction(idx)

	return idx
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
	doc.symmap.Delete()
}

func (doc *SoraDocument) Disasm(address uint32) *models.MipsOpcode {
	if !bridge.MemoryIsValidAddress(address) {
		fmt.Println("invalid address")
		return nil
	}

	return bridge.MIPSAnalystGetOpcodeInfo(address)
}

func (doc *SoraDocument) GetFunctionByAddress(address uint32) (int, *SoraFunction)  {
	if idx, ok := doc.mapAddrToFunc[address]; ok {
		return idx, &doc.analyzed.Functions[idx]
	}
	return -1, nil
}

func (doc *SoraDocument) GetFunctionByIndex(idx int) *SoraFunction {
	if idx < 0 || idx >= len(doc.analyzed.Functions) {
		return nil
	}
	return &doc.analyzed.Functions[idx]
}

func (doc *SoraDocument) GetBB(bb_addr uint32) *SoraBasicBlock {
	if bb_addr == 0 {
		return nil
	}



	return nil
}

func (doc *SoraDocument) Parser() *BBTraceParser {
	return doc.bbtraceparser
}
