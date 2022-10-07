package internal

import "github.com/firodj/ppsspp/disasm/pspdisasm/bridge"

type SymbolMap struct {
	ptr bridge.CSymbolMap
}

func CreateSymbolMap() *SymbolMap {
	symmap := &SymbolMap{
		ptr: bridge.NewSymbolMap(),
	}
	return symmap
}

func (symmap *SymbolMap) Delete() {
	bridge.DeleteSymbolMap(symmap.ptr)
}

func (symmap *SymbolMap) GetFunctionSize(startAddress uint32) uint32 {
	return bridge.SymbolMap_GetFunctionSize(symmap.ptr, startAddress)
}

func (symmap *SymbolMap) GetFunctionStart(address uint32) uint32 {
	return bridge.SymbolMap_GetFunctionStart(symmap.ptr, address)
}

func (symmap *SymbolMap) GetLabelName(address uint32) *string {
	return bridge.SymbolMap_GetLabelName(symmap.ptr, address)
}

func (symmap *SymbolMap) AddFunction(name string, address uint32, size uint32, moduleIndex int) {
	bridge.SymbolMap_AddFunction(symmap.ptr, name, address, size, moduleIndex)
}

func (symmap *SymbolMap) AddModule(name string, address uint32, size uint32) {
	bridge.SymbolMap_AddModule(symmap.ptr, name, address, size)
}