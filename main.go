package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func startREPL(fc *ForthCompiler) {
	scanner := bufio.NewScanner(os.Stdin)
	fvm := NewForthVM()

	for {
		fmt.Print("forth> ")
		scanner.Scan()
		text := scanner.Text()

		if text == "exit" {
			break
		}

		if text[0] == ':' {
			// just parse
			fc.Parse(text)
			continue
		}

		if text[0] == '%' && len(text) > 1 {
			// show just one definition
			printDefinition(fc, text[2:])
			continue
		} else if text[0] == '%' && len(text) == 1 {
			// show all definitions
			printAllDefinitions(fc)
			continue
		} else if text[0] == '#' && len(text) > 1 {
			// open a file an parse its contents
			fc.ParseFile(text[2:])
			continue
		}

		bc := fc.Compile(": main " + text + " ;")

		fc.output.Reset()

		if fc.fgen {
			fc.printSubs()
		}

		fc.printResult(bc)

		fvm.Run(fc.output.String())
		fmt.Println("")
	}
}

var (
	fgen    bool
	colored bool
	fname   string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "compile file and execute")
	flag.BoolVar(&fgen, "s", true, "Use subs")
	flag.BoolVar(&colored, "c", true, "Use colors")

	flag.Parse()
}

func main() {
	initFlags()

	// -----------------------

	fc := NewForthCompiler()
	fc.fgen = fgen

	// fc.Parse("\\ comment here \n: add2 2 + ;")
	fc.ParseFile("core.fs")

	if len(fname) > 0 {
		fvm := NewForthVM()
		fc.ParseFile(fname)
		bc := fc.CompileMain()
		fc.output.Reset()

		if fc.fgen {
			fc.printSubs()
		}

		fc.printResult(bc)

		fvm.Run(fc.output.String())
		fmt.Println("")
		return
	}

	startREPL(fc)
}
