package internal

import (
	"fmt"

	"github.com/firodj/pspsora/binarysearchtree"
)

type FunctionManager struct {
	doc *SoraDocument
	functions binarysearchtree.AVLTree[uint32, *SoraFunction]
	mapNameToFunc map[string][]uint32
}

func NewFunctionManager(doc *SoraDocument) *FunctionManager {
	return &FunctionManager{
		doc: doc,
		mapNameToFunc: make(map[string][]uint32),
	}
}

// RegisterExistingFunction got fun from yaml.SymFunctions and store  into analyzed.Functions
func (funmgr *FunctionManager) RegisterExistingFunction(fun *SoraFunction)  {
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
		Name: *name,
		Size: size,
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
