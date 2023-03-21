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
	"strconv"
	"strings"
	"time"

	"github.com/firodj/pspsora/internal"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func testDisasm(doc *internal.SoraDocument) *ffcli.Command {
	fs := flag.NewFlagSet("testDisasm", flag.ExitOnError)
	var addr uint = 0
	fs.UintVar(&addr, "addr", 0, "start address")
	// 0x08a0e874, 0x08a0e890
	return &ffcli.Command{
		Name:    "testDisasm",
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if addr == 0 {
				return errors.New("missing addr")
			}
			if doc.Disasm(uint32(addr)) == nil {
				return errors.New("invalid addr, missing instruction")
			}
			doc.ProcessBB(uint32(addr), 0, doc.GetPrintLines)
			return nil
		},
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
	fs := flag.NewFlagSet("testBBTrace", flag.ExitOnError)
	var length uint = 0
	var funcs string
	fs.UintVar(&length, "length", 0, "length trace, 0 =all")
	fs.StringVar(&funcs, "funcs", "start", "funcs to show, comma sep")

	return &ffcli.Command{
		Name:    "testBBTrace",
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			err := doc.Parser.Parse(ctx, length)
			doc.Parser.DumpAllFunGraph()
			doc.Parser.DumpAllCallHistory()
			if err != nil {
				return err
			}

			funs := make([]*internal.SoraFunction, 0)
			for _, fNorA := range strings.Split(funcs, ",") {
				fNorA = strings.TrimSpace(fNorA)

				if strings.HasPrefix(fNorA, "0x") {
					fmt.Printf("trying %s\n", fNorA)
					if addr, err := strconv.ParseInt(fNorA, 0, 32); err == nil {
						if funStart := doc.FunManager.Get(uint32(addr)); funStart != nil {
							funs = append(funs, funStart)
							continue
						}
					} else {
						fmt.Println(err)
					}
				}
				funs2 := doc.FunManager.GetByName(fNorA)
				funs = append(funs, funs2...)
			}

			// examples: 0x08a38a70
			for _, fun := range funs {
				fmt.Printf("func name=%s addr=0x%x size=%d last=0x%x\n", fun.Name, fun.Address, fun.Size, fun.LastAddress())
				anal := internal.NewFunctionAnalyzer(doc, fun)
				anal.Process()
				anal.Debug(doc.GetPrintCodes)
			}
			return nil
		},
	}
}

func testLongRunningProcess() *ffcli.Command {
	return &ffcli.Command{
		Name: "testLongRunningProcess",
		Exec: func(ctx context.Context, args []string) error {
			c := doRunningProcess(ctx)

			consumer := func() error {
				for n := range c {
					fmt.Printf("%d ", n)
					time.Sleep(500 * time.Millisecond)
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
			fmt.Printf("func name=%s addr=0x%x size=%d last=0x%x\n", funStart.Name, funStart.Address, funStart.Size, funStart.LastAddress())
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
			testDisasm(doc),
			testLongRunningProcess(),
			testBBTrace(doc),
			testFunAnalyzer(doc),
			serveCommand(doc),
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
