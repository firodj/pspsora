package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/davecgh/go-spew/spew"
	"github.com/firodj/pspsora/internal"
)

func testSysCall(doc *internal.SoraDocument) {
	entryName := doc.SymMap.GetLabelName(doc.EntryAddr)
	fmt.Println(doc.EntryAddr)
	if entryName != nil {
		fmt.Println(*entryName)
	}

	instr := doc.Disasm(doc.EntryAddr)
	fmt.Println(instr.Info.Dizz)
	spew.Dump(doc.ParseDizz(instr.Info.Dizz))

	instr = doc.Disasm(0x8A38A70)
	fmt.Println(instr.Info.Dizz)
	spew.Dump(doc.ParseDizz(instr.Info.Dizz))

	instr = doc.Disasm(0x8A38A74)
	fmt.Println(instr.Info.Dizz)
	spew.Dump(doc.ParseDizz(instr.Info.Dizz))

	instr = doc.Disasm(0x8804140)
	fmt.Println(instr.Info.Dizz)
	spew.Dump(doc.ParseDizz(instr.Info.Dizz))
}

func testBBTrace(doc *internal.SoraDocument) {
	err := doc.Parser.Parse(0)
	doc.Parser.DumpAllFunGraph()
	doc.Parser.DumpAllCallHistory()
	if err != nil {
		panic(err)
	}
}

func main() {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	fmt.Println(home)

	doc, err := internal.NewSoraDocument(home+"/Sora", true)
	if err != nil {
		fmt.Println(err)
	}

	testSysCall(doc)

	//funStart := doc.FunManager.Get(doc.EntryAddr)
	//fmt.Println(funStart.Name, funStart.Address, funStart.Size, funStart.LastAddress())
	//anal := internal.NewFunctionAnalyzer(doc, funStart)
	//anal.Process()

	testBBTrace(doc)

	defer doc.Delete()
}
