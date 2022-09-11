package main

import (
	"flag"

	"github.com/loscoala/goforth"
)

var (
	fname  string
	script string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "compile file and execute")
	flag.BoolVar(&goforth.Colored, "color", true, "Use colors")
	flag.StringVar(&script, "script", "", "program passed in as string")

	flag.Parse()
}

func main() {
	initFlags()

	fc := goforth.NewForthCompiler()

	// custom sys func
	//fc.Fvm.Sysfunc = func(fvm goforth.VM, syscall int64) {
	//	switch syscall {
	//	case 999:
	//		fmt.Println("This is a custom call")
	//	default:
	//		fmt.Println("Not implemented")
	//	}
	//}

	// load the core words
	if err := fc.Parse(goforth.Core); err != nil {
		goforth.PrintError(err)
	}

	if len(script) > 0 {
		if err := fc.Run(script); err != nil {
			goforth.PrintError(err)
		}
	} else if len(fname) > 0 {
		if err := fc.RunFile(fname); err != nil {
			goforth.PrintError(err)
		}
	} else {
		fc.StartREPL()
	}
}