package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/firodj/pspsora/internal"
)

func main() {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	fmt.Println(home)

	doc, err := internal.NewSoraDocument(home + "/Sora", true)
	if err != nil {
		fmt.Println(err)
	}

	entryName := doc.SymMap.GetLabelName(doc.EntryAddr)
	fmt.Println(doc.EntryAddr)
	if entryName != nil {
		fmt.Println(*entryName)
	}

	doc.Disasm(doc.EntryAddr)
	funStart := doc.FunManager.Get(doc.EntryAddr)
	fmt.Println(funStart.Name, funStart.Address, funStart.Size, funStart.LastAddress())
	anal := internal.NewFunctionAnalyzer(doc, funStart)
	anal.Process()

	err = doc.Parser.Parse(
		func (param internal.BBTraceParam) {

		},
		1000,
	)
	if err != nil {
		fmt.Println(err)
	}

	defer doc.Delete()
}
