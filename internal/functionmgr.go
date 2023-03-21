package internal

import (
	"fmt"

	"github.com/firodj/pspsora/binarysearchtree"
)

type FunctionManager struct {
	doc           *SoraDocument
	functions     binarysearchtree.AVLTree[uint32, *SoraFunction]
	mapNameToFunc map[string][]uint32
}

func NewFunctionManager(doc *SoraDocument) *FunctionManager {
	return &FunctionManager{
		doc:           doc,
		mapNameToFunc: make(map[string][]uint32),
	}
}

// RegisterExistingFunction got fun from yaml.SymFunctions and store into analyzed.Functions
func (funmgr *FunctionManager) RegisterExistingFunction(fun *SoraFunction) {
	funmgr.doc.SymMap.AddFunction(fun.Name, fun.Address, fun.Size, -1)
	funmgr.CreateNewFunction(fun.Address, fun.Size)
}

func (funmgr *FunctionManager) RegisterNameFunction(fun *SoraFunction) {
	if _, ok := funmgr.mapNameToFunc[fun.Name]; !ok {
		funmgr.mapNameToFunc[fun.Name] = make([]uint32, 0)
	}

	for _, ex_addr := range funmgr.mapNameToFunc[fun.Name] {
		if ex_addr == fun.Address {
			return
		}
	}

	funmgr.mapNameToFunc[fun.Name] = append(funmgr.mapNameToFunc[fun.Name], fun.Address)
}

func (funmgr *FunctionManager) GetByName(name string) []*SoraFunction {
	if addrs, ok := funmgr.mapNameToFunc[name]; ok {
		funs := make([]*SoraFunction, len(addrs))
		for i, addr := range addrs {
			funs[i] = funmgr.Get(addr)
		}
		return funs
	}
	return nil
}

func (funmgr *FunctionManager) CreateNewFunction(addr uint32, size uint32) *SoraFunction {
	fun := funmgr.Get(addr)
	if fun != nil {
		return nil
	}

	add_sym := false
	name := funmgr.doc.SymMap.GetLabelName(addr)
	if name == nil {
		name = new(string)
		*name = fmt.Sprintf("z_un_%08x", addr)

		add_sym = true
	}

	fun = &SoraFunction{
		Address: addr,
		Name:    *name,
		Size:    size,
	}
	funmgr.functions.Insert(addr, fun)
	funmgr.RegisterNameFunction(fun)

	if add_sym {
		funmgr.doc.SymMap.AddFunction(fun.Name, fun.Address, fun.Size, -1)
	}

	return fun
}

func (funmgr *FunctionManager) Get(addr uint32) *SoraFunction {
	it := funmgr.functions.Search(addr)
	if it.End() {
		return nil
	}
	return it.Value()
}

func (mgr *FunctionManager) SplitAt(split_addr uint32) (prev_func, split_func *SoraFunction) {
	fn_start := mgr.doc.SymMap.GetFunctionStart(split_addr)

	if fn_start == 0 {
		funcStart := mgr.Get(split_addr)
		if funcStart == nil {
			fmt.Printf("TODO:\tunimplemented create func when split at 0x%08x\n", split_addr)
		}
		return
	}
	prev_func = mgr.Get(fn_start)

	last_addr := prev_func.LastAddress()
	if prev_func.LastAddress() >= split_addr {
		prev_func.SetLastAddress(split_addr - 4)
	}

	split_size := last_addr - split_addr + 4
	split_func = mgr.CreateNewFunction(split_addr, split_size)

	if split_func == nil {
		prev_func.SetLastAddress(last_addr)
		fmt.Printf("ERROR:\tunable to create splitted func at 0x%08x\n", split_addr)
		return
	}

	mgr.doc.SymMap.SetFunctionSize(prev_func.Address, prev_func.Size)

	mgr.reBBAddresses(prev_func, split_func)
	return
}

func (mgr *FunctionManager) reBBAddresses(funs ...*SoraFunction) {
	removeds := make([]uint32, 0)
	for _, fun := range funs {
		owneds := make([]uint32, 0)
		for _, addr := range fun.BBAddresses {
			if addr >= fun.Address && addr <= fun.LastAddress() {
				fmt.Printf("debug\tfunc 0x%08x owning bb 0x%08x\n", fun.Address, addr)
				owneds = append(owneds, addr)
			} else {
				fmt.Printf("debug\tfunc 0x%08x reject bb 0x%08x\n", fun.Address, addr)
				removeds = append(removeds, addr)
			}
		}
		fun.BBAddresses = owneds
	}

	orphans := make([]uint32, 0)
	for _, addr := range removeds {
		picked := false
		for _, fun := range funs {
			if addr >= fun.Address && addr <= fun.LastAddress() {
				fmt.Printf("debug\tfunc 0x%08x accept bb 0x%08x\n", fun.Address, addr)
				fun.BBAddresses = append(fun.BBAddresses, addr)
				picked = true
				break
			}
		}
		if !picked {
			orphans = append(orphans, addr)
		}
	}

	for _, addr := range orphans {
		fmt.Printf("debug\torphan bb 0x%08x\n", addr)
	}
}
