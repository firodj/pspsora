package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/firodj/pspsora/internal"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func testSysCall(doc *internal.SoraDocument) *ffcli.Command {
	return &ffcli.Command{
		Name: "testSysCall",
		Exec: func(ctx context.Context, args []string) error {
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
			return nil
		},
	}
}


func testDisasm(doc *internal.SoraDocument) *ffcli.Command {
	return &ffcli.Command{
		Name: "testDisasm",
		Exec: func(ctx context.Context, args []string) error {
			doc.ProcessBB(0x08a0e874, 0, GetPrintLines(doc))
			doc.ProcessBB(0x08a0e890, 0, GetPrintLines(doc))
			return nil
		},
	}

}

func GetPrintLines(doc *internal.SoraDocument) internal.BBYieldFunc {
	return func(state internal.BBAnalState) {
		funStart := doc.SymMap.GetFunctionStart(state.BBAddr)
		var label *string
		if funStart != 0 {
			label = doc.SymMap.GetLabelName(funStart)
			fmt.Println(*label)
		}

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
			fmt.Printf("0x%08x\t%s\n", line.Address, line.Info.Dizz)
		}
		//fmt.Printf("last 0x%08x, branch 0x%08x\n", state.LastAddr, state.BranchAddr)
		fmt.Printf("---\n")
	}
}

func doRunningProcess(ctx context.Context) chan int {
	c := make(chan int)

	producer := func() {
		n := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Println("canceled")
				close(c)
				return
			case c <- n:
				n += 1
				if n > 20 {
					close(c)
					return
				}
			}
		}
	}

	go producer()

	return c
}

func testBBTrace(doc *internal.SoraDocument) *ffcli.Command {
	return &ffcli.Command {
		Name: "testBBTrace",
		Exec: func(ctx context.Context, args []string) error {

			err := doc.Parser.Parse(ctx, 0)
			doc.Parser.DumpAllFunGraph()
			doc.Parser.DumpAllCallHistory()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func testLongRunningProcess() *ffcli.Command {
	return &ffcli.Command {
		Name: "testLongRunningProcess",
		Exec: func(ctx context.Context, args []string) error {
			c := doRunningProcess(ctx)

			consumer := func () error {
				for n := range c {
					fmt.Printf("%d ", n)
					time.Sleep(500 * time.Millisecond )
				}
				return ctx.Err()
			}

			return consumer()
		},
	}
}

func testFunAnalyzer(doc *internal.SoraDocument) *ffcli.Command {
	return &ffcli.Command{
		Name: "testFunAnalyzer",
		Exec: func(ctx context.Context, args []string) error {
			funStart := doc.FunManager.Get(doc.EntryAddr)
			fmt.Println(funStart.Name, funStart.Address, funStart.Size, funStart.LastAddress())
			anal := internal.NewFunctionAnalyzer(doc, funStart)
			anal.Process()
			return nil
		},
	}
}
func main() {
	appName := filepath.Base(os.Args[0])

	rootFlagSet := flag.NewFlagSet(appName, flag.ExitOnError)

	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	fmt.Println("HOME ", home)

	doc, err := internal.NewSoraDocument(home+"/Sora", true)
	if err != nil {
		fmt.Println(err)
	}
	defer doc.Delete()

	ctx := context.Background()
	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	defer func() {
		signal.Stop(quit)
		cancel()
	}()

	go func() {
		<-quit
		cancel()
	}()

	root := &ffcli.Command{
		ShortUsage: appName + " [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Subcommands: []*ffcli.Command{
			testSysCall(doc),
			testDisasm(doc),
			testLongRunningProcess(),
			testBBTrace(doc),
			testFunAnalyzer(doc),
		},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	err = root.ParseAndRun(ctx, os.Args[1:])
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			panic(err)
		}
	}
}
